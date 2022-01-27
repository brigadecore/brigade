package testing

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3"
)

type MockSubstrateClient struct {
	CountRunningWorkersFn func(
		context.Context,
		*sdk.RunningWorkerCountOptions,
	) (sdk.SubstrateWorkerCount, error)
	CountRunningJobsFn func(
		context.Context,
		*sdk.RunningJobCountOptions,
	) (sdk.SubstrateJobCount, error)
}

func (m *MockSubstrateClient) CountRunningWorkers(
	ctx context.Context,
	opts *sdk.RunningWorkerCountOptions,
) (sdk.SubstrateWorkerCount, error) {
	return m.CountRunningWorkersFn(ctx, opts)
}

func (m *MockSubstrateClient) CountRunningJobs(
	ctx context.Context,
	opts *sdk.RunningJobCountOptions,
) (sdk.SubstrateJobCount, error) {
	return m.CountRunningJobsFn(ctx, opts)
}
