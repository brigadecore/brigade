package authz

import "context"

type MockRoleAssignmentsStore struct {
	GrantFn      func(context.Context, RoleAssignment) error
	RevokeFn     func(context.Context, RoleAssignment) error
	RevokeManyFn func(context.Context, RoleAssignment) error
	ExistsFn     func(context.Context, RoleAssignment) (bool, error)
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

func (m *MockRoleAssignmentsStore) RevokeMany(
	ctx context.Context,
	roleAssignment RoleAssignment,
) error {
	return m.RevokeManyFn(ctx, roleAssignment)
}

func (m *MockRoleAssignmentsStore) Exists(
	ctx context.Context,
	roleAssignment RoleAssignment,
) (bool, error) {
	return m.ExistsFn(ctx, roleAssignment)
}
