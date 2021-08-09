package api

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
)

func TestNewProjectAuthorizer(t *testing.T) {
	projectRoleAssignmentsStore := &mockProjectRoleAssignmentsStore{}
	svc := NewProjectAuthorizer(projectRoleAssignmentsStore)
	require.Same(
		t,
		projectRoleAssignmentsStore,
		svc.(*projectAuthorizer).projectRoleAssignmentsStore,
	)
}

func TestProjectAuthorizerAuthorize(t *testing.T) {
	const testProjectRole = "foo"
	const testProjectID = "foo"
	testCases := []struct {
		name              string
		principal         interface{}
		projectAuthorizer ProjectAuthorizer
		assertions        func(error)
	}{
		{
			name:              "principal is nil",
			principal:         nil,
			projectAuthorizer: &projectAuthorizer{},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name:              "projectRoleAssignmentsHolder does not have project role", // nolint: lll
			principal:         &principal{},
			projectAuthorizer: &projectAuthorizer{},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "projectRoleAssignmentsHolder has project role",
			principal: &principal{
				projectRoleAssignments: []ProjectRoleAssignment{
					{
						ProjectID: testProjectID,
						Role:      testProjectRole,
					},
				},
			},
			projectAuthorizer: &projectAuthorizer{},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
		{
			name:      "error looking up user project role assignment",
			principal: &User{},
			projectAuthorizer: &projectAuthorizer{
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						ProjectRoleAssignment,
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
			name:      "user does not have project role",
			principal: &User{},
			projectAuthorizer: &projectAuthorizer{
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						ProjectRoleAssignment,
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
			name:      "user has project role",
			principal: &User{},
			projectAuthorizer: &projectAuthorizer{
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						ProjectRoleAssignment,
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
			name:      "error looking up service account project role assignment",
			principal: &ServiceAccount{},
			projectAuthorizer: &projectAuthorizer{
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						ProjectRoleAssignment,
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
			name:      "service account does not have project role",
			principal: &ServiceAccount{},
			projectAuthorizer: &projectAuthorizer{
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						ProjectRoleAssignment,
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
			name:      "service account has project role",
			principal: &ServiceAccount{},
			projectAuthorizer: &projectAuthorizer{
				projectRoleAssignmentsStore: &mockProjectRoleAssignmentsStore{
					ExistsFn: func(
						context.Context,
						ProjectRoleAssignment,
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
			name:              "principal is an unknown type",
			principal:         struct{}{},
			projectAuthorizer: &projectAuthorizer{},
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
			err := testCase.projectAuthorizer.Authorize(
				ctx,
				testProjectID,
				testProjectRole,
			)
			testCase.assertions(err)
		})
	}
}

// alwaysProjectAuthorize is an implementation of the ProjectAuthorizeFn
// function signature that unconditionally passes authorization requests by
// returning nil. This is used only for testing purposes.
func alwaysProjectAuthorize(context.Context, string, Role) error {
	return nil
}

// neverProjectAuthorize is an implementation of the ProjectAuthorizeFn function
// signature that unconditionally fails authorization requests by returning a
// *meta.ErrAuthorization error. This is used only for testing purposes.
func neverProjectAuthorize(context.Context, string, Role) error {
	return &meta.ErrAuthorization{}
}
