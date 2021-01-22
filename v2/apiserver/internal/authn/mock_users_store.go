package authn

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
)

type MockUsersStore struct {
	CreateFn func(context.Context, User) error
	ListFn   func(context.Context, meta.ListOptions) (UserList, error)
	GetFn    func(context.Context, string) (User, error)
	LockFn   func(context.Context, string) error
	UnlockFn func(context.Context, string) error
}

func (m *MockUsersStore) Create(ctx context.Context, user User) error {
	return m.CreateFn(ctx, user)
}

func (m *MockUsersStore) List(
	ctx context.Context,
	opts meta.ListOptions,
) (UserList, error) {
	return m.ListFn(ctx, opts)
}

func (m *MockUsersStore) Get(ctx context.Context, id string) (User, error) {
	return m.GetFn(ctx, id)
}

func (m *MockUsersStore) Lock(ctx context.Context, id string) error {
	return m.LockFn(ctx, id)
}

func (m *MockUsersStore) Unlock(ctx context.Context, id string) error {
	return m.UnlockFn(ctx, id)
}
