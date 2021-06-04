package core

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v2/core"
)

type MockJobsClient struct {
	CreateFn func(ctx context.Context, eventID string, job core.Job) error
	StartFn  func(
		ctx context.Context,
		eventID string,
		jobName string,
	) error
	GetStatusFn func(
		ctx context.Context,
		eventID string,
		jobName string,
	) (core.JobStatus, error)
	WatchStatusFn func(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan core.JobStatus, <-chan error, error)
	UpdateStatusFn func(
		ctx context.Context,
		eventID string,
		jobName string,
		status core.JobStatus,
	) error
	CleanupFn func(ctx context.Context, eventID, jobName string) error
	TimeoutFn func(ctx context.Context, eventID, jobName string) error
}

func (m *MockJobsClient) Create(
	ctx context.Context,
	eventID string,
	job core.Job,
) error {
	return m.CreateFn(ctx, eventID, job)
}

func (m *MockJobsClient) Start(
	ctx context.Context,
	eventID string,
	jobName string,
) error {
	return m.StartFn(ctx, eventID, jobName)
}

func (m *MockJobsClient) GetStatus(
	ctx context.Context,
	eventID string,
	jobName string,
) (core.JobStatus, error) {
	return m.GetStatusFn(ctx, eventID, jobName)
}

func (m *MockJobsClient) WatchStatus(
	ctx context.Context,
	eventID string,
	jobName string,
) (<-chan core.JobStatus, <-chan error, error) {
	return m.WatchStatusFn(ctx, eventID, jobName)
}

func (m *MockJobsClient) UpdateStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status core.JobStatus,
) error {
	return m.UpdateStatusFn(ctx, eventID, jobName, status)
}

func (m *MockJobsClient) Cleanup(
	ctx context.Context,
	eventID string,
	jobName string,
) error {
	return m.CleanupFn(ctx, eventID, jobName)
}

func (m *MockJobsClient) Timeout(
	ctx context.Context,
	eventID string,
	jobName string,
) error {
	return m.TimeoutFn(ctx, eventID, jobName)
}
