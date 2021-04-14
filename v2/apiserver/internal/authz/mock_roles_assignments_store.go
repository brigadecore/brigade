package authz

import (
	"context"

	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
)

type MockRoleAssignmentsStore struct {
	GrantFn      func(context.Context, libAuthz.RoleAssignment) error
	RevokeFn     func(context.Context, libAuthz.RoleAssignment) error
	RevokeManyFn func(context.Context, libAuthz.RoleAssignment) error
	ExistsFn     func(context.Context, libAuthz.RoleAssignment) (bool, error)
}

func (m *MockRoleAssignmentsStore) Grant(
	ctx context.Context,
	roleAssignment libAuthz.RoleAssignment,
) error {
	return m.GrantFn(ctx, roleAssignment)
}

func (m *MockRoleAssignmentsStore) Revoke(
	ctx context.Context,
	roleAssignment libAuthz.RoleAssignment,
) error {
	return m.RevokeFn(ctx, roleAssignment)
}

func (m *MockRoleAssignmentsStore) RevokeMany(
	ctx context.Context,
	roleAssignment libAuthz.RoleAssignment,
) error {
	return m.RevokeManyFn(ctx, roleAssignment)
}

func (m *MockRoleAssignmentsStore) Exists(
	ctx context.Context,
	roleAssignment libAuthz.RoleAssignment,
) (bool, error) {
	return m.ExistsFn(ctx, roleAssignment)
}
