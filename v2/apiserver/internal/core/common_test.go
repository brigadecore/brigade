package core

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
)

// TODO: This isn't very DRY. It would be nice to figure out how to reuse these
// bits across a few different packages. The only way I (krancour) know of
// is to move these into their own package and NOT have them in files suffixed
// by _test.go. But were we to do that, Go would not recognize the functions as
// code used exclusively for testing and would therefore end up dinging us on
// coverage... for not testing the tests. :sigh:

func requireAPIVersionAndType(
	t *testing.T,
	obj interface{},
	expectedType string,
) {
	objJSON, err := json.Marshal(obj)
	require.NoError(t, err)
	objMap := map[string]interface{}{}
	err = json.Unmarshal(objJSON, &objMap)
	require.NoError(t, err)
	require.Equal(t, meta.APIVersion, objMap["apiVersion"])
	require.Equal(t, expectedType, objMap["kind"])
}

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
	GetFn        func(context.Context, string) (Event, error)
	CancelFn     func(context.Context, string) error
	CancelManyFn func(context.Context, EventsSelector,
	) (EventList, error)
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

func (m *mockEventsStore) Cancel(ctx context.Context, id string) error {
	return m.CancelFn(ctx, id)
}

func (m *mockEventsStore) CancelMany(
	ctx context.Context,
	selector EventsSelector,
) (EventList, error) {
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
	CreateProjectFn func(
		ctx context.Context,
		project Project,
	) (Project, error)
	DeleteProjectFn       func(context.Context, Project) error
	ScheduleWorkerFn      func(context.Context, Project, Event) error
	DeleteWorkerAndJobsFn func(context.Context, Project, Event) error
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

func (m *mockSubstrate) DeleteWorkerAndJobs(
	ctx context.Context,
	project Project,
	event Event,
) error {
	return m.DeleteWorkerAndJobsFn(ctx, project, event)
}
