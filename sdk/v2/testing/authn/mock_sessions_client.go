package authn

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v2/authn"
)

type MockSessionsClient struct {
	CreateRootSessionFn func(
		ctx context.Context,
		password string,
	) (authn.Token, error)
	CreateUserSessionFn func(
		context.Context,
		*authn.ThirdPartyAuthOptions,
	) (authn.ThirdPartyAuthDetails, error)
	DeleteFn func(context.Context) error
}

func (m *MockSessionsClient) CreateRootSession(
	ctx context.Context,
	password string,
) (authn.Token, error) {
	return m.CreateRootSessionFn(ctx, password)
}

func (m *MockSessionsClient) CreateUserSession(
	ctx context.Context,
	opts *authn.ThirdPartyAuthOptions,
) (authn.ThirdPartyAuthDetails, error) {
	return m.CreateUserSessionFn(ctx, opts)
}

func (m *MockSessionsClient) Delete(ctx context.Context) error {
	return m.Delete(ctx)
}
