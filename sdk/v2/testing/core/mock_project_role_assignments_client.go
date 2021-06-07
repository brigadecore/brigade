package core

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
)

type MockProjectRoleAssignmentsClient struct {
	GrantFn func(
		ctx context.Context,
		projectID string,
		projectRoleAssignment core.ProjectRoleAssignment,
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
	) error
}

func (m *MockProjectRoleAssignmentsClient) Grant(
	ctx context.Context,
	projectID string,
	projectRoleAssignment core.ProjectRoleAssignment,
) error {
	return m.GrantFn(ctx, projectID, projectRoleAssignment)
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
) error {
	return m.RevokeFn(ctx, projectID, projectRoleAssignment)
}
