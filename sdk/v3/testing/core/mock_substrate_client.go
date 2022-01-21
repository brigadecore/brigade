package core

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3/core"
)

type MockSubstrateClient struct {
	CountRunningWorkersFn func(
		context.Context,
		*core.RunningWorkerCountOptions,
	) (core.SubstrateWorkerCount, error)
	CountRunningJobsFn func(
		context.Context,
		*core.RunningJobCountOptions,
	) (core.SubstrateJobCount, error)
}

func (m *MockSubstrateClient) CountRunningWorkers(
	ctx context.Context,
	opts *core.RunningWorkerCountOptions,
) (core.SubstrateWorkerCount, error) {
	return m.CountRunningWorkersFn(ctx, opts)
}

func (m *MockSubstrateClient) CountRunningJobs(
	ctx context.Context,
	opts *core.RunningJobCountOptions,
) (core.SubstrateJobCount, error) {
	return m.CountRunningJobsFn(ctx, opts)
}
