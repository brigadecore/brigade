package authn

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3/authn"
	"github.com/brigadecore/brigade/sdk/v3/meta"
)

type MockUsersClient struct {
	ListFn func(
		context.Context,
		*authn.UsersSelector,
		*meta.ListOptions,
	) (authn.UserList, error)
	GetFn func(
		context.Context,
		string,
		*authn.UserGetOptions,
	) (authn.User, error)
	LockFn   func(context.Context, string, *authn.UserLockOptions) error
	UnlockFn func(context.Context, string, *authn.UserUnlockOptions) error
	DeleteFn func(context.Context, string, *authn.UserDeleteOptions) error
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
	opts *authn.UserGetOptions,
) (authn.User, error) {
	return m.GetFn(ctx, id, opts)
}

func (m *MockUsersClient) Lock(
	ctx context.Context,
	id string,
	opts *authn.UserLockOptions,
) error {
	return m.LockFn(ctx, id, opts)
}

func (m *MockUsersClient) Unlock(
	ctx context.Context,
	id string,
	opts *authn.UserUnlockOptions,
) error {
	return m.UnlockFn(ctx, id, opts)
}

func (m *MockUsersClient) Delete(
	ctx context.Context,
	id string,
	opts *authn.UserDeleteOptions,
) error {
	return m.DeleteFn(ctx, id, opts)
}
