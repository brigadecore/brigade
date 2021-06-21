package authn

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v2/authn"
	"github.com/brigadecore/brigade/sdk/v2/meta"
)

type MockUsersClient struct {
	ListFn func(
		context.Context,
		*authn.UsersSelector,
		*meta.ListOptions,
	) (authn.UserList, error)
	GetFn    func(context.Context, string) (authn.User, error)
	LockFn   func(context.Context, string) error
	UnlockFn func(context.Context, string) error
}

func (m *MockUsersClient) List(
	ctx context.Context,
	selector *authn.UsersSelector,
	opts *meta.ListOptions,
) (authn.UserList, error) {
	return m.ListFn(ctx, selector, opts)
}

func (m *MockUsersClient) Get(
	ctx context.Context,
	id string,
) (authn.User, error) {
	return m.GetFn(ctx, id)
}

func (m *MockUsersClient) Lock(ctx context.Context, id string) error {
	return m.LockFn(ctx, id)
}

func (m *MockUsersClient) Unlock(ctx context.Context, id string) error {
	return m.UnlockFn(ctx, id)
}
