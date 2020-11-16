package core

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewWorkersService(t *testing.T) {
	projectsStore := &mockProjectsStore{}
	eventsStore := &mockEventsStore{}
	workersStore := &mockWorkersStore{}
	substrate := &mockSubstrate{}
	svc := NewWorkersService(projectsStore, eventsStore, workersStore, substrate)
	require.Same(t, projectsStore, svc.(*workersService).projectsStore)
	require.Same(t, eventsStore, svc.(*workersService).eventsStore)
	require.Same(t, workersStore, svc.(*workersService).workersStore)
	require.Same(t, substrate, svc.(*workersService).substrate)
}

func TestWorkersServiceStart(t *testing.T) {
	const testEventID = "123456789"
	testCases := []struct {
		name       string
		service    WorkersService
		assertions func(error)
	}{
		{
			name: "error getting event from store",
			service: &workersService{
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error retrieving event")
				require.Contains(t, err.Error(), "something went wrong")
			},
		},
		{
			name: "worker is not currently pending",
			service: &workersService{
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Status: WorkerStatus{
									Phase: WorkerPhaseRunning,
								},
							},
						}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "worker has already been started")
			},
		},
		{
			name: "error getting project from store",
			service: &workersService{
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Status: WorkerStatus{
									Phase: WorkerPhasePending,
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error retrieving project")
				require.Contains(t, err.Error(), "something went wrong")
			},
		},
		{
			name: "error starting worker on substrate",
			service: &workersService{
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Status: WorkerStatus{
									Phase: WorkerPhasePending,
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				substrate: &mockSubstrate{
					StartWorkerFn: func(c context.Context, p Project, e Event) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error starting worker for event")
				require.Contains(t, err.Error(), "something went wrong")
			},
		},
		{
			name: "success",
			service: &workersService{
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Status: WorkerStatus{
									Phase: WorkerPhasePending,
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				substrate: &mockSubstrate{
					StartWorkerFn: func(c context.Context, p Project, e Event) error {
						return nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.service.Start(context.Background(), testEventID)
			testCase.assertions(err)
		})
	}
}

func TestWorkersServiceUpdateStatus(t *testing.T) {
	testEventID := "123456789"
	testCases := []struct {
		name       string
		service    WorkersService
		assertions func(error)
	}{
		{
			name: "error updating worker in store",
			service: &workersService{
				workersStore: &mockWorkersStore{
					UpdateStatusFn: func(context.Context, string, WorkerStatus) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error updating status of event")
			},
		},
		{
			name: "success",
			service: &workersService{
				workersStore: &mockWorkersStore{
					UpdateStatusFn: func(context.Context, string, WorkerStatus) error {
						return nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.service.UpdateStatus(
				context.Background(),
				testEventID,
				WorkerStatus{},
			)
			testCase.assertions(err)
		})
	}
}

type mockWorkersStore struct {
	UpdateStatusFn func(
		ctx context.Context,
		eventID string,
		status WorkerStatus,
	) error
}

func (m *mockWorkersStore) UpdateStatus(
	ctx context.Context,
	eventID string,
	status WorkerStatus,
) error {
	return m.UpdateStatusFn(ctx, eventID, status)
}
