package api

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	metaTesting "github.com/brigadecore/brigade/v2/apiserver/internal/meta/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestEventMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, &Event{}, EventKind)
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
	logsStore := &mockLogsStore{}
	substrate := &mockSubstrate{}
	svc, ok := NewEventsService(
		alwaysAuthorize,
		alwaysProjectAuthorize,
		projectsStore,
		eventsStore,
		logsStore,
		substrate,
	).(*eventsService)
	require.True(t, ok)
	require.NotNil(t, svc.authorize)
	require.Same(t, projectsStore, svc.projectsStore)
	require.Same(t, eventsStore, svc.eventsStore)
	require.Same(t, substrate, svc.substrate)
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
				authorize: neverAuthorize,
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
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, errors.New("projects store error")
					},
					ListSubscribersFn: func(context.Context, Event) (ProjectList, error) {
						return ProjectList{}, nil
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
			name: "create single event for specified but not subscribed project",
			event: Event{
				ProjectID: "blue-book",
			},
			service: &eventsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{
							ObjectMeta: meta.ObjectMeta{
								ID: "blue-book",
							},
						}, nil
					},
					ListSubscribersFn: func(context.Context, Event) (ProjectList, error) {
						return ProjectList{
							Items: []Project{
								{
									ObjectMeta: meta.ObjectMeta{
										ID: "orange-book",
									},
								},
							},
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
				require.Len(t, events.Items, 0)
			},
		},
		{
			name: "create single event for specified and subscribed project; " +
				"failure",
			event: Event{
				ProjectID: "blue-book",
			},
			service: &eventsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{
							ObjectMeta: meta.ObjectMeta{
								ID: "blue-book",
							},
						}, nil
					},
					ListSubscribersFn: func(context.Context, Event) (ProjectList, error) {
						return ProjectList{
							Items: []Project{
								{
									ObjectMeta: meta.ObjectMeta{
										ID: "blue-book",
									},
								},
							},
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
			name: "create single event for specified and subscribed project" +
				"success",
			event: Event{
				ProjectID: "blue-book",
			},
			service: &eventsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{
							ObjectMeta: meta.ObjectMeta{
								ID: "blue-book",
							},
						}, nil
					},
					ListSubscribersFn: func(context.Context, Event) (ProjectList, error) {
						return ProjectList{
							Items: []Project{
								{
									ObjectMeta: meta.ObjectMeta{
										ID: "blue-book",
									},
								},
							},
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
				require.Len(t, events.Items, 1)
			},
		},
		{
			name: "create multiple events for subscribed projects; failure " +
				"getting subscribed projects",
			event: Event{
				Source: "github-gateway",
				Type:   "push",
				Qualifiers: Qualifiers{
					"repo": "github.com/foo/bar",
				},
			},
			service: &eventsService{
				authorize: alwaysAuthorize,
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
				Qualifiers: Qualifiers{
					"repo": "github.com/foo/bar",
				},
			},
			service: &eventsService{
				authorize: alwaysAuthorize,
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
				Qualifiers: Qualifiers{
					"repo": "github.com/foo/bar",
				},
			},
			service: &eventsService{
				authorize: alwaysAuthorize,
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
	testEventRetryLabel := map[string]string{
		RetryLabelKey: "1234567",
	}
	testCases := []struct {
		name        string
		eventLabels map[string]string
		worker      Worker
		service     *eventsService
		assertions  func(Event, error)
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
					ScheduleWorkerFn: func(context.Context, Event) error {
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
			name:        "event retry - inherit worker spec and job",
			eventLabels: testEventRetryLabel,
			worker: Worker{
				Spec: WorkerSpec{
					DefaultConfigFiles: map[string]string{
						"defaultConfig": "myConfig",
					},
				},
				Jobs: []Job{
					{
						Name: "foo",
					},
				},
			},
			service: &eventsService{
				eventsStore: &mockEventsStore{
					CreateFn: func(context.Context, Event) error {
						return nil
					},
				},
				substrate: &mockSubstrate{
					ScheduleWorkerFn: func(context.Context, Event) error {
						return nil
					},
				},
			},
			assertions: func(event Event, err error) {
				require.NoError(t, err)
				require.Equal(t, 1, len(event.Worker.Jobs))
				require.Equal(t, Job{Name: "foo"}, event.Worker.Jobs[0])
				require.Equal(
					t,
					map[string]string{"defaultConfig": "myConfig"},
					event.Worker.Spec.DefaultConfigFiles,
				)
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
					ScheduleWorkerFn: func(context.Context, Event) error {
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
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testEvent.Labels = testCase.eventLabels
			testEvent.Worker = testCase.worker
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
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting events from store",
			service: &eventsService{
				authorize: alwaysAuthorize,
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
				authorize: alwaysAuthorize,
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
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting event from store",
			service: &eventsService{
				authorize: alwaysAuthorize,
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
				authorize: alwaysAuthorize,
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

func TestEventsServiceClone(t *testing.T) {
	testEventID := "123456789"
	testCases := []struct {
		name       string
		service    EventsService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &eventsService{
				authorize: neverAuthorize,
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
			name: "error getting event from store",
			service: &eventsService{
				authorize: alwaysAuthorize,
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
			name: "error creating cloned event",
			service: &eventsService{
				authorize: alwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
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
					return Event{}, errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
			},
		},
		{
			name: "success",
			service: &eventsService{
				authorize: alwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Source: "eventsource",
							Type:   "eventtype",
							Worker: Worker{
								Spec: WorkerSpec{
									DefaultConfigFiles: map[string]string{
										"brigade.js": "test",
									},
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					ListSubscribersFn: func(context.Context, Event) (ProjectList, error) {
						return ProjectList{
							Items: []Project{{}, {}},
						}, nil
					},
				},
				createSingleEventFn: func(
					_ context.Context,
					_ Project,
					event Event,
				) (Event, error) {
					// We expect to see a label for tracing purposes
					require.Contains(t, event.Labels, CloneLabelKey)
					// Event details like source and type should be carried over
					require.Equal(t, "eventsource", event.Source)
					require.Equal(t, "eventtype", event.Type)
					// But Worker config should not
					require.Empty(t, event.Worker.Spec)
					return Event{}, nil
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := testCase.service.Clone(
				context.Background(),
				testEventID,
			)
			testCase.assertions(err)
		})
	}
}

func TestEventsServiceUpdateSourceState(t *testing.T) {
	testEventID := "123456789"
	testCases := []struct {
		name       string
		service    EventsService
		assertions func(error)
	}{
		{
			name: "error getting event from store",
			service: &eventsService{
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
			name: "unauthorized",
			service: &eventsService{
				authorize: neverAuthorize,
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
			name: "error updating source state in store",
			service: &eventsService{
				authorize: alwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
					UpdateSourceStateFn: func(
						context.Context,
						string,
						SourceState,
					) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error updating source state of event")
			},
		},
		{
			name: "success",
			service: &eventsService{
				authorize: alwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
					UpdateSourceStateFn: func(
						context.Context,
						string,
						SourceState,
					) error {
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
			err := testCase.service.UpdateSourceState(
				context.Background(),
				testEventID,
				SourceState{},
			)
			testCase.assertions(err)
		})
	}
}

func TestEventsServiceUpdateSummary(t *testing.T) {
	testEventID := "123456789"
	testCases := []struct {
		name       string
		service    EventsService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &eventsService{
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error updating summary in store",
			service: &eventsService{
				authorize: alwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
					UpdateSummaryFn: func(
						context.Context,
						string,
						EventSummary,
					) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error updating summary of event")
			},
		},
		{
			name: "success",
			service: &eventsService{
				authorize: alwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
					UpdateSummaryFn: func(
						context.Context,
						string,
						EventSummary,
					) error {
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
			err := testCase.service.UpdateSummary(
				context.Background(),
				testEventID,
				EventSummary{},
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
				projectAuthorize: neverProjectAuthorize,
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
				projectAuthorize: alwaysProjectAuthorize,
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
				projectAuthorize: alwaysProjectAuthorize,
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
			name: "error deleting event from substrate",
			service: &eventsService{
				projectAuthorize: alwaysProjectAuthorize,
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
				projectAuthorize: alwaysProjectAuthorize,
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
				ebr, ok := err.(*meta.ErrBadRequest)
				require.True(t, ok)
				require.Equal(
					t,
					ebr.Reason,
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
				projectAuthorize: neverProjectAuthorize,
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
				projectAuthorize: alwaysProjectAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				ebr, ok := err.(*meta.ErrBadRequest)
				require.True(t, ok)
				require.Equal(
					t,
					ebr.Reason,
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
				projectAuthorize: alwaysProjectAuthorize,
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
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				eventsStore: &mockEventsStore{
					CancelManyFn: func(
						context.Context,
						EventsSelector,
					) (<-chan Event, int64, error) {
						return nil, 0, errors.New("events store error")
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
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				eventsStore: &mockEventsStore{
					CancelManyFn: func(
						context.Context,
						EventsSelector,
					) (<-chan Event, int64, error) {
						eventCh := make(chan Event)
						defer close(eventCh)
						return eventCh, 0, nil
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
				projectAuthorize: neverProjectAuthorize,
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
				projectAuthorize: alwaysProjectAuthorize,
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
				projectAuthorize: alwaysProjectAuthorize,
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
				projectAuthorize: alwaysProjectAuthorize,
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
			name: "error deleting event logs",
			service: &eventsService{
				projectAuthorize: alwaysProjectAuthorize,
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
				logsStore: &mockLogsStore{
					DeleteEventLogsFn: func(context.Context, string) error {
						return errors.New("error deleting logs")
					},
				},
				substrate: &mockSubstrate{
					DeleteWorkerAndJobsFn: func(context.Context, Project, Event) error {
						return nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error deleting logs")
			},
		},
		{
			name: "success",
			service: &eventsService{
				projectAuthorize: alwaysProjectAuthorize,
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
				logsStore: &mockLogsStore{
					DeleteEventLogsFn: func(context.Context, string) error {
						return nil
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
				ebr, ok := err.(*meta.ErrBadRequest)
				require.True(t, ok)
				require.Equal(
					t,
					ebr.Reason,
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
				projectAuthorize: neverProjectAuthorize,
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
				projectAuthorize: alwaysProjectAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				ebr, ok := err.(*meta.ErrBadRequest)
				require.True(t, ok)
				require.Equal(
					t,
					ebr.Reason,
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
				projectAuthorize: alwaysProjectAuthorize,
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
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				eventsStore: &mockEventsStore{
					DeleteManyFn: func(
						context.Context,
						EventsSelector,
					) (<-chan Event, int64, error) {
						return nil, 0, errors.New("events store error")
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
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				eventsStore: &mockEventsStore{
					DeleteManyFn: func(
						context.Context,
						EventsSelector,
					) (<-chan Event, int64, error) {
						eventCh := make(chan Event)
						defer close(eventCh)
						return eventCh, 0, nil
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

func TestEventsServiceRetry(t *testing.T) {
	testEventID := "123456789"
	testCases := []struct {
		name       string
		service    EventsService
		assertions func(error)
	}{
		{
			name: "error getting event from store",
			service: &eventsService{
				authorize: alwaysAuthorize,
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
			name: "unauthorized",
			service: &eventsService{
				authorize: neverAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Status: WorkerStatus{
									Phase: WorkerPhaseFailed,
								},
							},
						}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "original event worker has non-terminal phase",
			service: &eventsService{
				authorize: alwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Status: WorkerStatus{
									Phase: WorkerPhaseStarting,
								},
							},
						}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"non-terminal and may not yet be retried",
				)
			},
		},
		{
			name: "error creating retry event",
			service: &eventsService{
				authorize: alwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Status: WorkerStatus{
									Phase: WorkerPhaseFailed,
								},
							},
						}, nil
					},
				},
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
					return Event{}, errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
			},
		},
		{
			name: "inherit job",
			service: &eventsService{
				authorize: alwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Status: WorkerStatus{
									Phase: WorkerPhaseSucceeded,
								},
								Jobs: []Job{
									{
										Name: "foo",
										Status: &JobStatus{
											Phase: JobPhaseSucceeded,
										},
									},
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					ListSubscribersFn: func(context.Context, Event) (ProjectList, error) {
						return ProjectList{
							Items: []Project{{}, {}},
						}, nil
					},
				},
				createSingleEventFn: func(
					_ context.Context,
					_ Project,
					event Event,
				) (Event, error) {
					// We expect to inherit one job
					require.Equal(t, 1, len(event.Worker.Jobs))
					// We expect to see the job's LogsEventID field match the event ID
					require.Equal(t, testEventID, event.Worker.Jobs[0].Status.LogsEventID)
					return Event{}, nil
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "inherit job - logs ID already exists",
			service: &eventsService{
				authorize: alwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Status: WorkerStatus{
									Phase: WorkerPhaseSucceeded,
								},
								Jobs: []Job{
									{
										Name: "foo",
										Status: &JobStatus{
											Phase:       JobPhaseSucceeded,
											LogsEventID: "abcdefgh",
										},
									},
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					ListSubscribersFn: func(context.Context, Event) (ProjectList, error) {
						return ProjectList{
							Items: []Project{{}, {}},
						}, nil
					},
				},
				createSingleEventFn: func(
					_ context.Context,
					_ Project,
					event Event,
				) (Event, error) {
					// We expect to inherit one job
					require.Equal(t, 1, len(event.Worker.Jobs))
					// We expect to see the job's original LogsEventID field retained
					require.Equal(t, "abcdefgh", event.Worker.Jobs[0].Status.LogsEventID)
					return Event{}, nil
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "do not inherit job",
			service: &eventsService{
				authorize: alwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Status: WorkerStatus{
									Phase: WorkerPhaseSucceeded,
								},
								Jobs: []Job{
									{
										Name: "foo",
										Spec: JobSpec{
											PrimaryContainer: JobContainerSpec{
												WorkspaceMountPath: "/workspace",
											},
										},
										Status: &JobStatus{
											Phase: JobPhaseSucceeded,
										},
									},
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					ListSubscribersFn: func(context.Context, Event) (ProjectList, error) {
						return ProjectList{
							Items: []Project{{}, {}},
						}, nil
					},
				},
				createSingleEventFn: func(
					_ context.Context,
					_ Project,
					event Event,
				) (Event, error) {
					// We expect to inherit no jobs
					require.Equal(t, 0, len(event.Worker.Jobs))
					return Event{}, nil
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "success",
			service: &eventsService{
				authorize: alwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Source: "eventsource",
							Type:   "eventtype",
							Worker: Worker{
								Status: WorkerStatus{
									Phase: WorkerPhaseFailed,
								},
								Spec: WorkerSpec{
									DefaultConfigFiles: map[string]string{
										"brigade.js": "test",
									},
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					ListSubscribersFn: func(context.Context, Event) (ProjectList, error) {
						return ProjectList{
							Items: []Project{{}, {}},
						}, nil
					},
				},
				createSingleEventFn: func(
					_ context.Context,
					_ Project,
					event Event,
				) (Event, error) {
					// We expect to see a label for tracing purposes
					require.Contains(t, event.Labels, RetryLabelKey)
					// Event details like source and type should be carried over
					require.Equal(t, "eventsource", event.Source)
					require.Equal(t, "eventtype", event.Type)
					// As well as Worker config
					require.NotEmpty(t, event.Worker.Spec)
					require.Equal(
						t,
						map[string]string{"brigade.js": "test"},
						event.Worker.Spec.DefaultConfigFiles,
					)
					return Event{}, nil
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := testCase.service.Retry(
				context.Background(),
				testEventID,
			)
			testCase.assertions(err)
		})
	}
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
	UpdateSourceStateFn      func(context.Context, string, SourceState) error
	UpdateSummaryFn          func(context.Context, string, EventSummary) error
	CancelFn                 func(context.Context, string) error
	CancelManyFn             func(
		context.Context,
		EventsSelector,
	) (<-chan Event, int64, error)
	DeleteFn     func(context.Context, string) error
	DeleteManyFn func(
		context.Context,
		EventsSelector,
	) (<-chan Event, int64, error)
	DeleteByProjectIDFn func(context.Context, string) error
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

func (m *mockEventsStore) UpdateSourceState(
	ctx context.Context,
	id string,
	sourceState SourceState,
) error {
	return m.UpdateSourceStateFn(ctx, id, sourceState)
}

func (m *mockEventsStore) UpdateSummary(
	ctx context.Context,
	id string,
	summary EventSummary,
) error {
	return m.UpdateSummaryFn(ctx, id, summary)
}

func (m *mockEventsStore) Cancel(ctx context.Context, id string) error {
	return m.CancelFn(ctx, id)
}

func (m *mockEventsStore) CancelMany(
	ctx context.Context,
	selector EventsSelector,
) (<-chan Event, int64, error) {
	return m.CancelManyFn(ctx, selector)
}

func (m *mockEventsStore) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *mockEventsStore) DeleteMany(
	ctx context.Context,
	selector EventsSelector,
) (<-chan Event, int64, error) {
	return m.DeleteManyFn(ctx, selector)
}

func (m *mockEventsStore) DeleteByProjectID(
	ctx context.Context,
	projectID string,
) error {
	return m.DeleteByProjectIDFn(ctx, projectID)
}
