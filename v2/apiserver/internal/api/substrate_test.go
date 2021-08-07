package api

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	metaTesting "github.com/brigadecore/brigade/v2/apiserver/internal/meta/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestSubstrateWorkerCountMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(
		t,
		&SubstrateWorkerCount{},
		"SubstrateWorkerCount",
	)
}

func TestSubstrateJobCountMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(
		t,
		&SubstrateJobCount{},
		"SubstrateJobCount",
	)
}

func TestNewSubstrateService(t *testing.T) {
	substrate := &mockSubstrate{}
	svc := NewSubstrateService(alwaysAuthorize, substrate)
	require.NotNil(t, svc.(*substrateService).authorize)
	require.Same(t, substrate, svc.(*substrateService).substrate)
}

func TestSubstrateServiceCountRunningWorkers(t *testing.T) {
	const testCount = 5
	testCases := []struct {
		name       string
		service    SubstrateService
		assertions func(SubstrateWorkerCount, error)
	}{
		{
			name: "unauthorized",
			service: &substrateService{
				authorize: neverAuthorize,
			},
			assertions: func(_ SubstrateWorkerCount, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error counting workers in substrate",
			service: &substrateService{
				authorize: alwaysAuthorize,
				substrate: &mockSubstrate{
					CountRunningWorkersFn: func(
						context.Context,
					) (SubstrateWorkerCount, error) {
						return SubstrateWorkerCount{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(_ SubstrateWorkerCount, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(
					t,
					err.Error(),
					"error counting running workers on substrate",
				)
			},
		},
		{
			name: "success",
			service: &substrateService{
				authorize: alwaysAuthorize,
				substrate: &mockSubstrate{
					CountRunningWorkersFn: func(
						context.Context,
					) (SubstrateWorkerCount, error) {
						return SubstrateWorkerCount{
							Count: testCount,
						}, nil
					},
				},
			},
			assertions: func(count SubstrateWorkerCount, err error) {
				require.NoError(t, err)
				require.Equal(t, testCount, count.Count)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			count, err := testCase.service.CountRunningWorkers(context.Background())
			testCase.assertions(count, err)
		})
	}
}

func TestSubstrateServiceCountRunningJobs(t *testing.T) {
	const testCount = 5
	testCases := []struct {
		name       string
		service    SubstrateService
		assertions func(SubstrateJobCount, error)
	}{
		{
			name: "unauthorized",
			service: &substrateService{
				authorize: neverAuthorize,
			},
			assertions: func(_ SubstrateJobCount, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error counting jobs in substrate",
			service: &substrateService{
				authorize: alwaysAuthorize,
				substrate: &mockSubstrate{
					CountRunningJobsFn: func(
						context.Context,
					) (SubstrateJobCount, error) {
						return SubstrateJobCount{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(_ SubstrateJobCount, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(
					t,
					err.Error(),
					"error counting running jobs on substrate",
				)
			},
		},
		{
			name: "success",
			service: &substrateService{
				authorize: alwaysAuthorize,
				substrate: &mockSubstrate{
					CountRunningJobsFn: func(
						context.Context,
					) (SubstrateJobCount, error) {
						return SubstrateJobCount{
							Count: testCount,
						}, nil
					},
				},
			},
			assertions: func(count SubstrateJobCount, err error) {
				require.NoError(t, err)
				require.Equal(t, testCount, count.Count)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			count, err := testCase.service.CountRunningJobs(context.Background())
			testCase.assertions(count, err)
		})
	}
}

type mockSubstrate struct {
	CountRunningWorkersFn func(context.Context) (SubstrateWorkerCount, error)
	CountRunningJobsFn    func(context.Context) (SubstrateJobCount, error)
	CreateProjectFn       func(
		ctx context.Context,
		project Project,
	) (Project, error)
	DeleteProjectFn       func(context.Context, Project) error
	ScheduleWorkerFn      func(context.Context, Project, Event) error
	StartWorkerFn         func(context.Context, Project, Event) error
	StoreJobEnvironmentFn func(
		ctx context.Context,
		project Project,
		eventID string,
		jobName string,
		jobSpec JobSpec,
	) error
	ScheduleJobFn func(
		ctx context.Context,
		project Project,
		event Event,
		jobName string,
	) error
	StartJobFn func(
		ctx context.Context,
		project Project,
		event Event,
		jobName string,
	) error
	DeleteJobFn func(
		ctx context.Context,
		project Project,
		event Event,
		jobName string,
	) error
	DeleteWorkerAndJobsFn func(context.Context, Project, Event) error
}

func (m *mockSubstrate) CountRunningWorkers(
	ctx context.Context,
) (SubstrateWorkerCount, error) {
	return m.CountRunningWorkersFn(ctx)
}

func (m *mockSubstrate) CountRunningJobs(
	ctx context.Context,
) (SubstrateJobCount, error) {
	return m.CountRunningJobsFn(ctx)
}

func (m *mockSubstrate) CreateProject(
	ctx context.Context,
	project Project,
) (Project, error) {
	return m.CreateProjectFn(ctx, project)
}

func (m *mockSubstrate) DeleteProject(
	ctx context.Context,
	project Project,
) error {
	return m.DeleteProjectFn(ctx, project)
}

func (m *mockSubstrate) ScheduleWorker(
	ctx context.Context,
	project Project,
	event Event,
) error {
	return m.ScheduleWorkerFn(ctx, project, event)
}

func (m *mockSubstrate) StartWorker(
	ctx context.Context,
	project Project,
	event Event,
) error {
	return m.StartWorkerFn(ctx, project, event)
}

func (m *mockSubstrate) StoreJobEnvironment(
	ctx context.Context,
	project Project,
	eventID string,
	jobName string,
	jobSpec JobSpec,
) error {
	return m.StoreJobEnvironmentFn(ctx, project, eventID, jobName, jobSpec)
}

func (m *mockSubstrate) ScheduleJob(
	ctx context.Context,
	project Project,
	event Event,
	jobName string,
) error {
	return m.ScheduleJobFn(ctx, project, event, jobName)
}

func (m *mockSubstrate) StartJob(
	ctx context.Context,
	project Project,
	event Event,
	jobName string,
) error {
	return m.StartJobFn(ctx, project, event, jobName)
}

func (m *mockSubstrate) DeleteJob(
	ctx context.Context,
	project Project,
	event Event,
	jobName string,
) error {
	return m.DeleteJobFn(ctx, project, event, jobName)
}

func (m *mockSubstrate) DeleteWorkerAndJobs(
	ctx context.Context,
	project Project,
	event Event,
) error {
	return m.DeleteWorkerAndJobsFn(ctx, project, event)
}
