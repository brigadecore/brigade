package core

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3/core"
	"github.com/brigadecore/brigade/sdk/v3/meta"
)

type MockProjectRoleAssignmentsClient struct {
	GrantFn func(
		ctx context.Context,
		projectID string,
		projectRoleAssignment core.ProjectRoleAssignment,
		opts *core.ProjectRoleAssignmentGrantOptions,
	) error
	ListFn func(
		ctx context.Context,
		projectID string,
		selector *core.ProjectRoleAssignmentsSelector,
		opts *meta.ListOptions,
	) (core.ProjectRoleAssignmentList, error)
	RevokeFn func(
		ctx context.Context,
		projectID string,
		projectRoleAssignment core.ProjectRoleAssignment,
		opts *core.ProjectRoleAssignmentRevokeOptions,
	) error
}

func (m *MockProjectRoleAssignmentsClient) Grant(
	ctx context.Context,
	projectID string,
	projectRoleAssignment core.ProjectRoleAssignment,
	opts *core.ProjectRoleAssignmentGrantOptions,
) error {
	return m.GrantFn(ctx, projectID, projectRoleAssignment, opts)
}

func (m *MockProjectRoleAssignmentsClient) List(
	ctx context.Context,
	projectID string,
	selector *core.ProjectRoleAssignmentsSelector,
	opts *meta.ListOptions,
) (core.ProjectRoleAssignmentList, error) {
	return m.ListFn(ctx, projectID, selector, opts)
}

func (m *MockProjectRoleAssignmentsClient) Revoke(
	ctx context.Context,
	projectID string,
	projectRoleAssignment core.ProjectRoleAssignment,
	opts *core.ProjectRoleAssignmentRevokeOptions,
) error {
	return m.RevokeFn(ctx, projectID, projectRoleAssignment, opts)
}
