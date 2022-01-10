package system

import (
	"context"
	"io/ioutil"
	"net/http"

	rm "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v3/restmachinery"
	"github.com/pkg/errors"
)

// PingResponse represents the expected response object returned by the
// API Server's v2/ping endpoint
type PingResponse struct {
	Version string
	Commit  string
}

// PingOptions represents useful, optional criteria for pinging a Brigade API
// server. It currently has no fields, but exists to preserve the possibility of
// future expansion without having to change client function signatures.
type PingOptions struct{}

// APIClient is the client for system checks involving the Brigade API.
type APIClient interface {
	// Ping sends a GET request to the API Server's versioned ping endpoint
	Ping(context.Context, *PingOptions) (PingResponse, error)

	// UnversionedPing sends a GET request to the API Server's unversioned
	// ping endpoint
	UnversionedPing(context.Context) ([]byte, error)
}

type apiClient struct {
	*rm.BaseClient
}

// NewAPIClient returns a client to access system-related Brigade API
// endpoints.
func NewAPIClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) APIClient {
	return &apiClient{
		BaseClient: rm.NewBaseClient(apiAddress, apiToken, opts),
	}
}

func (a *apiClient) Ping(
	ctx context.Context,
	_ *PingOptions,
) (PingResponse, error) {
	pingResponse := PingResponse{}
	return pingResponse, a.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/ping",
			SuccessCode: http.StatusOK,
			RespObj:     &pingResponse,
		},
	)
}

func (a *apiClient) UnversionedPing(ctx context.Context) ([]byte, error) {
	resp, err := a.SubmitRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "ping",
			SuccessCode: http.StatusOK,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "error submitting request")
	}
	defer resp.Body.Close()
	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}
	return respBodyBytes, nil
}
