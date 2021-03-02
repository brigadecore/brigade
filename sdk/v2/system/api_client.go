package system

import (
	"context"
	"net/http"

	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// APIClient is the client for system checks involving the Brigade API.
type APIClient interface {
	// Ping pings the API Server
	Ping(ctx context.Context) error
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

func (a *apiClient) Ping(ctx context.Context) error {
	return a.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "ping",
			SuccessCode: http.StatusOK,
		},
	)
}
