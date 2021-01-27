package authz

import "context"

type MockRoleAssignmentsStore struct {
	GrantFn  func(context.Context, RoleAssignment) error
	RevokeFn func(context.Context, RoleAssignment) error
}

func (m *MockRoleAssignmentsStore) Grant(
	ctx context.Context,
	roleAssignment RoleAssignment,
) error {
	return m.GrantFn(ctx, roleAssignment)
}

func (m *MockRoleAssignmentsStore) Revoke(
	ctx context.Context,
	roleAssignment RoleAssignment,
) error {
	return m.RevokeFn(ctx, roleAssignment)
}
