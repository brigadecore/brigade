package authz

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3/authz"
	libAuthz "github.com/brigadecore/brigade/sdk/v3/lib/authz"
	"github.com/brigadecore/brigade/sdk/v3/meta"
)

type MockRoleAssignmentsClient struct {
	GrantFn func(
		context.Context,
		libAuthz.RoleAssignment,
		*authz.RoleAssignmentGrantOptions,
	) error
	ListFn func(
		context.Context,
		*authz.RoleAssignmentsSelector,
		*meta.ListOptions,
	) (authz.RoleAssignmentList, error)
	RevokeFn func(
		context.Context,
		libAuthz.RoleAssignment,
		*authz.RoleAssignmentRevokeOptions,
	) error
}

func (m *MockRoleAssignmentsClient) Grant(
	ctx context.Context,
	roleAssignment libAuthz.RoleAssignment,
	opts *authz.RoleAssignmentGrantOptions,
) error {
	return m.GrantFn(ctx, roleAssignment, opts)
}

func (m *MockRoleAssignmentsClient) List(
	ctx context.Context,
	selector *authz.RoleAssignmentsSelector,
	opts *meta.ListOptions,
) (authz.RoleAssignmentList, error) {
	return m.ListFn(ctx, selector, opts)
}

func (m *MockRoleAssignmentsClient) Revoke(
	ctx context.Context,
	roleAssignment libAuthz.RoleAssignment,
	opts *authz.RoleAssignmentRevokeOptions,
) error {
	return m.RevokeFn(ctx, roleAssignment, opts)
}
