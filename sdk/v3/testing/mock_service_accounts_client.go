package testing

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/brigadecore/brigade/sdk/v3/meta"
)

type MockServiceAccountsClient struct {
	CreateFn func(
		context.Context,
		sdk.ServiceAccount,
		*sdk.ServiceAccountCreateOptions,
	) (sdk.Token, error)
	ListFn func(
		context.Context,
		*sdk.ServiceAccountsSelector,
		*meta.ListOptions,
	) (sdk.ServiceAccountList, error)
	GetFn func(
		context.Context,
		string,
		*sdk.ServiceAccountGetOptions,
	) (sdk.ServiceAccount, error)
	LockFn func(
		context.Context,
		string,
		*sdk.ServiceAccountLockOptions,
	) error
	UnlockFn func(
		context.Context,
		string,
		*sdk.ServiceAccountUnlockOptions,
	) (sdk.Token, error)
	DeleteFn func(
		context.Context,
		string,
		*sdk.ServiceAccountDeleteOptions,
	) error
}

func (m *MockServiceAccountsClient) Create(
	ctx context.Context,
	serviceAccount sdk.ServiceAccount,
	opts *sdk.ServiceAccountCreateOptions,
) (sdk.Token, error) {
	return m.CreateFn(ctx, serviceAccount, opts)
}

func (m *MockServiceAccountsClient) List(
	ctx context.Context,
	selector *sdk.ServiceAccountsSelector,
	opts *meta.ListOptions,
) (sdk.ServiceAccountList, error) {
	return m.ListFn(ctx, selector, opts)
}

func (m *MockServiceAccountsClient) Get(
	ctx context.Context,
	id string,
	opts *sdk.ServiceAccountGetOptions,
) (sdk.ServiceAccount, error) {
	return m.GetFn(ctx, id, opts)
}

func (m *MockServiceAccountsClient) Lock(
	ctx context.Context,
	id string,
	opts *sdk.ServiceAccountLockOptions,
) error {
	return m.LockFn(ctx, id, opts)
}

func (m *MockServiceAccountsClient) Unlock(
	ctx context.Context,
	id string,
	opts *sdk.ServiceAccountUnlockOptions,
) (sdk.Token, error) {
	return m.UnlockFn(ctx, id, opts)
}

func (m *MockServiceAccountsClient) Delete(
	ctx context.Context,
	id string,
	opts *sdk.ServiceAccountDeleteOptions,
) error {
	return m.DeleteFn(ctx, id, opts)
}
