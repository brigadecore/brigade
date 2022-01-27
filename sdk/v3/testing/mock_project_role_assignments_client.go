package testing

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/brigadecore/brigade/sdk/v3/meta"
)

type MockProjectRoleAssignmentsClient struct {
	GrantFn func(
		ctx context.Context,
		projectID string,
		projectRoleAssignment sdk.ProjectRoleAssignment,
		opts *sdk.ProjectRoleAssignmentGrantOptions,
	) error
	ListFn func(
		ctx context.Context,
		projectID string,
		selector *sdk.ProjectRoleAssignmentsSelector,
		opts *meta.ListOptions,
	) (sdk.ProjectRoleAssignmentList, error)
	RevokeFn func(
		ctx context.Context,
		projectID string,
		projectRoleAssignment sdk.ProjectRoleAssignment,
		opts *sdk.ProjectRoleAssignmentRevokeOptions,
	) error
}

func (m *MockProjectRoleAssignmentsClient) Grant(
	ctx context.Context,
	projectID string,
	projectRoleAssignment sdk.ProjectRoleAssignment,
	opts *sdk.ProjectRoleAssignmentGrantOptions,
) error {
	return m.GrantFn(ctx, projectID, projectRoleAssignment, opts)
}

func (m *MockProjectRoleAssignmentsClient) List(
	ctx context.Context,
	projectID string,
	selector *sdk.ProjectRoleAssignmentsSelector,
	opts *meta.ListOptions,
) (sdk.ProjectRoleAssignmentList, error) {
	return m.ListFn(ctx, projectID, selector, opts)
}

func (m *MockProjectRoleAssignmentsClient) Revoke(
	ctx context.Context,
	projectID string,
	projectRoleAssignment sdk.ProjectRoleAssignment,
	opts *sdk.ProjectRoleAssignmentRevokeOptions,
) error {
	return m.RevokeFn(ctx, projectID, projectRoleAssignment, opts)
}
