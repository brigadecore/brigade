package core

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
)

func TestNewjobsService(t *testing.T) {
	projectsStore := &mockProjectsStore{}
	eventsStore := &mockEventsStore{}
	substrate := &mockSubstrate{}
	svc := NewJobsService(projectsStore, eventsStore, substrate)
	require.Same(t, projectsStore, svc.(*jobsService).projectsStore)
	require.Same(t, eventsStore, svc.(*jobsService).eventsStore)
	require.Same(t, substrate, svc.(*jobsService).substrate)
}

func TestJobsServiceStart(t *testing.T) {
	const testEventID = "123456789"
	const testJobName = "foo"
	testCases := []struct {
		name       string
		service    JobsService
		assertions func(error)
	}{
		{
			name: "error getting event from store",
			service: &jobsService{
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
			name: "event has no such job",
			service: &jobsService{
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
			},
		},
		{
			name: "job is not currently pending",
			service: &jobsService{
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: map[string]Job{
									testJobName: {
										Status: &JobStatus{
											Phase: JobPhaseRunning,
										},
									},
								},
							},
						}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrConflict{}, err)
			},
		},
		{
			name: "error getting project from store",
			service: &jobsService{
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: map[string]Job{
									testJobName: {
										Status: &JobStatus{
											Phase: JobPhasePending,
										},
									},
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
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving project")
			},
		},
		{
			name: "error starting job on substrate",
			service: &jobsService{
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: map[string]Job{
									testJobName: {
										Status: &JobStatus{
											Phase: JobPhasePending,
										},
									},
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
					StartJobFn: func(context.Context, Project, Event, string) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error starting event")
			},
		},
		{
			name: "success",
			service: &jobsService{
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: map[string]Job{
									testJobName: {
										Status: &JobStatus{
											Phase: JobPhasePending,
										},
									},
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
					StartJobFn: func(context.Context, Project, Event, string) error {
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
			err :=
				testCase.service.Start(context.Background(), testEventID, testJobName)
			testCase.assertions(err)
		})
	}
}
