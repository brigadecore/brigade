package api

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	metaTesting "github.com/brigadecore/brigade/v2/apiserver/internal/meta/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestProjectRoleAssignmentMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(
		t,
		&ProjectRoleAssignment{},
		ProjectRoleAssignmentKind,
	)
}

func TestProjectRoleAssignmentMatches(t *testing.T) {
	testCases := []struct {
		name                  string
		projectRoleAssignment ProjectRoleAssignment
		role                  Role
		projectID             string
		matches               bool
	}{
		{
			name: "names do not match",
			projectRoleAssignment: ProjectRoleAssignment{
				Role:      "foo",
				ProjectID: "foo",
			},
			role:      "bar",
			projectID: "foo",
			matches:   false,
		},
		{
			name: "scopes do not match",
			projectRoleAssignment: ProjectRoleAssignment{
				Role:      "foo",
				ProjectID: "foo",
			},
			role:      "foo",
			projectID: "bar",
			matches:   false,
		},
		{
			name: "scopes are an exact match",
			projectRoleAssignment: ProjectRoleAssignment{
				Role:      "foo",
				ProjectID: "foo",
			},
			role:      "foo",
			projectID: "foo",
			matches:   true,
		},
		{
			name: "a global project scope matches b project",
			projectRoleAssignment: ProjectRoleAssignment{
				Role:      "foo",
				ProjectID: ProjectRoleScopeGlobal,
			},
			role:      "foo",
			projectID: "foo",
			matches:   true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(
				t,
				testCase.matches,
				testCase.projectRoleAssignment.Matches(
					testCase.projectID,
					testCase.role,
				),
			)
		})
	}
}

func TestProjectRoleAssignmentListMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(
		t,
		&ProjectRoleAssignmentList{},
		ProjectRoleAssignmentListKind,
	)
}

func TestNewProjectRoleAssignmentsService(t *testing.T) {
	projectsStore := &mockProjectsStore{}
	usersStore := &mockUsersStore{}
	serviceAccountsStore := &mockServiceAccountStore{}
	projectRoleAssignmentsStore := &mockProjectRoleAssignmentsStore{}
	svc := NewProjectRoleAssignmentsService(
		alwaysAuthorize,
		alwaysProjectAuthorize,
		projectsStore,
		usersStore,
		serviceAccountsStore,
		projectRoleAssignmentsStore,
	)
	require.NotNil(t, svc.(*projectRoleAssignmentsService).authorize)
	require.NotNil(t, svc.(*projectRoleAssignmentsService).projectAuthorize)
	require.Same(
		t,
		projectsStore,
		svc.(*projectRoleAssignmentsService).projectsStore,
	)
	require.Same(t, usersStore, svc.(*projectRoleAssignmentsService).usersStore)
	require.Same(
		t,
		serviceAccountsStore,
		svc.(*projectRoleAssignmentsService).serviceAccountsStore,
	)
	require.Same(
		t,
		projectRoleAssignmentsStore,
		svc.(*projectRoleAssignmentsService).projectRoleAssignmentsStore,
	)
}

func TestProjectRoleAssignmentsServiceGrant(t *testing.T) {
	testCases := []struct {
		name                  string
		projectRoleAssignment ProjectRoleAssignment
		service               ProjectRoleAssignmentsService
		assertions            func(error)
	}{
		{
			name: "unauthorized",
			service: &projectRoleAssignmentsService{
				projectAuthorize: neverProjectAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error retrieving project from store",
			projectRoleAssignment: ProjectRoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeUser,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysProjectAuthorize,
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
			name: "error retrieving user from store",
			projectRoleAssignment: ProjectRoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeUser,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				usersStore: &mockUsersStore{
					GetFn: func(context.Context, string) (User, error) {
						return User{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving user")
			},
		},
		{
			name: "error retrieving service account from store",
			projectRoleAssignment: ProjectRoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				serviceAccountsStore: &mockServiceAccountStore{
					GetFn: func(context.Context, string) (ServiceAccount, error) {
						return ServiceAccount{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving service account")
			},
		},
		{
			name: "error granting the role",
			projectRoleAssignment: ProjectRoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				serviceAccountsStore: &mockServiceAccountStore{
					GetFn: func(context.Context, string) (ServiceAccount, error) {
						return ServiceAccount{}, nil
					},
				},
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					GrantFn: func(context.Context, ProjectRoleAssignment) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error granting project")
			},
		},
		{
			name: "success",
			projectRoleAssignment: ProjectRoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				serviceAccountsStore: &mockServiceAccountStore{
					GetFn: func(context.Context, string) (ServiceAccount, error) {
						return ServiceAccount{}, nil
					},
				},
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					GrantFn: func(context.Context, ProjectRoleAssignment) error {
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
			err := testCase.service.Grant(
				context.Background(),
				testCase.projectRoleAssignment,
			)
			testCase.assertions(err)
		})
	}
}

func TestProjectRoleAssignmentsServiceList(t *testing.T) {
	testCases := []struct {
		name       string
		service    ProjectRoleAssignmentsService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &projectRoleAssignmentsService{
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting project role assignments from store",
			service: &projectRoleAssignmentsService{
				authorize: alwaysAuthorize,
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					ListFn: func(
						context.Context,
						ProjectRoleAssignmentsSelector,
						meta.ListOptions,
					) (ProjectRoleAssignmentList, error) {
						return ProjectRoleAssignmentList{},
							errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(
					t,
					err.Error(),
					"error retrieving project role assignments from store",
				)
			},
		},
		{
			name: "success",
			service: &projectRoleAssignmentsService{
				authorize: alwaysAuthorize,
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					ListFn: func(
						context.Context,
						ProjectRoleAssignmentsSelector,
						meta.ListOptions,
					) (ProjectRoleAssignmentList, error) {
						return ProjectRoleAssignmentList{}, nil
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
		})
	}
}

func TestProjectRoleAssignmentsServiceRevoke(t *testing.T) {
	testCases := []struct {
		name                  string
		projectRoleAssignment ProjectRoleAssignment
		service               ProjectRoleAssignmentsService
		assertions            func(error)
	}{
		{
			name: "unauthorized",
			service: &projectRoleAssignmentsService{
				projectAuthorize: neverProjectAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error retrieving project from store",
			projectRoleAssignment: ProjectRoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeUser,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysProjectAuthorize,
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
			name: "error retrieving user from store",
			projectRoleAssignment: ProjectRoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeUser,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				usersStore: &mockUsersStore{
					GetFn: func(context.Context, string) (User, error) {
						return User{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving user")
			},
		},
		{
			name: "error retrieving service account from store",
			projectRoleAssignment: ProjectRoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				serviceAccountsStore: &mockServiceAccountStore{
					GetFn: func(context.Context, string) (ServiceAccount, error) {
						return ServiceAccount{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving service account")
			},
		},
		{
			name: "error revoking the role",
			projectRoleAssignment: ProjectRoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				serviceAccountsStore: &mockServiceAccountStore{
					GetFn: func(context.Context, string) (ServiceAccount, error) {
						return ServiceAccount{}, nil
					},
				},
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					RevokeFn: func(context.Context, ProjectRoleAssignment) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error revoking project")
			},
		},
		{
			name: "success",
			projectRoleAssignment: ProjectRoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				serviceAccountsStore: &mockServiceAccountStore{
					GetFn: func(context.Context, string) (ServiceAccount, error) {
						return ServiceAccount{}, nil
					},
				},
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					RevokeFn: func(context.Context, ProjectRoleAssignment) error {
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
			err := testCase.service.Revoke(
				context.Background(),
				testCase.projectRoleAssignment,
			)
			testCase.assertions(err)
		})
	}
}

type mockProjectRoleAssignmentsStore struct {
	GrantFn func(context.Context, ProjectRoleAssignment) error
	ListFn  func(
		context.Context,
		ProjectRoleAssignmentsSelector,
		meta.ListOptions,
	) (ProjectRoleAssignmentList, error)
	RevokeFn     func(context.Context, ProjectRoleAssignment) error
	RevokeManyFn func(ctx context.Context, projectID string) error
	ExistsFn     func(context.Context, ProjectRoleAssignment) (bool, error)
}

func (m *mockProjectRoleAssignmentsStore) Grant(
	ctx context.Context,
	projectRoleAssignment ProjectRoleAssignment,
) error {
	return m.GrantFn(ctx, projectRoleAssignment)
}

func (m *mockProjectRoleAssignmentsStore) List(
	ctx context.Context,
	selector ProjectRoleAssignmentsSelector,
	opts meta.ListOptions,
) (ProjectRoleAssignmentList, error) {
	return m.ListFn(ctx, selector, opts)
}

func (m *mockProjectRoleAssignmentsStore) Revoke(
	ctx context.Context,
	projectRoleAssignment ProjectRoleAssignment,
) error {
	return m.RevokeFn(ctx, projectRoleAssignment)
}

func (m *mockProjectRoleAssignmentsStore) RevokeByProjectID(
	ctx context.Context,
	projectID string,
) error {
	return m.RevokeManyFn(ctx, projectID)
}

func (m *mockProjectRoleAssignmentsStore) Exists(
	ctx context.Context,
	projectRoleAssignment ProjectRoleAssignment,
) (bool, error) {
	return m.ExistsFn(ctx, projectRoleAssignment)
}
