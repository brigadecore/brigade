package core

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
)

type MockSecretsClient struct {
	ListFn func(
		ctx context.Context,
		projectID string,
		opts *meta.ListOptions,
	) (core.SecretList, error)
	SetFn   func(ctx context.Context, projectID string, secret core.Secret) error
	UnsetFn func(ctx context.Context, projectID string, key string) error
}

func (m *MockSecretsClient) List(
	ctx context.Context,
	projectID string,
	opts *meta.ListOptions,
) (core.SecretList, error) {
	return m.ListFn(ctx, projectID, opts)
}

func (m *MockSecretsClient) Set(
	ctx context.Context,
	projectID string,
	secret core.Secret,
) error {
	return m.SetFn(ctx, projectID, secret)
}

func (m *MockSecretsClient) Unset(
	ctx context.Context,
	projectID string,
	key string,
) error {
	return m.UnsetFn(ctx, projectID, key)
}
