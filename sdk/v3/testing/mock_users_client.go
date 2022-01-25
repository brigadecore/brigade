package testing

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/brigadecore/brigade/sdk/v3/meta"
)

type MockUsersClient struct {
	ListFn func(
		context.Context,
		*sdk.UsersSelector,
		*meta.ListOptions,
	) (sdk.UserList, error)
	GetFn func(
		context.Context,
		string,
		*sdk.UserGetOptions,
	) (sdk.User, error)
	LockFn   func(context.Context, string, *sdk.UserLockOptions) error
	UnlockFn func(context.Context, string, *sdk.UserUnlockOptions) error
	DeleteFn func(context.Context, string, *sdk.UserDeleteOptions) error
}

func (m *MockUsersClient) List(
	ctx context.Context,
	selector *sdk.UsersSelector,
	opts *meta.ListOptions,
) (sdk.UserList, error) {
	return m.ListFn(ctx, selector, opts)
}

func (m *MockUsersClient) Get(
	ctx context.Context,
	id string,
	opts *sdk.UserGetOptions,
) (sdk.User, error) {
	return m.GetFn(ctx, id, opts)
}

func (m *MockUsersClient) Lock(
	ctx context.Context,
	id string,
	opts *sdk.UserLockOptions,
) error {
	return m.LockFn(ctx, id, opts)
}

func (m *MockUsersClient) Unlock(
	ctx context.Context,
	id string,
	opts *sdk.UserUnlockOptions,
) error {
	return m.UnlockFn(ctx, id, opts)
}

func (m *MockUsersClient) Delete(
	ctx context.Context,
	id string,
	opts *sdk.UserDeleteOptions,
) error {
	return m.DeleteFn(ctx, id, opts)
}
