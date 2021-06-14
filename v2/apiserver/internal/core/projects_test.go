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

func TestProjectMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, &Project{}, ProjectKind)
}

func TestProjectListMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, &ProjectList{}, "ProjectList")
}

func TestProjectListContains(t *testing.T) {
	testProject := Project{
		ObjectMeta: meta.ObjectMeta{
			ID: "myproject",
		},
		Kubernetes: &KubernetesDetails{
			Namespace: "mynamespace",
		},
	}

	t.Run("empty list", func(t *testing.T) {
		projectList := ProjectList{}
		require.False(t, projectList.Contains(testProject))
	})

	t.Run("project not in list", func(t *testing.T) {
		projectList := ProjectList{
			Items: []Project{
				{
					ObjectMeta: meta.ObjectMeta{
						ID: "diffproject",
					},
				},
			},
		}
		require.False(t, projectList.Contains(testProject))
	})

	t.Run("name matches, namespace does not", func(t *testing.T) {
		projectList := ProjectList{
			Items: []Project{
				{
					ObjectMeta: meta.ObjectMeta{
						ID: "myproject",
					},
				},
			},
		}
		require.False(t, projectList.Contains(testProject))
	})

	t.Run("project matches", func(t *testing.T) {
		projectList := ProjectList{
			Items: []Project{
				{
					ObjectMeta: meta.ObjectMeta{
						ID: "myproject",
					},
					Kubernetes: &KubernetesDetails{
						Namespace: "mynamespace",
					},
				},
			},
		}
		require.True(t, projectList.Contains(testProject))
	})
}

func TestNewProjectsService(t *testing.T) {
	projectsStore := &mockProjectsStore{}
	eventsStore := &mockEventsStore{}
	logsStore := &mockLogsStore{}
	projectRoleAssignmentsStore := &mockProjectRoleAssignmentsStore{}
	substrate := &mockSubstrate{}
	svc := NewProjectsService(
		libAuthz.AlwaysAuthorize,
		alwaysProjectAuthorize,
		projectsStore,
		eventsStore,
		logsStore,
		projectRoleAssignmentsStore,
		substrate,
	)
	require.NotNil(t, svc.(*projectsService).authorize)
	require.NotNil(t, svc.(*projectsService).projectAuthorize)
	require.Same(t, projectsStore, svc.(*projectsService).projectsStore)
	require.Same(t, eventsStore, svc.(*projectsService).eventsStore)
	require.Same(
		t,
		projectRoleAssignmentsStore,
		svc.(*projectsService).projectRoleAssignmentsStore,
	)
	require.Same(t, substrate, svc.(*projectsService).substrate)
}

func TestProjectServiceCreate(t *testing.T) {
	testCases := []struct {
		name       string
		service    ProjectsService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &projectsService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error creating project in substrate",
			service: &projectsService{
				authorize: libAuthz.AlwaysAuthorize,
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
				authorize: libAuthz.AlwaysAuthorize,
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
				authorize: libAuthz.AlwaysAuthorize,
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
			name: "unauthorized",
			service: &projectsService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting projects from store",
			service: &projectsService{
				authorize: libAuthz.AlwaysAuthorize,
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
				authorize: libAuthz.AlwaysAuthorize,
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
			name: "unauthorized",
			service: &projectsService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting projects from store",
			service: &projectsService{
				authorize: libAuthz.AlwaysAuthorize,
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
				authorize: libAuthz.AlwaysAuthorize,
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
			name: "unauthorized",
			service: &projectsService{
				projectAuthorize: neverProjectAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error updating project in store",
			service: &projectsService{
				projectAuthorize: alwaysProjectAuthorize,
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
				projectAuthorize: alwaysProjectAuthorize,
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
			name: "unauthorized",
			service: &projectsService{
				projectAuthorize: neverProjectAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error retrieving project from store",
			service: &projectsService{
				projectAuthorize: alwaysProjectAuthorize,
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
			name: "error deleting events associated with project",
			service: &projectsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return nil
					},
				},
				eventsStore: &mockEventsStore{
					DeleteManyFn: func(
						context.Context,
						EventsSelector,
					) (<-chan Event, int64, error) {
						return nil,
							0,
							errors.New("error deleting events associated with project")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"error deleting events associated with project",
				)
			},
		},
		{
			name: "error deleting project logs",
			service: &projectsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return nil
					},
				},
				eventsStore: &mockEventsStore{
					DeleteManyFn: func(
						context.Context,
						EventsSelector,
					) (<-chan Event, int64, error) {
						return nil, 0, nil
					},
				},
				logsStore: &mockLogsStore{
					DeleteProjectLogsFn: func(
						context.Context,
						string,
					) error {
						return errors.New("error deleting project logs")
					},
				},
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					RevokeManyFn: func(context.Context, string) error {
						return nil
					},
				},
				substrate: &mockSubstrate{
					DeleteProjectFn: func(context.Context, Project) error {
						return nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error deleting project logs")
			},
		},
		{
			name: "error deleting role assignments associated with project",
			service: &projectsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return nil
					},
				},
				eventsStore: &mockEventsStore{
					DeleteManyFn: func(
						context.Context,
						EventsSelector,
					) (<-chan Event, int64, error) {
						return nil, 0, nil
					},
				},
				logsStore: &mockLogsStore{
					DeleteProjectLogsFn: func(
						context.Context,
						string,
					) error {
						return nil
					},
				},
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					RevokeManyFn: func(context.Context, string) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(
					t,
					err.Error(),
					"error revoking all role assignments associated with project",
				)
			},
		},
		{
			name: "error deleting project from store",
			service: &projectsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return errors.New("store error")
					},
				},
				eventsStore: &mockEventsStore{
					DeleteManyFn: func(
						context.Context,
						EventsSelector,
					) (<-chan Event, int64, error) {
						return nil, 0, nil
					},
				},
				logsStore: &mockLogsStore{
					DeleteProjectLogsFn: func(
						context.Context,
						string,
					) error {
						return nil
					},
				},
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					RevokeManyFn: func(context.Context, string) error {
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
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return nil
					},
				},
				eventsStore: &mockEventsStore{
					DeleteManyFn: func(
						context.Context,
						EventsSelector,
					) (<-chan Event, int64, error) {
						return nil, 0, nil
					},
				},
				logsStore: &mockLogsStore{
					DeleteProjectLogsFn: func(
						context.Context,
						string,
					) error {
						return nil
					},
				},
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					RevokeManyFn: func(context.Context, string) error {
						return nil
					},
				},
				substrate: &mockSubstrate{
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
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
					DeleteFn: func(context.Context, string) error {
						return nil
					},
				},
				eventsStore: &mockEventsStore{
					DeleteManyFn: func(
						context.Context,
						EventsSelector,
					) (<-chan Event, int64, error) {
						return nil, 0, nil
					},
				},
				logsStore: &mockLogsStore{
					DeleteProjectLogsFn: func(
						context.Context,
						string,
					) error {
						return nil
					},
				},
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					RevokeManyFn: func(context.Context, string) error {
						return nil
					},
				},
				substrate: &mockSubstrate{
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
