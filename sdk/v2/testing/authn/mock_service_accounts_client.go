package authn

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v2/authn"
	"github.com/brigadecore/brigade/sdk/v2/meta"
)

type MockServiceAccountsClient struct {
	CreateFn func(context.Context, authn.ServiceAccount) (authn.Token, error)
	ListFn   func(
		context.Context,
		*authn.ServiceAccountsSelector,
		*meta.ListOptions,
	) (authn.ServiceAccountList, error)
	GetFn    func(context.Context, string) (authn.ServiceAccount, error)
	LockFn   func(context.Context, string) error
	UnlockFn func(context.Context, string) (authn.Token, error)
}

func (m *MockServiceAccountsClient) Create(
	ctx context.Context,
	serviceAccount authn.ServiceAccount,
) (authn.Token, error) {
	return m.CreateFn(ctx, serviceAccount)
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
) (authn.ServiceAccount, error) {
	return m.GetFn(ctx, id)
}

func (m *MockServiceAccountsClient) Lock(ctx context.Context, id string) error {
	return m.LockFn(ctx, id)
}

func (m *MockServiceAccountsClient) Unlock(
	ctx context.Context,
	id string,
) (authn.Token, error) {
	return m.UnlockFn(ctx, id)
}
