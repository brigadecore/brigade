package testing

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/brigadecore/brigade/sdk/v3/meta"
)

type MockRoleAssignmentsClient struct {
	GrantFn func(
		context.Context,
		sdk.RoleAssignment,
		*sdk.RoleAssignmentGrantOptions,
	) error
	ListFn func(
		context.Context,
		*sdk.RoleAssignmentsSelector,
		*meta.ListOptions,
	) (sdk.RoleAssignmentList, error)
	RevokeFn func(
		context.Context,
		sdk.RoleAssignment,
		*sdk.RoleAssignmentRevokeOptions,
	) error
}

func (m *MockRoleAssignmentsClient) Grant(
	ctx context.Context,
	roleAssignment sdk.RoleAssignment,
	opts *sdk.RoleAssignmentGrantOptions,
) error {
	return m.GrantFn(ctx, roleAssignment, opts)
}

func (m *MockRoleAssignmentsClient) List(
	ctx context.Context,
	selector *sdk.RoleAssignmentsSelector,
	opts *meta.ListOptions,
) (sdk.RoleAssignmentList, error) {
	return m.ListFn(ctx, selector, opts)
}

func (m *MockRoleAssignmentsClient) Revoke(
	ctx context.Context,
	roleAssignment sdk.RoleAssignment,
	opts *sdk.RoleAssignmentRevokeOptions,
) error {
	return m.RevokeFn(ctx, roleAssignment, opts)
}
