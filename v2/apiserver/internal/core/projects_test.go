package core

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	metaTesting "github.com/brigadecore/brigade/v2/apiserver/internal/meta/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestProjectMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, &Project{}, "Project")
}

func TestProjectListMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, &ProjectList{}, "ProjectList")
}

func TestNewProjectsService(t *testing.T) {
	projectsStore := &mockProjectsStore{}
	eventsStore := &mockEventsStore{}
	substrate := &mockSubstrate{}
	svc := NewProjectsService(projectsStore, eventsStore, substrate)
	require.Same(t, projectsStore, svc.(*projectsService).projectsStore)
	require.Same(t, substrate, svc.(*projectsService).substrate)
}

func TestProjectServiceCreate(t *testing.T) {
	testCases := []struct {
		name       string
		service    ProjectsService
		assertions func(error)
	}{
		{
			name: "error creating project in substrate",
			service: &projectsService{
				substrate: &mockSubstrate{
					CreateProjectFn: func(
						_ context.Context,
						project Project,
					) (Project, error) {
						return project, errors.New("substrate error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "substrate error")
				require.Contains(t, err.Error(), "on the substrate")
			},
		},
		{
			name: "error creating project in store",
			service: &projectsService{
				substrate: &mockSubstrate{
					CreateProjectFn: func(
						_ context.Context,
						project Project,
					) (Project, error) {
						return project, nil
					},
				},
				projectsStore: &mockProjectsStore{
					CreateFn: func(context.Context, Project) error {
						return errors.New("store error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "store error")
				require.Contains(t, err.Error(), "error storing new project")
			},
		},
		{
			name: "success",
			service: &projectsService{
				substrate: &mockSubstrate{
					CreateProjectFn: func(
						ctx context.Context,
						project Project,
					) (Project, error) {
						return project, nil
					},
				},
				projectsStore: &mockProjectsStore{
					CreateFn: func(context.Context, Project) error {
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
			_, err := testCase.service.Create(context.Background(), Project{})
			testCase.assertions(err)
		})
	}
}

func TestProjectServiceList(t *testing.T) {
	testCases := []struct {
		name       string
		service    ProjectsService
		assertions func(error)
	}{
		{
			name: "error getting projects from store",
			service: &projectsService{
				projectsStore: &mockProjectsStore{
					ListFn: func(context.Context, meta.ListOptions) (ProjectList, error) {
						return ProjectList{}, errors.New("error listing projects")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error listing projects")
				require.Contains(t, err.Error(), "error retrieving projects from store")
			},
		},
		{
			name: "success",
			service: &projectsService{
				projectsStore: &mockProjectsStore{
					ListFn: func(context.Context, meta.ListOptions) (ProjectList, error) {
						return ProjectList{}, nil
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
				testCase.service.List(context.Background(), meta.ListOptions{})
			testCase.assertions(err)
		})
	}
}

func TestProjectServiceGet(t *testing.T) {
	testCases := []struct {
		name       string
		service    ProjectsService
		assertions func(error)
	}{
		{
			name: "error getting projects from store",
			service: &projectsService{
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, errors.New("error getting project")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error getting project")
				require.Contains(t, err.Error(), "error retrieving project")
			},
		},
		{
			name: "success",
			service: &projectsService{
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
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
				testCase.service.Get(context.Background(), "foo")
			testCase.assertions(err)
		})
	}
}

func TestProjectServiceUpdate(t *testing.T) {
	testCases := []struct {
		name       string
		service    ProjectsService
		assertions func(error)
	}{
		{
			name: "error updating project in store",
			service: &projectsService{
				projectsStore: &mockProjectsStore{
					UpdateFn: func(context.Context, Project) error {
						return errors.New("store error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "store error")
				require.Contains(t, err.Error(), "error updating project")
			},
		},
		{
			name: "success",
			service: &projectsService{
				projectsStore: &mockProjectsStore{
					UpdateFn: func(context.Context, Project) error {
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
			err := testCase.service.Update(context.Background(), Project{})
			testCase.assertions(err)
		})
	}
}

func TestProjectServiceDelete(t *testing.T) {
	testCases := []struct {
		name       string
		service    ProjectsService
		assertions func(error)
	}{
		{
			name: "error retrieving project from store",
			service: &projectsService{
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, errors.New("store error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "store error")
				require.Contains(t, err.Error(), "error retrieving project")
			},
		},
		{
			name: "error listing events associated with project",
			service: &projectsService{
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return nil
					},
				},
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
			},
		},
		{
			name: "error deleting events associated with project",
			service: &projectsService{
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return nil
					},
				},
				eventsStore: &mockEventsStore{
					ListFn: func(
						context.Context,
						EventsSelector,
						meta.ListOptions,
					) (EventList, error) {
						return EventList{Items: []Event{{ProjectID: "foo"}}}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return errors.New("error deleting events")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error deleting events")
			},
		},
		{
			name: "error deleting worker/jobs associated with the project/event",
			service: &projectsService{
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return nil
					},
				},
				eventsStore: &mockEventsStore{
					ListFn: func(
						context.Context,
						EventsSelector,
						meta.ListOptions,
					) (EventList, error) {
						return EventList{Items: []Event{{ProjectID: "foo"}}}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return nil
					},
				},
				substrate: &mockSubstrate{
					DeleteWorkerAndJobsFn: func(context.Context, Project, Event) error {
						return errors.New("error deleting worker and jobs")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error deleting worker and jobs")
			},
		},
		{
			name: "error deleting project from store",
			service: &projectsService{
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return errors.New("store error")
					},
				},
				eventsStore: &mockEventsStore{
					ListFn: func(
						context.Context,
						EventsSelector,
						meta.ListOptions,
					) (EventList, error) {
						return EventList{}, nil
					},
					DeleteFn: func(context.Context, string) error {
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
				require.Error(t, err)
				require.Contains(t, err.Error(), "store error")
				require.Contains(t, err.Error(), "error removing project")
			},
		},
		{
			name: "error deleting project from substrate",
			service: &projectsService{
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return nil
					},
				},
				eventsStore: &mockEventsStore{
					ListFn: func(
						context.Context,
						EventsSelector,
						meta.ListOptions,
					) (EventList, error) {
						return EventList{}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return nil
					},
				},
				substrate: &mockSubstrate{
					DeleteWorkerAndJobsFn: func(context.Context, Project, Event) error {
						return nil
					},
					DeleteProjectFn: func(context.Context, Project) error {
						return errors.New("substrate error")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "substrate error")
				require.Contains(t, err.Error(), "error deleting project")
			},
		},
		{
			name: "success",
			service: &projectsService{
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return nil
					},
				},
				eventsStore: &mockEventsStore{
					ListFn: func(
						context.Context,
						EventsSelector,
						meta.ListOptions,
					) (EventList, error) {
						return EventList{}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return nil
					},
				},
				substrate: &mockSubstrate{
					DeleteWorkerAndJobsFn: func(context.Context, Project, Event) error {
						return nil
					},
					DeleteProjectFn: func(context.Context, Project) error {
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
			err := testCase.service.Delete(context.Background(), "foo")
			testCase.assertions(err)
		})
	}
}
