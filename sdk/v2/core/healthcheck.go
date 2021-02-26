package core

import (
	"context"
	"net/http"

	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// HealthcheckClient is the specialized client for accessing healthchecks on
// the Brigade API.
type HealthcheckClient interface {
	// Ping pings the API Server
	Ping(ctx context.Context) error
}

type healthcheckClient struct {
	*rm.BaseClient
}

// NewHealthcheckClient returns a specialized client for accessing the
// healthcheck API.
func NewHealthcheckClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) HealthcheckClient {
	return &healthcheckClient{
		BaseClient: rm.NewBaseClient(apiAddress, apiToken, opts),
	}
}

func (h *healthcheckClient) Ping(ctx context.Context) error {
	return h.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "ping",
			SuccessCode: http.StatusOK,
		},
	)
}
