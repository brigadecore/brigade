package core

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3/core"
)

type MockWorkersClient struct {
	StartFn func(
		ctx context.Context,
		eventID string,
		opts *core.WorkerStartOptions,
	) error
	GetStatusFn func(
		ctx context.Context,
		eventID string,
		opts *core.WorkerStatusGetOptions,
	) (core.WorkerStatus, error)
	WatchStatusFn func(
		ctx context.Context,
		eventID string,
		opts *core.WorkerStatusWatchOptions,
	) (<-chan core.WorkerStatus, <-chan error, error)
	UpdateStatusFn func(
		ctx context.Context,
		eventID string,
		status core.WorkerStatus,
		opts *core.WorkerStatusUpdateOptions,
	) error
	CleanupFn func(
		ctx context.Context,
		eventID string,
		opts *core.WorkerCleanupOptions,
	) error
	TimeoutFn func(
		ctx context.Context,
		eventID string,
		opts *core.WorkerTimeoutOptions,
	) error
	JobsClient core.JobsClient
}

func (m *MockWorkersClient) Start(
	ctx context.Context,
	eventID string,
	opts *core.WorkerStartOptions,
) error {
	return m.StartFn(ctx, eventID, opts)
}

func (m *MockWorkersClient) GetStatus(
	ctx context.Context,
	eventID string,
	opts *core.WorkerStatusGetOptions,
) (core.WorkerStatus, error) {
	return m.GetStatusFn(ctx, eventID, opts)
}

func (m *MockWorkersClient) WatchStatus(
	ctx context.Context,
	eventID string,
	opts *core.WorkerStatusWatchOptions,
) (<-chan core.WorkerStatus, <-chan error, error) {
	return m.WatchStatusFn(ctx, eventID, opts)
}

func (m *MockWorkersClient) UpdateStatus(
	ctx context.Context,
	eventID string,
	status core.WorkerStatus,
	opts *core.WorkerStatusUpdateOptions,
) error {
	return m.UpdateStatusFn(ctx, eventID, status, opts)
}

func (m *MockWorkersClient) Cleanup(
	ctx context.Context,
	eventID string,
	opts *core.WorkerCleanupOptions,
) error {
	return m.CleanupFn(ctx, eventID, opts)
}

func (m *MockWorkersClient) Timeout(
	ctx context.Context,
	eventID string,
	opts *core.WorkerTimeoutOptions,
) error {
	return m.TimeoutFn(ctx, eventID, opts)
}

func (m *MockWorkersClient) Jobs() core.JobsClient {
	return m.JobsClient
}
