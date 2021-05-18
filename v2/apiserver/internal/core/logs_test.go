package core

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	metaTesting "github.com/brigadecore/brigade/v2/apiserver/internal/meta/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestLogEntryMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, &LogEntry{}, "LogEntry")
}

func TestLogsService(t *testing.T) {
	projectsStore := &mockProjectsStore{}
	eventsStore := &mockEventsStore{}
	warmLogsStore := &mockLogsStore{}
	coolLogsStore := &mockLogsStore{}
	svc := NewLogsService(
		alwaysProjectAuthorize,
		projectsStore,
		eventsStore,
		warmLogsStore,
		coolLogsStore,
	)
	require.NotNil(t, svc.(*logsService).projectAuthorize)
	require.Same(t, projectsStore, svc.(*logsService).projectsStore)
	require.Same(t, eventsStore, svc.(*logsService).eventsStore)
	require.Same(t, warmLogsStore, svc.(*logsService).warmLogsStore)
	require.Same(t, coolLogsStore, svc.(*logsService).coolLogsStore)
}

func TestLogsServiceStream(t *testing.T) {
	const testEventID = "123456789"
	testCases := []struct {
		name       string
		service    LogsService
		selector   LogsSelector
		assertions func(<-chan LogEntry, error)
	}{
		{
			name: "invalid worker container name",
			selector: LogsSelector{
				Container: "foo",
			},
			service: &logsService{},
			assertions: func(_ <-chan LogEntry, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
				require.Equal(t, "WorkerContainer", err.(*meta.ErrNotFound).Type)
				require.Equal(t, "foo", err.(*meta.ErrNotFound).ID)
			},
		},
		{
			name:     "error retrieving event from store",
			selector: LogsSelector{},
			service: &logsService{
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(_ <-chan LogEntry, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving event")
			},
		},
		{
			name: "unauthorized",
			service: &logsService{
				projectAuthorize: neverProjectAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
			},
			assertions: func(_ <-chan LogEntry, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "invalid job name",
			selector: LogsSelector{
				Job: "foo",
			},
			service: &logsService{
				projectAuthorize: alwaysProjectAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
			},
			assertions: func(_ <-chan LogEntry, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
				require.Equal(t, JobKind, err.(*meta.ErrNotFound).Type)
				require.Equal(t, "foo", err.(*meta.ErrNotFound).ID)
			},
		},
		{
			name: "invalid job container name",
			selector: LogsSelector{
				Job:       "foo",
				Container: "bar",
			},
			service: &logsService{
				projectAuthorize: alwaysProjectAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: []Job{
									{
										Name: "foo",
									},
								},
							},
						}, nil
					},
				},
			},
			assertions: func(_ <-chan LogEntry, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
				require.Equal(t, "JobContainer", err.(*meta.ErrNotFound).Type)
				require.Equal(t, "bar", err.(*meta.ErrNotFound).ID)
			},
		},
		{
			name:     "error retrieving project from store",
			selector: LogsSelector{},
			service: &logsService{
				projectAuthorize: alwaysProjectAuthorize,
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
			assertions: func(_ <-chan LogEntry, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving project")
			},
		},
		{
			name: "error retrieving inherited logs - event not found",
			selector: LogsSelector{
				Job: "italian",
			},
			service: &logsService{
				projectAuthorize: alwaysProjectAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(_ context.Context, eventID string) (Event, error) {
						if eventID == testEventID {
							return Event{
								Worker: Worker{
									Jobs: []Job{
										{
											Name: "italian",
											Status: &JobStatus{
												LogsEventID: "abcdefg",
											},
										},
									},
								},
							}, nil
						}
						// Else, return ErrNotFound to simulate not being able to locate
						// logs for the "abcdefg" LogsEventID
						return Event{}, &meta.ErrNotFound{}
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(_ <-chan LogEntry, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
				require.Contains(t, err.Error(), "Unable to retrieve logs for job")
			},
		},
		{
			name: "error retrieving inherited logs - something went wrong",
			selector: LogsSelector{
				Job: "italian",
			},
			service: &logsService{
				projectAuthorize: alwaysProjectAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(_ context.Context, eventID string) (Event, error) {
						if eventID == testEventID {
							return Event{
								Worker: Worker{
									Jobs: []Job{
										{
											Name: "italian",
											Status: &JobStatus{
												LogsEventID: "abcdefg",
											},
										},
									},
								},
							}, nil
						}
						// Else, return a generic error when attempting to locate logs
						// for the "abcdefg" LogsEventID
						return Event{}, errors.New("something went wrong")
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(_ <-chan LogEntry, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving logs")
			},
		},
		{
			name:     "warm logs succeed",
			selector: LogsSelector{},
			service: &logsService{
				projectAuthorize: alwaysProjectAuthorize,
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
				warmLogsStore: &mockLogsStore{
					StreamLogsFn: func(
						context.Context,
						Project,
						Event,
						LogsSelector,
						LogStreamOptions,
					) (<-chan LogEntry, error) {
						return make(chan LogEntry), nil
					},
				},
				coolLogsStore: &mockLogsStore{
					StreamLogsFn: func(
						context.Context,
						Project,
						Event,
						LogsSelector,
						LogStreamOptions,
					) (<-chan LogEntry, error) {
						require.Fail(t, "cool logs should not have been accessed, but were")
						return nil, nil
					},
				},
			},
			assertions: func(logCh <-chan LogEntry, err error) {
				require.NoError(t, err)
				require.NotNil(t, logCh)
			},
		},
		{
			name:     "warm logs store has unexpected error",
			selector: LogsSelector{},
			service: &logsService{
				projectAuthorize: alwaysProjectAuthorize,
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
				warmLogsStore: &mockLogsStore{
					StreamLogsFn: func(
						context.Context,
						Project,
						Event,
						LogsSelector,
						LogStreamOptions,
					) (<-chan LogEntry, error) {
						return nil, errors.New("something went wrong")
					},
				},
			},
			assertions: func(logCh <-chan LogEntry, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
			},
		},
		{
			name:     "cool logs succeed",
			selector: LogsSelector{},
			service: &logsService{
				projectAuthorize: alwaysProjectAuthorize,
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
				warmLogsStore: &mockLogsStore{
					StreamLogsFn: func(
						context.Context,
						Project,
						Event,
						LogsSelector,
						LogStreamOptions,
					) (<-chan LogEntry, error) {
						// This error will signal the service to fall back to cool logs
						return nil, &meta.ErrNotFound{}
					},
				},
				coolLogsStore: &mockLogsStore{
					StreamLogsFn: func(
						context.Context,
						Project,
						Event,
						LogsSelector,
						LogStreamOptions,
					) (<-chan LogEntry, error) {
						return make(chan LogEntry), nil
					},
				},
			},
			assertions: func(logCh <-chan LogEntry, err error) {
				require.NoError(t, err)
				require.NotNil(t, logCh)
			},
		},
		{
			name:     "warm and cool logs both fail",
			selector: LogsSelector{},
			service: &logsService{
				projectAuthorize: alwaysProjectAuthorize,
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
				warmLogsStore: &mockLogsStore{
					StreamLogsFn: func(
						context.Context,
						Project,
						Event,
						LogsSelector,
						LogStreamOptions,
					) (<-chan LogEntry, error) {
						return nil, errors.New("something went wrong")
					},
				},
				coolLogsStore: &mockLogsStore{
					StreamLogsFn: func(
						context.Context,
						Project,
						Event,
						LogsSelector,
						LogStreamOptions,
					) (<-chan LogEntry, error) {
						return nil, errors.New("something went wrong")
					},
				},
			},
			assertions: func(_ <-chan LogEntry, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			logCh, err := testCase.service.Stream(
				context.Background(),
				testEventID,
				testCase.selector,
				LogStreamOptions{},
			)
			testCase.assertions(logCh, err)
		})
	}
}
