package core

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	"github.com/brigadecore/brigade/v2/apiserver/internal/authz"
	"github.com/stretchr/testify/require"
)

func TestNewProjectRoleAssignmentsService(t *testing.T) {
	projectsStore := &mockProjectsStore{}
	usersStore := &authn.MockUsersStore{}
	serviceAccountsStore := &authn.MockServiceAccountStore{}
	roleAssignmentsStore := &authz.MockRoleAssignmentsStore{}
	svc := NewProjectRoleAssignmentsService(
		projectsStore,
		usersStore,
		serviceAccountsStore,
		roleAssignmentsStore,
	)
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
		roleAssignmentsStore,
		svc.(*projectRoleAssignmentsService).roleAssignmentsStore,
	)
}

func TestRoleAssignmentsServiceGrant(t *testing.T) {
	testCases := []struct {
		name           string
		roleAssignment authz.RoleAssignment
		service        authz.RoleAssignmentsService
		assertions     func(error)
	}{
		{
			name: "error retrieving project from store",
			roleAssignment: authz.RoleAssignment{
				Principal: authz.PrincipalReference{
					Type: authz.PrincipalTypeUser,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
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
			roleAssignment: authz.RoleAssignment{
				Principal: authz.PrincipalReference{
					Type: authz.PrincipalTypeUser,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
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
			roleAssignment: authz.RoleAssignment{
				Principal: authz.PrincipalReference{
					Type: authz.PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
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
			roleAssignment: authz.RoleAssignment{
				Principal: authz.PrincipalReference{
					Type: authz.PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
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
				roleAssignmentsStore: &authz.MockRoleAssignmentsStore{
					GrantFn: func(context.Context, authz.RoleAssignment) error {
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
			roleAssignment: authz.RoleAssignment{
				Principal: authz.PrincipalReference{
					Type: authz.PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
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
				roleAssignmentsStore: &authz.MockRoleAssignmentsStore{
					GrantFn: func(context.Context, authz.RoleAssignment) error {
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
				testCase.roleAssignment,
			)
			testCase.assertions(err)
		})
	}
}

func TestRoleAssignmentsServiceRevoke(t *testing.T) {
	testCases := []struct {
		name           string
		roleAssignment authz.RoleAssignment
		service        authz.RoleAssignmentsService
		assertions     func(error)
	}{
		{
			name: "error retrieving project from store",
			roleAssignment: authz.RoleAssignment{
				Principal: authz.PrincipalReference{
					Type: authz.PrincipalTypeUser,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
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
			roleAssignment: authz.RoleAssignment{
				Principal: authz.PrincipalReference{
					Type: authz.PrincipalTypeUser,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
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
			roleAssignment: authz.RoleAssignment{
				Principal: authz.PrincipalReference{
					Type: authz.PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
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
			roleAssignment: authz.RoleAssignment{
				Principal: authz.PrincipalReference{
					Type: authz.PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
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
				roleAssignmentsStore: &authz.MockRoleAssignmentsStore{
					RevokeFn: func(context.Context, authz.RoleAssignment) error {
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
			roleAssignment: authz.RoleAssignment{
				Principal: authz.PrincipalReference{
					Type: authz.PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
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
				roleAssignmentsStore: &authz.MockRoleAssignmentsStore{
					RevokeFn: func(context.Context, authz.RoleAssignment) error {
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
				testCase.roleAssignment,
			)
			testCase.assertions(err)
		})
	}
}
