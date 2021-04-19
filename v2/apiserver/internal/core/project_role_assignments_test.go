package core

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	"github.com/brigadecore/brigade/v2/apiserver/internal/authz"
	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
)

func TestProjectRoleAssignmentMatches(t *testing.T) {
	testCases := []struct {
		name                  string
		projectRoleAssignment ProjectRoleAssignment
		role                  libAuthz.Role
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

func TestNewProjectRoleAssignmentsService(t *testing.T) {
	projectsStore := &mockProjectsStore{}
	usersStore := &authn.MockUsersStore{}
	serviceAccountsStore := &authn.MockServiceAccountStore{}
	projectRoleAssignmentsStore := &mockProjectRoleAssignmentsStore{}
	svc := NewProjectRoleAssignmentsService(
		alwaysProjectAuthorize,
		projectsStore,
		usersStore,
		serviceAccountsStore,
		projectRoleAssignmentsStore,
	)
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
				Principal: libAuthz.PrincipalReference{
					Type: authz.PrincipalTypeUser,
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
				Principal: libAuthz.PrincipalReference{
					Type: authz.PrincipalTypeUser,
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
				usersStore: &authn.MockUsersStore{
					GetFn: func(context.Context, string) (authn.User, error) {
						return authn.User{}, errors.New("something went wrong")
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
				Principal: libAuthz.PrincipalReference{
					Type: authz.PrincipalTypeServiceAccount,
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
				serviceAccountsStore: &authn.MockServiceAccountStore{
					GetFn: func(context.Context, string) (authn.ServiceAccount, error) {
						return authn.ServiceAccount{}, errors.New("something went wrong")
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
				Principal: libAuthz.PrincipalReference{
					Type: authz.PrincipalTypeServiceAccount,
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
				serviceAccountsStore: &authn.MockServiceAccountStore{
					GetFn: func(context.Context, string) (authn.ServiceAccount, error) {
						return authn.ServiceAccount{}, nil
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
				Principal: libAuthz.PrincipalReference{
					Type: authz.PrincipalTypeServiceAccount,
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
				serviceAccountsStore: &authn.MockServiceAccountStore{
					GetFn: func(context.Context, string) (authn.ServiceAccount, error) {
						return authn.ServiceAccount{}, nil
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
				Principal: libAuthz.PrincipalReference{
					Type: authz.PrincipalTypeUser,
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
				Principal: libAuthz.PrincipalReference{
					Type: authz.PrincipalTypeUser,
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
				usersStore: &authn.MockUsersStore{
					GetFn: func(context.Context, string) (authn.User, error) {
						return authn.User{}, errors.New("something went wrong")
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
				Principal: libAuthz.PrincipalReference{
					Type: authz.PrincipalTypeServiceAccount,
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
				serviceAccountsStore: &authn.MockServiceAccountStore{
					GetFn: func(context.Context, string) (authn.ServiceAccount, error) {
						return authn.ServiceAccount{}, errors.New("something went wrong")
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
				Principal: libAuthz.PrincipalReference{
					Type: authz.PrincipalTypeServiceAccount,
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
				serviceAccountsStore: &authn.MockServiceAccountStore{
					GetFn: func(context.Context, string) (authn.ServiceAccount, error) {
						return authn.ServiceAccount{}, nil
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
				Principal: libAuthz.PrincipalReference{
					Type: authz.PrincipalTypeServiceAccount,
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
				serviceAccountsStore: &authn.MockServiceAccountStore{
					GetFn: func(context.Context, string) (authn.ServiceAccount, error) {
						return authn.ServiceAccount{}, nil
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
	GrantFn      func(context.Context, ProjectRoleAssignment) error
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
