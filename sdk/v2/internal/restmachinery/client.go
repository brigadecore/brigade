package restmachinery

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/pkg/errors"
)

// APIClientOptions encapsulates optional API client configuration.
type APIClientOptions struct {
	// AllowInsecureConnections indicates whether SSL-related errors should be
	// ignored when connecting to the API server.
	AllowInsecureConnections bool
}

// BaseClient provides "API machinery" used by all the specialized API clients.
// Its various functions remove the tedium from common API-related operations
// like managing authentication headers, encoding request bodies, interpretting
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
	opts *APIClientOptions,
) *BaseClient {
	if opts == nil {
		opts = &APIClientOptions{}
	}
	return &BaseClient{
		APIAddress: apiAddress,
		APIToken:   apiToken,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: opts.AllowInsecureConnections,
				},
			},
		},
	}
}

// BasicAuthHeaders returns a map[string]string populated with a Basic Auth
// header based on the provided username and password.
func (b *BaseClient) BasicAuthHeaders(
	username string,
	password string,
) map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf(
			"Basic %s",
			base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf("%s:%s", username, password)),
			),
		),
	}
}

// BearerTokenAuthHeaders returns a map[string]string populated with an
// authentication header that makes use of the client's bearer token.
func (b *BaseClient) BearerTokenAuthHeaders() map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", b.APIToken),
	}
}

// AppendListQueryParams returns the provided map[string]string with key/value
// pairs related to pagination of large lists appended. If a nil map is
// provided, a new one is instantiated.
func (b *BaseClient) AppendListQueryParams(
	queryParams map[string]string,
	opts meta.ListOptions,
) map[string]string {
	if queryParams == nil {
		queryParams = map[string]string{}
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
	for k, v := range req.AuthHeaders {
		r.Header.Add(k, v)
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
