package core

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
)

type mockProjectsStore struct {
	CreateFn          func(context.Context, Project) error
	ListFn            func(context.Context, meta.ListOptions) (ProjectList, error)
	ListSubscribersFn func(context.Context, Event) (ProjectList, error)
	GetFn             func(context.Context, string) (Project, error)
	UpdateFn          func(context.Context, Project) error
	DeleteFn          func(context.Context, string) error
}

func (m *mockProjectsStore) Create(ctx context.Context, project Project) error {
	return m.CreateFn(ctx, project)
}

func (m *mockProjectsStore) List(
	ctx context.Context,
	opts meta.ListOptions,
) (ProjectList, error) {
	return m.ListFn(ctx, opts)
}

func (m *mockProjectsStore) ListSubscribers(
	ctx context.Context,
	event Event,
) (ProjectList, error) {
	return m.ListSubscribersFn(ctx, event)
}

func (m *mockProjectsStore) Get(
	ctx context.Context,
	id string,
) (Project, error) {
	return m.GetFn(ctx, id)
}

func (m *mockProjectsStore) Update(ctx context.Context, project Project) error {
	return m.UpdateFn(ctx, project)
}

func (m *mockProjectsStore) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

type mockEventsStore struct {
	CreateFn func(context.Context, Event) error
	ListFn   func(
		context.Context,
		EventsSelector,
		meta.ListOptions,
	) (EventList, error)
	GetFn                    func(context.Context, string) (Event, error)
	GetByHashedWorkerTokenFn func(context.Context, string) (Event, error)
	CancelFn                 func(context.Context, string) error
	CancelManyFn             func(context.Context, EventsSelector,
	) (<-chan Event, <-chan error, error)
	DeleteFn     func(context.Context, string) error
	DeleteManyFn func(context.Context, EventsSelector) (EventList, error)
}

func (m *mockEventsStore) Create(ctx context.Context, event Event) error {
	return m.CreateFn(ctx, event)
}

func (m *mockEventsStore) List(
	ctx context.Context,
	selector EventsSelector,
	opts meta.ListOptions,
) (EventList, error) {
	return m.ListFn(ctx, selector, opts)
}

func (m *mockEventsStore) Get(ctx context.Context, id string) (Event, error) {
	return m.GetFn(ctx, id)
}

func (m *mockEventsStore) GetByHashedWorkerToken(
	ctx context.Context,
	hashedToken string,
) (Event, error) {
	return m.GetByHashedWorkerTokenFn(ctx, hashedToken)
}

func (m *mockEventsStore) Cancel(ctx context.Context, id string) error {
	return m.CancelFn(ctx, id)
}

func (m *mockEventsStore) CancelMany(
	ctx context.Context,
	selector EventsSelector,
) (<-chan Event, <-chan error, error) {
	return m.CancelManyFn(ctx, selector)
}

func (m *mockEventsStore) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *mockEventsStore) DeleteMany(
	ctx context.Context,
	selector EventsSelector,
) (EventList, error) {
	return m.DeleteManyFn(ctx, selector)
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
