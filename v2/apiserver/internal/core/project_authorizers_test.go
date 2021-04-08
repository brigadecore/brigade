package core

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	libAuthn "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authn"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
)

type principal struct {
	projectRoles []ProjectRole
}

func (p *principal) ProjectRoles() []ProjectRole {
	return p.projectRoles
}

func TestNewProjectRoleAuthorizer(t *testing.T) {
	roleAssignmentsStore := &mockProjectRoleAssignmentsStore{}
	svc := NewProjectRoleAuthorizer(roleAssignmentsStore)
	require.Same(
		t,
		roleAssignmentsStore,
		svc.(*projectRoleAuthorizer).projectRoleAssignmentsStore,
	)
}

func TestProjectRoleAuthorizerAuthorize(t *testing.T) {
	testRequiredRole := ProjectRole{
		Name:      "foo",
		ProjectID: "foo",
	}
	testCases := []struct {
		name                  string
		principal             interface{}
		projectRoleAuthorizer ProjectRoleAuthorizer
		assertions            func(error)
	}{
		{
			name:                  "principal is nil",
			principal:             nil,
			projectRoleAuthorizer: &projectRoleAuthorizer{},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name:                  "roleHolder does not have role",
			principal:             &principal{},
			projectRoleAuthorizer: &projectRoleAuthorizer{},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "roleHolder has role",
			principal: &principal{
				projectRoles: []ProjectRole{
					testRequiredRole,
				},
			},
			projectRoleAuthorizer: &projectRoleAuthorizer{},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
		{
			name:      "error looking up user role assignment",
			principal: &authn.User{},
			projectRoleAuthorizer: &projectRoleAuthorizer{
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					ExistsFn: func(context.Context, ProjectRoleAssignment) (bool, error) {
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
			projectRoleAuthorizer: &projectRoleAuthorizer{
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					ExistsFn: func(context.Context, ProjectRoleAssignment) (bool, error) {
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
			projectRoleAuthorizer: &projectRoleAuthorizer{
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					ExistsFn: func(context.Context, ProjectRoleAssignment) (bool, error) {
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
			projectRoleAuthorizer: &projectRoleAuthorizer{
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					ExistsFn: func(context.Context, ProjectRoleAssignment) (bool, error) {
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
			projectRoleAuthorizer: &projectRoleAuthorizer{
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					ExistsFn: func(context.Context, ProjectRoleAssignment) (bool, error) {
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
			projectRoleAuthorizer: &projectRoleAuthorizer{
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					ExistsFn: func(context.Context, ProjectRoleAssignment) (bool, error) {
						return true, nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
		{
			name:                  "principal is an unknown type",
			principal:             struct{}{},
			projectRoleAuthorizer: &projectRoleAuthorizer{},
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
			err := testCase.projectRoleAuthorizer.Authorize(ctx, testRequiredRole)
			testCase.assertions(err)
		})
	}
}

type mockProjectRoleAssignmentsStore struct {
	GrantFn      func(context.Context, ProjectRoleAssignment) error
	RevokeFn     func(context.Context, ProjectRoleAssignment) error
	RevokeManyFn func(context.Context, string) error
	ExistsFn     func(context.Context, ProjectRoleAssignment) (bool, error)
}

func (m *mockProjectRoleAssignmentsStore) Grant(
	ctx context.Context,
	roleAssignment ProjectRoleAssignment,
) error {
	return m.GrantFn(ctx, roleAssignment)
}

func (m *mockProjectRoleAssignmentsStore) Revoke(
	ctx context.Context,
	roleAssignment ProjectRoleAssignment,
) error {
	return m.RevokeFn(ctx, roleAssignment)
}

func (m *mockProjectRoleAssignmentsStore) RevokeMany(
	ctx context.Context,
	projectID string,
) error {
	return m.RevokeManyFn(ctx, projectID)
}

func (m *mockProjectRoleAssignmentsStore) Exists(
	ctx context.Context,
	roleAssignment ProjectRoleAssignment,
) (bool, error) {
	return m.ExistsFn(ctx, roleAssignment)
}

// alwaysAuthorize is an implementation of the ProjectAuthorizeFn function
// signature that unconditionally passes authorization requests by returning
// nil. This is used only for testing purposes.
func alwaysAuthorize(context.Context, ...ProjectRole) error {
	return nil
}

// neverAuthorize is an implementation of the ProjectAuthorizeFn function
// signature that unconditionally fails authorization requests by returning a
// *meta.ErrAuthorization error. This is used only for testing purposes.
func neverAuthorize(context.Context, ...ProjectRole) error {
	return &meta.ErrAuthorization{}
}
