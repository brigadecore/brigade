package testing

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/brigadecore/brigade/sdk/v3/meta"
)

type MockSecretsClient struct {
	ListFn func(
		ctx context.Context,
		projectID string,
		opts *meta.ListOptions,
	) (sdk.SecretList, error)
	SetFn func(
		ctx context.Context,
		projectID string,
		secret sdk.Secret,
		opts *sdk.SecretSetOptions,
	) error
	UnsetFn func(
		ctx context.Context,
		projectID string,
		key string,
		opts *sdk.SecretUnsetOptions,
	) error
}

func (m *MockSecretsClient) List(
	ctx context.Context,
	projectID string,
	opts *meta.ListOptions,
) (sdk.SecretList, error) {
	return m.ListFn(ctx, projectID, opts)
}

func (m *MockSecretsClient) Set(
	ctx context.Context,
	projectID string,
	secret sdk.Secret,
	opts *sdk.SecretSetOptions,
) error {
	return m.SetFn(ctx, projectID, secret, opts)
}

func (m *MockSecretsClient) Unset(
	ctx context.Context,
	projectID string,
	key string,
	opts *sdk.SecretUnsetOptions,
) error {
	return m.UnsetFn(ctx, projectID, key, opts)
}
