package authz

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	"github.com/brigadecore/brigade/v2/apiserver/internal/authz"
	libAuthn "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authn"
	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
)

type principal struct {
	roleAssignments []libAuthz.RoleAssignment
}

func (p *principal) RoleAssignments() []libAuthz.RoleAssignment {
	return p.roleAssignments
}

func TestNewAuthorizer(t *testing.T) {
	roleAssignmentsStore := &authz.MockRoleAssignmentsStore{}
	svc := NewAuthorizer(roleAssignmentsStore)
	require.Same(t, roleAssignmentsStore, svc.(*authorizer).roleAssignmentsStore)
}

func TestAuthorizerAuthorize(t *testing.T) {
	testRole := libAuthz.Role{
		Type: "foo",
		Name: "foo",
	}
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
				roleAssignments: []libAuthz.RoleAssignment{
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
			principal: &authn.User{},
			authorizer: &authorizer{
				roleAssignmentsStore: &authz.MockRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						libAuthz.RoleAssignment,
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
			principal: &authn.User{},
			authorizer: &authorizer{
				roleAssignmentsStore: &authz.MockRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						libAuthz.RoleAssignment,
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
			principal: &authn.User{},
			authorizer: &authorizer{
				roleAssignmentsStore: &authz.MockRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						libAuthz.RoleAssignment,
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
			principal: &authn.ServiceAccount{},
			authorizer: &authorizer{
				roleAssignmentsStore: &authz.MockRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						libAuthz.RoleAssignment,
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
			principal: &authn.ServiceAccount{},
			authorizer: &authorizer{
				roleAssignmentsStore: &authz.MockRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						libAuthz.RoleAssignment,
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
			principal: &authn.ServiceAccount{},
			authorizer: &authorizer{
				roleAssignmentsStore: &authz.MockRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						libAuthz.RoleAssignment,
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
			ctx = libAuthn.ContextWithPrincipal(ctx, testCase.principal)
			err := testCase.authorizer.Authorize(ctx, testRole, testScope)
			testCase.assertions(err)
		})
	}
}
