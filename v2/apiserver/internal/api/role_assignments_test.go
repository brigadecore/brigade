package api

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	metaTesting "github.com/brigadecore/brigade/v2/apiserver/internal/meta/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestRoleAssignmentMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, &RoleAssignment{}, RoleAssignmentKind)
}

func TestMatches(t *testing.T) {
	testCases := []struct {
		name           string
		roleAssignment RoleAssignment
		role           Role
		scope          string
		matches        bool
	}{
		{
			name: "names do not match",
			roleAssignment: RoleAssignment{
				Role:  "foo",
				Scope: "foo",
			},
			role:    "bar",
			scope:   "foo",
			matches: false,
		},
		{
			name: "scopes do not match",
			roleAssignment: RoleAssignment{
				Role:  "foo",
				Scope: "foo",
			},
			role:    "foo",
			scope:   "bar",
			matches: false,
		},
		{
			name: "scopes are an exact match",
			roleAssignment: RoleAssignment{
				Role:  "foo",
				Scope: "foo",
			},
			role:    "foo",
			scope:   "foo",
			matches: true,
		},
		{
			name: "a global scope matches b scope",
			roleAssignment: RoleAssignment{
				Role:  "foo",
				Scope: RoleScopeGlobal,
			},
			role:    "foo",
			scope:   "foo",
			matches: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(
				t,
				testCase.matches,
				testCase.roleAssignment.Matches(testCase.role, testCase.scope),
			)
		})
	}
}

func TestRoleAssignmentListMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(
		t,
		&RoleAssignmentList{},
		RoleAssignmentListKind,
	)
}

func TestNewRoleAssignmentsService(t *testing.T) {
	usersStore := &mockUsersStore{}
	serviceAccountsStore := &mockServiceAccountStore{}
	roleAssignmentsStore := &mockRoleAssignmentsStore{}
	svc, ok := NewRoleAssignmentsService(
		alwaysAuthorize,
		usersStore,
		serviceAccountsStore,
		roleAssignmentsStore,
	).(*roleAssignmentsService)
	require.True(t, ok)
	require.NotNil(t, svc.authorize)
	require.Same(t, usersStore, svc.usersStore)
	require.Same(t, serviceAccountsStore, svc.serviceAccountsStore)
	require.Same(t, roleAssignmentsStore, svc.roleAssignmentsStore)
}

func TestRoleAssignmentsServiceGrant(t *testing.T) {
	testCases := []struct {
		name           string
		roleAssignment RoleAssignment
		service        RoleAssignmentsService
		assertions     func(error)
	}{
		{
			name: "unauthorized",
			service: &roleAssignmentsService{
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error retrieving user from store",
			roleAssignment: RoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeUser,
					ID:   "foo",
				},
			},
			service: &roleAssignmentsService{
				authorize: alwaysAuthorize,
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
			roleAssignment: RoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &roleAssignmentsService{
				authorize: alwaysAuthorize,
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
			roleAssignment: RoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &roleAssignmentsService{
				authorize: alwaysAuthorize,
				serviceAccountsStore: &mockServiceAccountStore{
					GetFn: func(context.Context, string) (ServiceAccount, error) {
						return ServiceAccount{}, nil
					},
				},
				roleAssignmentsStore: &mockRoleAssignmentsStore{
					GrantFn: func(context.Context, RoleAssignment) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error granting role")
			},
		},
		{
			name: "success",
			roleAssignment: RoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &roleAssignmentsService{
				authorize: alwaysAuthorize,
				serviceAccountsStore: &mockServiceAccountStore{
					GetFn: func(context.Context, string) (ServiceAccount, error) {
						return ServiceAccount{}, nil
					},
				},
				roleAssignmentsStore: &mockRoleAssignmentsStore{
					GrantFn: func(context.Context, RoleAssignment) error {
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

func TestRoleAssignmentsServiceList(t *testing.T) {
	testCases := []struct {
		name       string
		service    RoleAssignmentsService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &roleAssignmentsService{
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting role assignments from store",
			service: &roleAssignmentsService{
				authorize: alwaysAuthorize,
				roleAssignmentsStore: &mockRoleAssignmentsStore{
					ListFn: func(
						context.Context,
						RoleAssignmentsSelector,
						meta.ListOptions,
					) (RoleAssignmentList, error) {
						return RoleAssignmentList{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(
					t,
					err.Error(),
					"error retrieving role assignments from store",
				)
			},
		},
		{
			name: "success",
			service: &roleAssignmentsService{
				authorize: alwaysAuthorize,
				roleAssignmentsStore: &mockRoleAssignmentsStore{
					ListFn: func(
						context.Context,
						RoleAssignmentsSelector,
						meta.ListOptions,
					) (RoleAssignmentList, error) {
						return RoleAssignmentList{}, nil
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

func TestRoleAssignmentsServiceRevoke(t *testing.T) {
	testCases := []struct {
		name           string
		roleAssignment RoleAssignment
		service        RoleAssignmentsService
		assertions     func(error)
	}{
		{
			name: "unauthorized",
			service: &roleAssignmentsService{
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error retrieving user from store",
			roleAssignment: RoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeUser,
					ID:   "foo",
				},
			},
			service: &roleAssignmentsService{
				authorize: alwaysAuthorize,
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
			roleAssignment: RoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &roleAssignmentsService{
				authorize: alwaysAuthorize,
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
			roleAssignment: RoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &roleAssignmentsService{
				authorize: alwaysAuthorize,
				serviceAccountsStore: &mockServiceAccountStore{
					GetFn: func(context.Context, string) (ServiceAccount, error) {
						return ServiceAccount{}, nil
					},
				},
				roleAssignmentsStore: &mockRoleAssignmentsStore{
					RevokeFn: func(context.Context, RoleAssignment) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error revoking role")
			},
		},
		{
			name: "success",
			roleAssignment: RoleAssignment{
				Principal: PrincipalReference{
					Type: PrincipalTypeServiceAccount,
					ID:   "foo",
				},
			},
			service: &roleAssignmentsService{
				authorize: alwaysAuthorize,
				serviceAccountsStore: &mockServiceAccountStore{
					GetFn: func(context.Context, string) (ServiceAccount, error) {
						return ServiceAccount{}, nil
					},
				},
				roleAssignmentsStore: &mockRoleAssignmentsStore{
					RevokeFn: func(context.Context, RoleAssignment) error {
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

type mockRoleAssignmentsStore struct {
	GrantFn func(context.Context, RoleAssignment) error
	ListFn  func(
		context.Context,
		RoleAssignmentsSelector,
		meta.ListOptions,
	) (RoleAssignmentList, error)
	RevokeFn            func(context.Context, RoleAssignment) error
	RevokeByPrincipalFn func(context.Context, PrincipalReference) error
	ExistsFn            func(context.Context, RoleAssignment) (bool, error)
}

func (m *mockRoleAssignmentsStore) Grant(
	ctx context.Context,
	roleAssignment RoleAssignment,
) error {
	return m.GrantFn(ctx, roleAssignment)
}

func (m *mockRoleAssignmentsStore) List(
	ctx context.Context,
	selector RoleAssignmentsSelector,
	opts meta.ListOptions,
) (RoleAssignmentList, error) {
	return m.ListFn(ctx, selector, opts)
}

func (m *mockRoleAssignmentsStore) Revoke(
	ctx context.Context,
	roleAssignment RoleAssignment,
) error {
	return m.RevokeFn(ctx, roleAssignment)
}

func (m *mockRoleAssignmentsStore) RevokeByPrincipal(
	ctx context.Context,
	principalReference PrincipalReference,
) error {
	return m.RevokeByPrincipalFn(ctx, principalReference)
}

func (m *mockRoleAssignmentsStore) Exists(
	ctx context.Context,
	roleAssignment RoleAssignment,
) (bool, error) {
	return m.ExistsFn(ctx, roleAssignment)
}
