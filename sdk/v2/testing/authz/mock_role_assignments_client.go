package authz

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v2/authz"
	libAuthz "github.com/brigadecore/brigade/sdk/v2/lib/authz"
	"github.com/brigadecore/brigade/sdk/v2/meta"
)

type MockRoleAssignmentsClient struct {
	GrantFn func(context.Context, libAuthz.RoleAssignment) error
	ListFn  func(
		context.Context,
		*authz.RoleAssignmentsSelector,
		*meta.ListOptions,
	) (authz.RoleAssignmentList, error)
	RevokeFn func(context.Context, libAuthz.RoleAssignment) error
}

func (m *MockRoleAssignmentsClient) Grant(
	ctx context.Context,
	roleAssignment libAuthz.RoleAssignment,
) error {
	return m.GrantFn(ctx, roleAssignment)
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
) error {
	return m.RevokeFn(ctx, roleAssignment)
}
