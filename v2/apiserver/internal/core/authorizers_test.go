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
	projectRoleAssignments []ProjectRoleAssignment
}

func (p *principal) ProjectRoleAssignments() []ProjectRoleAssignment {
	return p.projectRoleAssignments
}

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
			principal: &authn.User{},
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
			principal: &authn.User{},
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
			principal: &authn.User{},
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
			principal: &authn.ServiceAccount{},
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
			principal: &authn.ServiceAccount{},
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
			principal: &authn.ServiceAccount{},
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
			ctx = libAuthn.ContextWithPrincipal(ctx, testCase.principal)
			err := testCase.projectAuthorizer.Authorize(
				ctx,
				testProjectID,
				testProjectRole,
			)
			testCase.assertions(err)
		})
	}
}
