package core

import (
	"context"
	"errors"
	"testing"

	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	metaTesting "github.com/brigadecore/brigade/v2/apiserver/internal/meta/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestEventMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, &Event{}, "Event")
}

func TestEventListMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, &EventList{}, "EventList")
}

func TestCancelManyEventsResultMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(
		t,
		&CancelManyEventsResult{},
		"CancelManyEventsResult",
	)
}

func TestDeleteManyEventsResultMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(
		t,
		&DeleteManyEventsResult{},
		"DeleteManyEventsResult",
	)
}

func TestNewEventsService(t *testing.T) {
	projectsStore := &mockProjectsStore{}
	eventsStore := &mockEventsStore{}
	jobsStore := &mockJobsStore{}
	substrate := &mockSubstrate{}
	svc := NewEventsService(
		libAuthz.AlwaysAuthorize,
		projectsStore,
		eventsStore,
		jobsStore,
		substrate,
	)
	require.NotNil(t, svc.(*eventsService).authorize)
	require.Same(t, projectsStore, svc.(*eventsService).projectsStore)
	require.Same(t, eventsStore, svc.(*eventsService).eventsStore)
	require.Same(t, substrate, svc.(*eventsService).substrate)
}

func TestEventsServiceCreate(t *testing.T) {
	testCases := []struct {
		name       string
		event      Event
		service    EventsService
		assertions func(EventList, error)
	}{
		{
			name: "unauthorized",
			service: &eventsService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(_ EventList, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "create single event for specified project; error getting " +
				"project from store",
			event: Event{
				ProjectID: "blue-book",
			},
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, errors.New("projects store error")
					},
				},
			},
			assertions: func(_ EventList, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error retrieving project")
				require.Contains(t, err.Error(), "projects store error")
			},
		},
		{
			name: "create single event for specified project; create single " +
				"event failure ",
			event: Event{
				ProjectID: "blue-book",
			},
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				createSingleEventFn: func(
					context.Context,
					Project,
					Event,
				) (Event, error) {
					return Event{}, errors.New("error creating single event")
				},
			},
			assertions: func(_ EventList, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error creating single event")
			},
		},
		{
			name: "create single event for specified project; success",
			event: Event{
				ProjectID: "blue-book",
			},
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				createSingleEventFn: func(
					context.Context,
					Project,
					Event,
				) (Event, error) {
					return Event{}, nil
				},
			},
			assertions: func(events EventList, err error) {
				require.NoError(t, err)
				require.Len(t, events.Items, 1)
			},
		},
		{
			name: "create multiple events for subscribed projects; failure " +
				"getting subscribed projects",
			event: Event{
				Source: "github-gateway",
				Type:   "push",
				Labels: Labels{
					"repo": "github.com/foo/bar",
				},
			},
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				projectsStore: &mockProjectsStore{
					ListSubscribersFn: func(context.Context, Event) (ProjectList, error) {
						return ProjectList{}, errors.New(
							"error getting subscribed projects",
						)
					},
				},
			},
			assertions: func(_ EventList, err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"error retrieving subscribed projects from store",
				)
				require.Contains(t, err.Error(), "error getting subscribed projects")
			},
		},
		{
			name: "create multiple events for subscribed projects; create single " +
				"event failure ",
			event: Event{
				Source: "github-gateway",
				Type:   "push",
				Labels: Labels{
					"repo": "github.com/foo/bar",
				},
			},
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				projectsStore: &mockProjectsStore{
					ListSubscribersFn: func(context.Context, Event) (ProjectList, error) {
						return ProjectList{
							Items: []Project{{}, {}},
						}, nil
					},
				},
				createSingleEventFn: func(
					context.Context,
					Project,
					Event,
				) (Event, error) {
					return Event{}, errors.New("error creating single event")
				},
			},
			assertions: func(_ EventList, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error creating single event")
			},
		},
		{
			name: "create multiple events for subscribed projects; success",
			event: Event{
				Source: "github-gateway",
				Type:   "push",
				Labels: Labels{
					"repo": "github.com/foo/bar",
				},
			},
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				projectsStore: &mockProjectsStore{
					ListSubscribersFn: func(context.Context, Event) (ProjectList, error) {
						return ProjectList{
							Items: []Project{{}, {}},
						}, nil
					},
				},
				createSingleEventFn: func(
					context.Context,
					Project,
					Event,
				) (Event, error) {
					return Event{}, nil
				},
			},
			assertions: func(events EventList, err error) {
				require.NoError(t, err)
				require.Len(t, events.Items, 2)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			events, err :=
				testCase.service.Create(context.Background(), testCase.event)
			testCase.assertions(events, err)
		})
	}
}

func TestEventsServiceCreateSingleEvent(t *testing.T) {
	testProject := Project{}
	testEvent := Event{
		Git: &GitDetails{
			CloneURL: "github.com/foo/bar.git",
			Commit:   "123456789",
			Ref:      "dev",
		},
	}
	testCases := []struct {
		name       string
		service    *eventsService
		assertions func(Event, error)
	}{
		{
			name: "error creating event in store",
			service: &eventsService{
				eventsStore: &mockEventsStore{
					CreateFn: func(context.Context, Event) error {
						return errors.New("store error")
					},
				},
			},
			assertions: func(_ Event, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error storing new event")
				require.Contains(t, err.Error(), "store error")
			},
		},
		{
			name: "error scheduling worker in substrate",
			service: &eventsService{
				eventsStore: &mockEventsStore{
					CreateFn: func(context.Context, Event) error {
						return nil
					},
				},
				substrate: &mockSubstrate{
					ScheduleWorkerFn: func(c context.Context, p Project, e Event) error {
						return errors.New("substrate error")
					},
				},
			},
			assertions: func(_ Event, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error scheduling event")
				require.Contains(t, err.Error(), "substrate error")
			},
		},
		{
			name: "success",
			service: &eventsService{
				eventsStore: &mockEventsStore{
					CreateFn: func(context.Context, Event) error {
						return nil
					},
				},
				substrate: &mockSubstrate{
					ScheduleWorkerFn: func(c context.Context, p Project, e Event) error {
						return nil
					},
				},
			},
			assertions: func(event Event, err error) {
				require.NoError(t, err)
				// Make sure the Event looks like what we expect-- i.e. ID generated,
				// default applied, etc.
				require.NotEmpty(t, event.ID)
				require.Equal(t, defaultWorkspaceSize, event.Worker.Spec.WorkspaceSize)
				require.NotNil(t, event.Worker.Spec.Git)
				require.Equal(t, testEvent.Git.CloneURL, event.Worker.Spec.Git.CloneURL)
				require.Equal(t, testEvent.Git.Commit, event.Worker.Spec.Git.Commit)
				require.Equal(t, testEvent.Git.Ref, event.Worker.Spec.Git.Ref)
				require.Equal(t, LogLevelInfo, event.Worker.Spec.LogLevel)
				require.Equal(t, WorkerPhasePending, event.Worker.Status.Phase)
				require.NotEmpty(t, event.Worker.Token)
				require.NotEmpty(t, event.Worker.HashedToken)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			event, err := testCase.service.createSingleEvent(
				context.Background(),
				testProject,
				testEvent,
			)
			testCase.assertions(event, err)
		})
	}
}

func TestEventsServiceList(t *testing.T) {
	testCases := []struct {
		name       string
		service    EventsService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &eventsService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting events from store",
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					ListFn: func(
						context.Context,
						EventsSelector,
						meta.ListOptions,
					) (EventList, error) {
						return EventList{}, errors.New("error listing events")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error listing events")
				require.Contains(t, err.Error(), "error retrieving events from store")
			},
		},
		{
			name: "success",
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					ListFn: func(
						context.Context,
						EventsSelector,
						meta.ListOptions,
					) (EventList, error) {
						return EventList{}, nil
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
			_, err := testCase.service.List(
				context.Background(),
				EventsSelector{},
				meta.ListOptions{},
			)
			testCase.assertions(err)
		})
	}
}

func TestEventsServiceGet(t *testing.T) {
	testCases := []struct {
		name       string
		service    EventsService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &eventsService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting event from store",
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, errors.New("error getting event")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error getting event")
				require.Contains(t, err.Error(), "error retrieving event")
			},
		},
		{
			name: "success",
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
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
			_, err := testCase.service.Get(context.Background(), "foobar")
			testCase.assertions(err)
		})
	}
}

func TestEventsServiceGetByWorkerToken(t *testing.T) {
	testCases := []struct {
		name       string
		service    EventsService
		assertions func(error)
	}{
		{
			name: "error getting event from store",
			service: &eventsService{
				eventsStore: &mockEventsStore{
					GetByHashedWorkerTokenFn: func(
						context.Context,
						string,
					) (Event, error) {
						return Event{}, errors.New("error getting event")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error getting event")
				require.Contains(t, err.Error(), "error retrieving event")
			},
		},
		{
			name: "success",
			service: &eventsService{
				eventsStore: &mockEventsStore{
					GetByHashedWorkerTokenFn: func(
						context.Context,
						string,
					) (Event, error) {
						return Event{}, nil
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
			_, err := testCase.service.GetByWorkerToken(
				context.Background(),
				"foobar",
			)
			testCase.assertions(err)
		})
	}
}

func TestEventsServiceCancel(t *testing.T) {
	const testEventID = "123456789"
	testCases := []struct {
		name       string
		service    EventsService
		assertions func(error)
	}{
		{
			name: "error retrieving event from store",
			service: &eventsService{
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, errors.New("events store error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error retrieving event")
				require.Contains(t, err.Error(), "events store error")
			},
		},
		{
			name: "unauthorized",
			service: &eventsService{
				authorize: libAuthz.NeverAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error retrieving project from store",
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, errors.New("projects store error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error retrieving project")
				require.Contains(t, err.Error(), "projects store error")
			},
		},
		{
			name: "error canceling event in store",
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
					CancelFn: func(context.Context, string) error {
						return errors.New("events store error")
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error canceling event")
				require.Contains(t, err.Error(), "events store error")
			},
		},
		{
			name: "error canceling worker jobs in store",
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: map[string]Job{
									"italian": {},
								},
							},
						}, nil
					},
					CancelFn: func(context.Context, string) error {
						return nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				jobsStore: &mockJobsStore{
					CancelFn: func(ctx context.Context,
						eventID string,
						jobName string,
					) error {
						return errors.New("jobs store error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error canceling event")
				require.Contains(t, err.Error(), "jobs store error")
			},
		},
		{
			name: "error deleting event from substrate",
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
					CancelFn: func(context.Context, string) error {
						return nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				substrate: &mockSubstrate{
					DeleteWorkerAndJobsFn: func(context.Context, Project, Event) error {
						return errors.New("substrate error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error deleting event")
				require.Contains(t, err.Error(), "worker and jobs")
				require.Contains(t, err.Error(), "from the substrate")
				require.Contains(t, err.Error(), "substrate error")
			},
		},
		{
			name: "success",
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
					CancelFn: func(context.Context, string) error {
						return nil
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
			err := testCase.service.Cancel(context.Background(), testEventID)
			testCase.assertions(err)
		})
	}
}

func TestEventsServiceCancelMany(t *testing.T) {
	testCases := []struct {
		name       string
		selector   EventsSelector
		service    EventsService
		assertions func(error)
	}{
		{
			name:     "request not qualified by project",
			selector: EventsSelector{},
			service:  &eventsService{},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrBadRequest{}, err)
				require.Equal(
					t,
					err.(*meta.ErrBadRequest).Reason,
					"Requests to cancel multiple events must be qualified by project.",
				)
			},
		},
		{
			name: "unauthorized",
			selector: EventsSelector{
				ProjectID: "blue-book",
			},
			service: &eventsService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "request not qualified by worker phase",
			selector: EventsSelector{
				ProjectID: "blue-book",
			},
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrBadRequest{}, err)
				require.Equal(
					t,
					err.(*meta.ErrBadRequest).Reason,
					"Requests to cancel multiple events must be qualified by "+
						"worker phase(s).",
				)
			},
		},
		{
			name: "error getting project from store",
			selector: EventsSelector{
				ProjectID:    "blue-book",
				WorkerPhases: []WorkerPhase{WorkerPhaseFailed},
			},
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, errors.New("error getting project")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error retrieving project")
				require.Contains(t, err.Error(), "error getting project")
			},
		},
		{
			name: "error canceling events in store",
			selector: EventsSelector{
				ProjectID:    "blue-book",
				WorkerPhases: []WorkerPhase{WorkerPhaseFailed},
			},
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				eventsStore: &mockEventsStore{
					CancelManyFn: func(
						context.Context,
						EventsSelector,
					) (EventList, error) {
						return EventList{}, errors.New("events store error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error canceling events in store")
				require.Contains(t, err.Error(), "events store error")
			},
		},
		{
			name: "success",
			selector: EventsSelector{
				ProjectID:    "blue-book",
				WorkerPhases: []WorkerPhase{WorkerPhaseFailed},
			},
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				eventsStore: &mockEventsStore{
					CancelManyFn: func(
						context.Context,
						EventsSelector,
					) (EventList, error) {
						return EventList{}, nil
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
			_, err :=
				testCase.service.CancelMany(context.Background(), testCase.selector)
			testCase.assertions(err)
		})
	}
}

func TestEventsServiceDelete(t *testing.T) {
	const testEventID = "123456789"
	testCases := []struct {
		name       string
		service    EventsService
		assertions func(error)
	}{
		{
			name: "error retrieving event from store",
			service: &eventsService{
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, errors.New("events store error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error retrieving event")
				require.Contains(t, err.Error(), "events store error")
			},
		},
		{
			name: "unauthorized",
			service: &eventsService{
				authorize: libAuthz.NeverAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error retrieving project from store",
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, errors.New("projects store error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error retrieving project")
				require.Contains(t, err.Error(), "projects store error")
			},
		},
		{
			name: "error deleting event from store",
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return errors.New("events store error")
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error deleting event")
				require.Contains(t, err.Error(), "events store error")
			},
		},
		{
			name: "error deleting event from substrate",
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				substrate: &mockSubstrate{
					DeleteWorkerAndJobsFn: func(context.Context, Project, Event) error {
						return errors.New("substrate error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error deleting event")
				require.Contains(t, err.Error(), "worker and jobs")
				require.Contains(t, err.Error(), "from the substrate")
				require.Contains(t, err.Error(), "substrate error")
			},
		},
		{
			name: "success",
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return nil
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
			err := testCase.service.Delete(context.Background(), testEventID)
			testCase.assertions(err)
		})
	}
}

func TestEventsServiceDeleteMany(t *testing.T) {
	testCases := []struct {
		name       string
		selector   EventsSelector
		service    EventsService
		assertions func(error)
	}{
		{
			name:     "request not qualified by project",
			selector: EventsSelector{},
			service:  &eventsService{},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrBadRequest{}, err)
				require.Equal(
					t,
					err.(*meta.ErrBadRequest).Reason,
					"Requests to delete multiple events must be qualified by project.",
				)
			},
		},
		{
			name: "unauthorized",
			selector: EventsSelector{
				ProjectID: "blue-book",
			},
			service: &eventsService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "request not qualified by worker phase",
			selector: EventsSelector{
				ProjectID: "blue-book",
			},
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrBadRequest{}, err)
				require.Equal(
					t,
					err.(*meta.ErrBadRequest).Reason,
					"Requests to delete multiple events must be qualified by "+
						"worker phase(s).",
				)
			},
		},
		{
			name: "error getting project from store",
			selector: EventsSelector{
				ProjectID:    "blue-book",
				WorkerPhases: []WorkerPhase{WorkerPhaseFailed},
			},
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, errors.New("error getting project")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error retrieving project")
				require.Contains(t, err.Error(), "error getting project")
			},
		},
		{
			name: "error deleting events from store",
			selector: EventsSelector{
				ProjectID:    "blue-book",
				WorkerPhases: []WorkerPhase{WorkerPhaseFailed},
			},
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				eventsStore: &mockEventsStore{
					DeleteManyFn: func(
						context.Context,
						EventsSelector,
					) (EventList, error) {
						return EventList{}, errors.New("events store error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error deleting events from store")
				require.Contains(t, err.Error(), "events store error")
			},
		},
		{
			name: "success",
			selector: EventsSelector{
				ProjectID:    "blue-book",
				WorkerPhases: []WorkerPhase{WorkerPhaseFailed},
			},
			service: &eventsService{
				authorize: libAuthz.AlwaysAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				eventsStore: &mockEventsStore{
					DeleteManyFn: func(
						context.Context,
						EventsSelector,
					) (EventList, error) {
						return EventList{}, nil
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
			_, err :=
				testCase.service.DeleteMany(context.Background(), testCase.selector)
			testCase.assertions(err)
		})
	}
}
