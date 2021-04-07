package core

import (
	"context"
	"errors"
	"testing"
	"time"

	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
)

func TestNewWorkersService(t *testing.T) {
	projectsStore := &mockProjectsStore{}
	eventsStore := &mockEventsStore{}
	workersStore := &mockWorkersStore{}
	substrate := &mockSubstrate{}
	svc := NewWorkersService(
		libAuthz.AlwaysAuthorize,
		projectsStore,
		eventsStore,
		workersStore,
		substrate,
	)
	require.NotNil(t, svc.(*workersService).authorize)
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
			name: "unauthorized",
			service: &workersService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting event from store",
			service: &workersService{
				authorize: libAuthz.AlwaysAuthorize,
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
				authorize: libAuthz.AlwaysAuthorize,
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
				authorize: libAuthz.AlwaysAuthorize,
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
			name: "error updating worker status",
			service: &workersService{
				authorize: libAuthz.AlwaysAuthorize,
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
			name: "error starting worker on substrate",
			service: &workersService{
				authorize: libAuthz.AlwaysAuthorize,
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
				workersStore: &mockWorkersStore{
					UpdateStatusFn: func(context.Context, string, WorkerStatus) error {
						return nil
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
				authorize: libAuthz.AlwaysAuthorize,
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
				workersStore: &mockWorkersStore{
					UpdateStatusFn: func(context.Context, string, WorkerStatus) error {
						return nil
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

func TestWorkersServiceGetStatus(t *testing.T) {
	const testEventID = "123456789"
	testWorkerStatus := WorkerStatus{
		Phase: WorkerPhaseRunning,
	}
	testCases := []struct {
		name       string
		service    WorkersService
		assertions func(WorkerStatus, error)
	}{
		{
			name: "unauthorized",
			service: &workersService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(_ WorkerStatus, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting event from store",
			service: &workersService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(_ WorkerStatus, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving event")
			},
		},
		{
			name: "success",
			service: &workersService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Status: testWorkerStatus,
							},
						}, nil
					},
				},
			},
			assertions: func(status WorkerStatus, err error) {
				require.NoError(t, err)
				require.Equal(t, testWorkerStatus, status)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.assertions(
				testCase.service.GetStatus(context.Background(), testEventID),
			)
		})
	}
}

func TestWorkersServiceWatchStatus(t *testing.T) {
	const testEventID = "123456789"
	testWorkerStatus := WorkerStatus{
		Phase: WorkerPhaseRunning,
	}
	testCases := []struct {
		name       string
		service    WorkersService
		assertions func(context.Context, <-chan WorkerStatus, error)
	}{
		{
			name: "unauthorized",
			service: &workersService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(_ context.Context, _ <-chan WorkerStatus, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting event from store",
			service: &workersService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(_ context.Context, _ <-chan WorkerStatus, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving event")
			},
		},
		{
			name: "success",
			service: &workersService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Status: testWorkerStatus,
							},
						}, nil
					},
				},
			},
			assertions: func(
				ctx context.Context,
				statusCh <-chan WorkerStatus,
				err error,
			) {
				require.NoError(t, err)
				select {
				case status := <-statusCh:
					require.Equal(t, testWorkerStatus, status)
				case <-ctx.Done():
					require.Fail(t, "didn't receive status update over channel")
				}
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			statusCh, err := testCase.service.WatchStatus(ctx, testEventID)
			testCase.assertions(ctx, statusCh, err)
			cancel()
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
			name: "unauthorized",
			service: &workersService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error retrieving event from store",
			service: &workersService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving event")
			},
		},
		{
			name: "error updating worker in store",
			service: &workersService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
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
			name: "worker's phase already terminal",
			service: &workersService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Status: WorkerStatus{
									Phase: WorkerPhaseCanceled,
								},
							},
						}, nil
					},
				},
				workersStore: &mockWorkersStore{
					UpdateStatusFn: func(context.Context, string, WorkerStatus) error {
						require.Fail(
							t,
							"UpdateStatusFn should not have been called, but was",
						)
						return nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"worker has already reached a terminal phase",
				)
			},
		},
		{
			name: "success",
			service: &workersService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
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

func TestWorkersServiceCleanup(t *testing.T) {
	const testEventID = "123456789"
	testCases := []struct {
		name       string
		service    WorkersService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &workersService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting event from store",
			service: &workersService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving event")
			},
		},
		{
			name: "error getting project from store",
			service: &workersService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
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
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving project")
			},
		},
		{
			name: "error deleting worker and jobs from substrate",
			service: &workersService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				substrate: &mockSubstrate{
					DeleteWorkerAndJobsFn: func(context.Context, Project, Event) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error deleting event")
			},
		},
		{
			name: "success",
			service: &workersService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				substrate: &mockSubstrate{
					DeleteWorkerAndJobsFn: func(context.Context, Project, Event) error {
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
			err := testCase.service.Cleanup(context.Background(), testEventID)
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
