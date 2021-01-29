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
	roles []libAuthz.Role
}

func (p *principal) Roles() []libAuthz.Role {
	return p.roles
}

func TestNewAuthorizer(t *testing.T) {
	roleAssignmentsStore := &authz.MockRoleAssignmentsStore{}
	svc := NewAuthorizer(roleAssignmentsStore)
	require.Same(t, roleAssignmentsStore, svc.(*authorizer).roleAssignmentsStore)
}

func TestAuthorizerAuthorize(t *testing.T) {
	testRequiredRole := libAuthz.Role{
		Type:  "foo",
		Name:  "foo",
		Scope: "foo",
	}
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
			name:       "roleHolder does not have role",
			principal:  &principal{},
			authorizer: &authorizer{},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "roleHolder has role",
			principal: &principal{
				roles: []libAuthz.Role{
					testRequiredRole,
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
					ExistsFn: func(context.Context, authz.RoleAssignment) (bool, error) {
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
					ExistsFn: func(context.Context, authz.RoleAssignment) (bool, error) {
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
					ExistsFn: func(context.Context, authz.RoleAssignment) (bool, error) {
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
					ExistsFn: func(context.Context, authz.RoleAssignment) (bool, error) {
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
					ExistsFn: func(context.Context, authz.RoleAssignment) (bool, error) {
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
					ExistsFn: func(context.Context, authz.RoleAssignment) (bool, error) {
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
			err := testCase.authorizer.Authorize(ctx, testRequiredRole)
			testCase.assertions(err)
		})
	}
}
