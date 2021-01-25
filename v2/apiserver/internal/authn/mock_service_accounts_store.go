package authn

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
)

type MockServiceAccountStore struct {
	CreateFn func(context.Context, ServiceAccount) error
	ListFn   func(
		context.Context,
		meta.ListOptions,
	) (ServiceAccountList, error)
	GetFn              func(context.Context, string) (ServiceAccount, error)
	GetByHashedTokenFn func(context.Context, string) (ServiceAccount, error)
	LockFn             func(context.Context, string) error
	UnlockFn           func(
		ctx context.Context,
		id string,
		newHashedToken string,
	) error
}

func (m *MockServiceAccountStore) Create(
	ctx context.Context,
	serviceAccount ServiceAccount,
) error {
	return m.CreateFn(ctx, serviceAccount)
}

func (m *MockServiceAccountStore) List(
	ctx context.Context,
	opts meta.ListOptions,
) (ServiceAccountList, error) {
	return m.ListFn(ctx, opts)
}

func (m *MockServiceAccountStore) Get(
	ctx context.Context,
	id string,
) (ServiceAccount, error) {
	return m.GetFn(ctx, id)
}

func (m *MockServiceAccountStore) GetByHashedToken(
	ctx context.Context,
	token string,
) (ServiceAccount, error) {
	return m.GetByHashedTokenFn(ctx, token)
}

func (m *MockServiceAccountStore) Lock(ctx context.Context, id string) error {
	return m.LockFn(ctx, id)
}

func (m *MockServiceAccountStore) Unlock(
	ctx context.Context,
	id string,
	newHashedToken string,
) error {
	return m.UnlockFn(ctx, id, newHashedToken)
}
