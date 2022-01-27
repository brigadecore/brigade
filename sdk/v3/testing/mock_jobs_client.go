package testing

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3"
)

type MockJobsClient struct {
	CreateFn func(
		ctx context.Context,
		eventID string,
		job sdk.Job,
		opts *sdk.JobCreateOptions,
	) error
	StartFn func(
		ctx context.Context,
		eventID string,
		jobName string,
		opts *sdk.JobStartOptions,
	) error
	GetStatusFn func(
		ctx context.Context,
		eventID string,
		jobName string,
		opts *sdk.JobStatusGetOptions,
	) (sdk.JobStatus, error)
	WatchStatusFn func(
		ctx context.Context,
		eventID string,
		jobName string,
		opts *sdk.JobStatusWatchOptions,
	) (<-chan sdk.JobStatus, <-chan error, error)
	UpdateStatusFn func(
		ctx context.Context,
		eventID string,
		jobName string,
		status sdk.JobStatus,
		opts *sdk.JobStatusUpdateOptions,
	) error
	CleanupFn func(
		ctx context.Context,
		eventID string,
		jobName string,
		opts *sdk.JobCleanupOptions,
	) error
	TimeoutFn func(
		ctx context.Context,
		eventID string,
		jobName string,
		opts *sdk.JobTimeoutOptions,
	) error
}

func (m *MockJobsClient) Create(
	ctx context.Context,
	eventID string,
	job sdk.Job,
	opts *sdk.JobCreateOptions,
) error {
	return m.CreateFn(ctx, eventID, job, opts)
}

func (m *MockJobsClient) Start(
	ctx context.Context,
	eventID string,
	jobName string,
	opts *sdk.JobStartOptions,
) error {
	return m.StartFn(ctx, eventID, jobName, opts)
}

func (m *MockJobsClient) GetStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	opts *sdk.JobStatusGetOptions,
) (sdk.JobStatus, error) {
	return m.GetStatusFn(ctx, eventID, jobName, opts)
}

func (m *MockJobsClient) WatchStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	opts *sdk.JobStatusWatchOptions,
) (<-chan sdk.JobStatus, <-chan error, error) {
	return m.WatchStatusFn(ctx, eventID, jobName, opts)
}

func (m *MockJobsClient) UpdateStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status sdk.JobStatus,
	opts *sdk.JobStatusUpdateOptions,
) error {
	return m.UpdateStatusFn(ctx, eventID, jobName, status, opts)
}

func (m *MockJobsClient) Cleanup(
	ctx context.Context,
	eventID string,
	jobName string,
	opts *sdk.JobCleanupOptions,
) error {
	return m.CleanupFn(ctx, eventID, jobName, opts)
}

func (m *MockJobsClient) Timeout(
	ctx context.Context,
	eventID string,
	jobName string,
	opts *sdk.JobTimeoutOptions,
) error {
	return m.TimeoutFn(ctx, eventID, jobName, opts)
}
