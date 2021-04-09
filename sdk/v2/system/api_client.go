package system

import (
	"context"
	"net/http"

	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// PingResponse represents the expected response object returned by the
// API Server's v2/ping endpoint
type PingResponse struct {
	Version string
}

// APIClient is the client for system checks involving the Brigade API.
type APIClient interface {
	// Ping sends a GET request to the API Server's versioned ping endpoint
	Ping(ctx context.Context) (PingResponse, error)

	// PingRaw sends a GET request to the API Server's unversioned ping endpoint
	PingRaw(ctx context.Context) (string, error)
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

func (a *apiClient) Ping(ctx context.Context) (PingResponse, error) {
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

func (a *apiClient) PingRaw(ctx context.Context) (string, error) {
	var pingResponse string
	return pingResponse, a.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "ping",
			SuccessCode: http.StatusOK,
			RespObj:     &pingResponse,
		},
	)
}
