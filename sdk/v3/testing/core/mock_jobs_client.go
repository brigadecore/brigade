package core

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3/core"
)

type MockJobsClient struct {
	CreateFn func(
		ctx context.Context,
		eventID string,
		job core.Job,
		opts *core.JobCreateOptions,
	) error
	StartFn func(
		ctx context.Context,
		eventID string,
		jobName string,
		opts *core.JobStartOptions,
	) error
	GetStatusFn func(
		ctx context.Context,
		eventID string,
		jobName string,
		opts *core.JobStatusGetOptions,
	) (core.JobStatus, error)
	WatchStatusFn func(
		ctx context.Context,
		eventID string,
		jobName string,
		opts *core.JobStatusWatchOptions,
	) (<-chan core.JobStatus, <-chan error, error)
	UpdateStatusFn func(
		ctx context.Context,
		eventID string,
		jobName string,
		status core.JobStatus,
		opts *core.JobStatusUpdateOptions,
	) error
	CleanupFn func(
		ctx context.Context,
		eventID string,
		jobName string,
		opts *core.JobCleanupOptions,
	) error
	TimeoutFn func(
		ctx context.Context,
		eventID string,
		jobName string,
		opts *core.JobTimeoutOptions,
	) error
}

func (m *MockJobsClient) Create(
	ctx context.Context,
	eventID string,
	job core.Job,
	opts *core.JobCreateOptions,
) error {
	return m.CreateFn(ctx, eventID, job, opts)
}

func (m *MockJobsClient) Start(
	ctx context.Context,
	eventID string,
	jobName string,
	opts *core.JobStartOptions,
) error {
	return m.StartFn(ctx, eventID, jobName, opts)
}

func (m *MockJobsClient) GetStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	opts *core.JobStatusGetOptions,
) (core.JobStatus, error) {
	return m.GetStatusFn(ctx, eventID, jobName, opts)
}

func (m *MockJobsClient) WatchStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	opts *core.JobStatusWatchOptions,
) (<-chan core.JobStatus, <-chan error, error) {
	return m.WatchStatusFn(ctx, eventID, jobName, opts)
}

func (m *MockJobsClient) UpdateStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status core.JobStatus,
	opts *core.JobStatusUpdateOptions,
) error {
	return m.UpdateStatusFn(ctx, eventID, jobName, status, opts)
}

func (m *MockJobsClient) Cleanup(
	ctx context.Context,
	eventID string,
	jobName string,
	opts *core.JobCleanupOptions,
) error {
	return m.CleanupFn(ctx, eventID, jobName, opts)
}

func (m *MockJobsClient) Timeout(
	ctx context.Context,
	eventID string,
	jobName string,
	opts *core.JobTimeoutOptions,
) error {
	return m.TimeoutFn(ctx, eventID, jobName, opts)
}
