package core

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v2/core"
)

type MockWorkersClient struct {
	StartFn     func(ctx context.Context, eventID string) error
	GetStatusFn func(
		ctx context.Context,
		eventID string,
	) (core.WorkerStatus, error)
	WatchStatusFn func(
		ctx context.Context,
		eventID string,
	) (<-chan core.WorkerStatus, <-chan error, error)
	UpdateStatusFn func(
		ctx context.Context,
		eventID string,
		status core.WorkerStatus,
	) error
	CleanupFn  func(ctx context.Context, eventID string) error
	JobsClient core.JobsClient
}

func (m *MockWorkersClient) Start(ctx context.Context, eventID string) error {
	return m.StartFn(ctx, eventID)
}

func (m *MockWorkersClient) GetStatus(
	ctx context.Context,
	eventID string,
) (core.WorkerStatus, error) {
	return m.GetStatusFn(ctx, eventID)
}

func (m *MockWorkersClient) WatchStatus(
	ctx context.Context,
	eventID string,
) (<-chan core.WorkerStatus, <-chan error, error) {
	return m.WatchStatusFn(ctx, eventID)
}

func (m *MockWorkersClient) UpdateStatus(
	ctx context.Context,
	eventID string,
	status core.WorkerStatus,
) error {
	return m.UpdateStatusFn(ctx, eventID, status)
}

func (m *MockWorkersClient) Cleanup(ctx context.Context, eventID string) error {
	return m.CleanupFn(ctx, eventID)
}

func (m *MockWorkersClient) Jobs() core.JobsClient {
	return m.JobsClient
}
