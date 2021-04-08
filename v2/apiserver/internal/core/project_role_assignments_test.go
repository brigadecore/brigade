package core

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
)

func TestNewProjectRoleAssignmentsService(t *testing.T) {
	projectsStore := &mockProjectsStore{}
	usersStore := &authn.MockUsersStore{}
	serviceAccountsStore := &authn.MockServiceAccountStore{}
	projectRoleAssignmentsStore := &mockProjectRoleAssignmentsStore{}
	svc := NewProjectRoleAssignmentsService(
		alwaysAuthorize,
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
				projectAuthorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error retrieving project from store",
			projectRoleAssignment: ProjectRoleAssignment{
				Principal: authn.PrincipalReference{
					Type: authn.PrincipalTypeUser,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysAuthorize,
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
				Principal: authn.PrincipalReference{
					Type: authn.PrincipalTypeUser,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysAuthorize,
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
				Principal: authn.PrincipalReference{
					Type: authn.PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysAuthorize,
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
				Principal: authn.PrincipalReference{
					Type: authn.PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysAuthorize,
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
				Principal: authn.PrincipalReference{
					Type: authn.PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysAuthorize,
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
				projectAuthorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error retrieving project from store",
			projectRoleAssignment: ProjectRoleAssignment{
				Principal: authn.PrincipalReference{
					Type: authn.PrincipalTypeUser,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysAuthorize,
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
				Principal: authn.PrincipalReference{
					Type: authn.PrincipalTypeUser,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysAuthorize,
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
				Principal: authn.PrincipalReference{
					Type: authn.PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysAuthorize,
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
				Principal: authn.PrincipalReference{
					Type: authn.PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysAuthorize,
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
				Principal: authn.PrincipalReference{
					Type: authn.PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &projectRoleAssignmentsService{
				projectAuthorize: alwaysAuthorize,
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
