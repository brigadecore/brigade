package testing

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3"
)

type MockWorkersClient struct {
	StartFn func(
		ctx context.Context,
		eventID string,
		opts *sdk.WorkerStartOptions,
	) error
	GetStatusFn func(
		ctx context.Context,
		eventID string,
		opts *sdk.WorkerStatusGetOptions,
	) (sdk.WorkerStatus, error)
	WatchStatusFn func(
		ctx context.Context,
		eventID string,
		opts *sdk.WorkerStatusWatchOptions,
	) (<-chan sdk.WorkerStatus, <-chan error, error)
	UpdateStatusFn func(
		ctx context.Context,
		eventID string,
		status sdk.WorkerStatus,
		opts *sdk.WorkerStatusUpdateOptions,
	) error
	CleanupFn func(
		ctx context.Context,
		eventID string,
		opts *sdk.WorkerCleanupOptions,
	) error
	TimeoutFn func(
		ctx context.Context,
		eventID string,
		opts *sdk.WorkerTimeoutOptions,
	) error
	JobsClient sdk.JobsClient
}

func (m *MockWorkersClient) Start(
	ctx context.Context,
	eventID string,
	opts *sdk.WorkerStartOptions,
) error {
	return m.StartFn(ctx, eventID, opts)
}

func (m *MockWorkersClient) GetStatus(
	ctx context.Context,
	eventID string,
	opts *sdk.WorkerStatusGetOptions,
) (sdk.WorkerStatus, error) {
	return m.GetStatusFn(ctx, eventID, opts)
}

func (m *MockWorkersClient) WatchStatus(
	ctx context.Context,
	eventID string,
	opts *sdk.WorkerStatusWatchOptions,
) (<-chan sdk.WorkerStatus, <-chan error, error) {
	return m.WatchStatusFn(ctx, eventID, opts)
}

func (m *MockWorkersClient) UpdateStatus(
	ctx context.Context,
	eventID string,
	status sdk.WorkerStatus,
	opts *sdk.WorkerStatusUpdateOptions,
) error {
	return m.UpdateStatusFn(ctx, eventID, status, opts)
}

func (m *MockWorkersClient) Cleanup(
	ctx context.Context,
	eventID string,
	opts *sdk.WorkerCleanupOptions,
) error {
	return m.CleanupFn(ctx, eventID, opts)
}

func (m *MockWorkersClient) Timeout(
	ctx context.Context,
	eventID string,
	opts *sdk.WorkerTimeoutOptions,
) error {
	return m.TimeoutFn(ctx, eventID, opts)
}

func (m *MockWorkersClient) Jobs() sdk.JobsClient {
	return m.JobsClient
}
