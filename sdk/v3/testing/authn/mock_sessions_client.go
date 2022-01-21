package authn

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3/authn"
)

type MockSessionsClient struct {
	CreateRootSessionFn func(
		ctx context.Context,
		password string,
		opts *authn.RootSessionCreateOptions,
	) (authn.Token, error)
	CreateUserSessionFn func(
		context.Context,
		*authn.UserSessionCreateOptions,
	) (authn.ThirdPartyAuthDetails, error)
	DeleteFn func(context.Context, *authn.SessionDeleteOptions) error
}

func (m *MockSessionsClient) CreateRootSession(
	ctx context.Context,
	password string,
	opts *authn.RootSessionCreateOptions,
) (authn.Token, error) {
	return m.CreateRootSessionFn(ctx, password, opts)
}

func (m *MockSessionsClient) CreateUserSession(
	ctx context.Context,
	opts *authn.UserSessionCreateOptions,
) (authn.ThirdPartyAuthDetails, error) {
	return m.CreateUserSessionFn(ctx, opts)
}

func (m *MockSessionsClient) Delete(
	ctx context.Context,
	opts *authn.SessionDeleteOptions,
) error {
	return m.DeleteFn(ctx, opts)
}
