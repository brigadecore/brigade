package api

import (
	"context"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type principal struct {
	roleAssignments        []RoleAssignment
	projectRoleAssignments []ProjectRoleAssignment
}

func (p *principal) RoleAssignments() []RoleAssignment {
	return p.roleAssignments
}

func (p *principal) ProjectRoleAssignments() []ProjectRoleAssignment {
	return p.projectRoleAssignments
}

func TestNewAuthorizer(t *testing.T) {
	roleAssignmentsStore := &mockRoleAssignmentsStore{}
	svc := NewAuthorizer(roleAssignmentsStore)
	require.Same(t, roleAssignmentsStore, svc.(*authorizer).roleAssignmentsStore)
}

func TestAuthorizerAuthorize(t *testing.T) {
	const testRole = "foo"
	const testScope = "foo"
	testCases := []struct {
		name       string
		principal  interface{}
		authorizer Authorizer
		assertions func(error)
	}{
		{
			name:       "principal is nil",
			principal:  nil,
			authorizer: &authorizer{},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name:       "roleAssignmentsHolder does not have role",
			principal:  &principal{},
			authorizer: &authorizer{},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "roleAssignmentsHolder has role",
			principal: &principal{
				roleAssignments: []RoleAssignment{
					{
						Role:  testRole,
						Scope: testScope,
					},
				},
			},
			authorizer: &authorizer{},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
		{
			name:      "error looking up user role assignment",
			principal: &User{},
			authorizer: &authorizer{
				roleAssignmentsStore: &mockRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						RoleAssignment,
					) (bool, error) {
						return false, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name:      "user does not have role",
			principal: &User{},
			authorizer: &authorizer{
				roleAssignmentsStore: &mockRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						RoleAssignment,
					) (bool, error) {
						return false, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name:      "user has role",
			principal: &User{},
			authorizer: &authorizer{
				roleAssignmentsStore: &mockRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						RoleAssignment,
					) (bool, error) {
						return true, nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
		{
			name:      "error looking up service account role assignment",
			principal: &ServiceAccount{},
			authorizer: &authorizer{
				roleAssignmentsStore: &mockRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						RoleAssignment,
					) (bool, error) {
						return false, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name:      "service account does not have role",
			principal: &ServiceAccount{},
			authorizer: &authorizer{
				roleAssignmentsStore: &mockRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						RoleAssignment,
					) (bool, error) {
						return false, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name:      "service account has role",
			principal: &ServiceAccount{},
			authorizer: &authorizer{
				roleAssignmentsStore: &mockRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						RoleAssignment,
					) (bool, error) {
						return true, nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
		{
			name:       "principal is an unknown type",
			principal:  struct{}{},
			authorizer: &authorizer{},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = ContextWithPrincipal(ctx, testCase.principal)
			err := testCase.authorizer.Authorize(ctx, testRole, testScope)
			testCase.assertions(err)
		})
	}
}

// alwaysAuthorize is an implementation of the AuthorizeFn function signature
// that unconditionally passes authorization requests by returning nil. This is
// used only for testing purposes.
func alwaysAuthorize(ctx context.Context, role Role, scope string) error {
	return nil
}

// neverAuthorize is an implementation of the AuthorizeFn function signature
// that unconditionally fails authorization requests by returning a
// *meta.ErrAuthorization error. This is used only for testing purposes.
func neverAuthorize(ctx context.Context, role Role, scope string) error {
	return &meta.ErrAuthorization{}
}
