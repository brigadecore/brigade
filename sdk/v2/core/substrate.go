package core

import (
	"context"
	"net/http"

	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// SubstrateWorkerCount represents a count of Workers currently executing on
// the substrate.
type SubstrateWorkerCount struct {
	// Count is the cardinality of Workers currently executing on the substrate.
	Count int `json:"count"`
}

// SubstrateJobCount represents a count of Workers currently executing on
// the substrate.
type SubstrateJobCount struct {
	// Count is the cardinality of Jobs currently executing on the substrate.
	Count int `json:"count"`
}

// SubstrateClient is the specialized client for monitoring the substrate.
type SubstrateClient interface {
	// CountRunningWorkers returns a count of Workers currently executing on the
	// substrate.
	CountRunningWorkers(context.Context) (SubstrateWorkerCount, error)
	// CountRunningJobs returns a count of Jobs currently executing on the
	// substrate.
	CountRunningJobs(context.Context) (SubstrateJobCount, error)
}

type substrateClient struct {
	*rm.BaseClient
}

// NewSubstrateClient returns a specialized client for monitoring the substrate.
func NewSubstrateClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) SubstrateClient {
	return &substrateClient{
		BaseClient: rm.NewBaseClient(apiAddress, apiToken, opts),
	}
}

func (s *substrateClient) CountRunningWorkers(
	ctx context.Context,
) (SubstrateWorkerCount, error) {
	count := SubstrateWorkerCount{}
	return count, s.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/substrate/running-workers",
			SuccessCode: http.StatusOK,
			RespObj:     &count,
		},
	)
}

func (s *substrateClient) CountRunningJobs(
	ctx context.Context,
) (SubstrateJobCount, error) {
	count := SubstrateJobCount{}
	return count, s.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/substrate/running-jobs",
			SuccessCode: http.StatusOK,
			RespObj:     &count,
		},
	)
}
