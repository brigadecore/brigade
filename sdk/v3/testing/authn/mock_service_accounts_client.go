package authn

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3/authn"
	"github.com/brigadecore/brigade/sdk/v3/meta"
)

type MockServiceAccountsClient struct {
	CreateFn func(
		context.Context,
		authn.ServiceAccount,
		*authn.ServiceAccountCreateOptions,
	) (authn.Token, error)
	ListFn func(
		context.Context,
		*authn.ServiceAccountsSelector,
		*meta.ListOptions,
	) (authn.ServiceAccountList, error)
	GetFn func(
		context.Context,
		string,
		*authn.ServiceAccountGetOptions,
	) (authn.ServiceAccount, error)
	LockFn func(
		context.Context,
		string,
		*authn.ServiceAccountLockOptions,
	) error
	UnlockFn func(
		context.Context,
		string,
		*authn.ServiceAccountUnlockOptions,
	) (authn.Token, error)
	DeleteFn func(
		context.Context,
		string,
		*authn.ServiceAccountDeleteOptions,
	) error
}

func (m *MockServiceAccountsClient) Create(
	ctx context.Context,
	serviceAccount authn.ServiceAccount,
	opts *authn.ServiceAccountCreateOptions,
) (authn.Token, error) {
	return m.CreateFn(ctx, serviceAccount, opts)
}

func (m *MockServiceAccountsClient) List(
	ctx context.Context,
	selector *authn.ServiceAccountsSelector,
	opts *meta.ListOptions,
) (authn.ServiceAccountList, error) {
	return m.ListFn(ctx, selector, opts)
}

func (m *MockServiceAccountsClient) Get(
	ctx context.Context,
	id string,
	opts *authn.ServiceAccountGetOptions,
) (authn.ServiceAccount, error) {
	return m.GetFn(ctx, id, opts)
}

func (m *MockServiceAccountsClient) Lock(
	ctx context.Context,
	id string,
	opts *authn.ServiceAccountLockOptions,
) error {
	return m.LockFn(ctx, id, opts)
}

func (m *MockServiceAccountsClient) Unlock(
	ctx context.Context,
	id string,
	opts *authn.ServiceAccountUnlockOptions,
) (authn.Token, error) {
	return m.UnlockFn(ctx, id, opts)
}

func (m *MockServiceAccountsClient) Delete(
	ctx context.Context,
	id string,
	opts *authn.ServiceAccountDeleteOptions,
) error {
	return m.DeleteFn(ctx, id, opts)
}
