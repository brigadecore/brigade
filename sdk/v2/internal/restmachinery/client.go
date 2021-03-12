package restmachinery

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
)

// BaseClient provides "API machinery" used by all the specialized API clients.
// Its various functions remove the tedium from common API-related operations
// like managing authentication headers, encoding request bodies, interpreting
// response codes, decoding responses bodies, and more.
type BaseClient struct {
	APIAddress string
	APIToken   string
	HTTPClient *http.Client
}

// NewBaseClient returns a new BaseClient.
func NewBaseClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) *BaseClient {
	if opts == nil {
		opts = &restmachinery.APIClientOptions{}
	}
	retryClient := retryablehttp.NewClient()
	retryClient.CheckRetry = defaultRetryPolicy
	retryClient.HTTPClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: opts.AllowInsecureConnections,
		},
	}
	retryClient.Logger = log.New(ioutil.Discard, "", log.LstdFlags)
	return &BaseClient{
		APIAddress: apiAddress,
		APIToken:   apiToken,
		HTTPClient: retryClient.StandardClient(),
	}
}

// AppendListQueryParams returns the provided map[string]string with key/value
// pairs related to pagination of large lists appended. If a nil map is
// provided, a new one is instantiated.
func (b *BaseClient) AppendListQueryParams(
	queryParams map[string]string,
	opts *meta.ListOptions,
) map[string]string {
	if queryParams == nil {
		queryParams = map[string]string{}
	}
	if opts == nil {
		return queryParams
	}
	if opts.Continue != "" {
		queryParams["continue"] = opts.Continue
	}
	if opts.Limit != 0 {
		queryParams["limit"] = strconv.FormatInt(opts.Limit, 10)
	}
	return queryParams
}

// ExecuteRequest accepts one argument-- an outboundRequest-- that models all
// aspects of a single API call in a succinct fashion. Based on this
// information, this function prepares and executes an HTTP request, interprets
// the HTTP response code and decodes the response body into a user-supplied
// type.
func (b *BaseClient) ExecuteRequest(
	ctx context.Context,
	req OutboundRequest,
) error {
	resp, err := b.SubmitRequest(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if req.RespObj != nil {
		respBodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "error reading response body")
		}
		if err := json.Unmarshal(respBodyBytes, req.RespObj); err != nil {
			return errors.Wrap(err, "error unmarshaling response body")
		}
	}
	return nil
}

// SubmitRequest accepts one argument-- an outboundRequest-- that models all
// aspects of a single API call in a succinct fashion. Based on this
// information, this function prepares and executes an HTTP request and returns
// the HTTP response. This is a lower-level function than executeRequest().
// It is used by executeRequest(), but is also suitable for uses in cases where
// specialized response handling is required.
// nolint: gocyclo
func (b *BaseClient) SubmitRequest(
	ctx context.Context,
	req OutboundRequest,
) (*http.Response, error) {
	var reqBodyReader io.Reader
	if req.ReqBodyObj != nil {
		switch rb := req.ReqBodyObj.(type) {
		case []byte:
			reqBodyReader = bytes.NewBuffer(rb)
		default:
			reqBodyBytes, err := json.Marshal(req.ReqBodyObj)
			if err != nil {
				return nil, errors.Wrap(err, "error marshaling request body")
			}
			reqBodyReader = bytes.NewBuffer(reqBodyBytes)
		}
	}

	r, err := http.NewRequest(
		req.Method,
		fmt.Sprintf("%s/%s", b.APIAddress, req.Path),
		reqBodyReader,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error creating request %s %s",
			req.Method,
			req.Path,
		)
	}
	r = r.WithContext(ctx)
	if len(req.QueryParams) > 0 {
		q := r.URL.Query()
		for k, v := range req.QueryParams {
			q.Set(k, v)
		}
		r.URL.RawQuery = q.Encode()
	}
	r.Header.Add("Accept", "application/json")
	if req.IncludeAuthHeader == nil || *req.IncludeAuthHeader {
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", b.APIToken))
	}
	for k, v := range req.Headers {
		r.Header.Add(k, v)
	}

	resp, err := b.HTTPClient.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking API")
	}

	if (req.SuccessCode == 0 && resp.StatusCode != http.StatusOK) ||
		(req.SuccessCode != 0 && resp.StatusCode != req.SuccessCode) {
		// HTTP Response code hints at what sort of error might be in the body
		// of the response
		var apiErr error
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			apiErr = &meta.ErrAuthentication{}
		case http.StatusForbidden:
			apiErr = &meta.ErrAuthorization{}
		case http.StatusBadRequest:
			apiErr = &meta.ErrBadRequest{}
		case http.StatusNotFound:
			apiErr = &meta.ErrNotFound{}
		case http.StatusConflict:
			apiErr = &meta.ErrConflict{}
		case http.StatusNotImplemented:
			apiErr = &meta.ErrNotSupported{}
		case http.StatusInternalServerError:
			apiErr = &meta.ErrInternalServer{}
		default:
			return nil, errors.Errorf("received %d from API server", resp.StatusCode)
		}
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "error reading error response body")
		}
		if err = json.Unmarshal(bodyBytes, apiErr); err != nil {
			return nil, errors.Wrap(err, "error unmarshaling error response body")
		}
		return nil, apiErr
	}
	return resp, nil
}

// defaultRetryPolicy was pulled from github.com/hashicorp/go-retryablehttp
// in order to support modifying baseRetryPolicy (defined below) to our needs.
// It represents the default callback for our baseClient's retryable http
// client, which will retry on connection errors and server errors.
func defaultRetryPolicy(
	ctx context.Context,
	resp *http.Response,
	err error,
) (bool, error) {
	// do not retry on context.Canceled or context.DeadlineExceeded
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	// don't propagate other errors
	shouldRetry, _ := baseRetryPolicy(resp, err)
	return shouldRetry, nil
}

// baseRetryPolicy is a slightly modified version of the same function in
// github.com/hashicorp/go-retryablehttp.  We wish to exclude retries on 500
// errors (and the original library already excludes 501), so we've updated
// the pertinent logic here.
func baseRetryPolicy(resp *http.Response, err error) (bool, error) {
	var redirectsErrorRe = regexp.MustCompile(`stopped after \d+ redirects\z`)
	var schemeErrorRe = regexp.MustCompile(`unsupported protocol scheme`)

	if err != nil {
		if v, ok := err.(*url.Error); ok {
			// Don't retry if the error was due to too many redirects.
			if redirectsErrorRe.MatchString(v.Error()) {
				return false, v
			}

			// Don't retry if the error was due to an invalid protocol scheme.
			if schemeErrorRe.MatchString(v.Error()) {
				return false, v
			}

			// Don't retry if the error was due to TLS cert verification failure.
			if _, ok := v.Err.(x509.UnknownAuthorityError); ok {
				return false, v
			}
		}

		// The error is likely recoverable so retry.
		return true, nil
	}

	// 429 Too Many Requests is recoverable. Sometimes the server puts
	// a Retry-After response header to indicate when the server is
	// available to start processing request from client.
	if resp.StatusCode == http.StatusTooManyRequests {
		return true, nil
	}

	// Check the response code. We retry on 5xx-range responses greater than
	// 501 to allow the server time to recover, as 5xx's are typically not
	// permanent errors and may relate to outages on the server side. This will
	// catch invalid response codes as well, like 0 and 999.
	if resp.StatusCode == 0 || resp.StatusCode > 501 {
		return true, fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}

	return false, nil
}
