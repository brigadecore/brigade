package core

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v2/core"
)

type MockSubstrateClient struct {
	CountRunningWorkersFn func(context.Context) (core.SubstrateWorkerCount, error)
	CountRunningJobsFn    func(context.Context) (core.SubstrateJobCount, error)
}

func (m *MockSubstrateClient) CountRunningWorkers(
	ctx context.Context,
) (core.SubstrateWorkerCount, error) {
	return m.CountRunningWorkersFn(ctx)
}

func (m *MockSubstrateClient) CountRunningJobs(
	ctx context.Context,
) (core.SubstrateJobCount, error) {
	return m.CountRunningJobsFn(ctx)
}
