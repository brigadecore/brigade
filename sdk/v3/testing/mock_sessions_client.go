package testing

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3"
)

type MockSessionsClient struct {
	CreateRootSessionFn func(
		ctx context.Context,
		password string,
		opts *sdk.RootSessionCreateOptions,
	) (sdk.Token, error)
	CreateUserSessionFn func(
		context.Context,
		*sdk.UserSessionCreateOptions,
	) (sdk.ThirdPartyAuthDetails, error)
	DeleteFn func(context.Context, *sdk.SessionDeleteOptions) error
}

func (m *MockSessionsClient) CreateRootSession(
	ctx context.Context,
	password string,
	opts *sdk.RootSessionCreateOptions,
) (sdk.Token, error) {
	return m.CreateRootSessionFn(ctx, password, opts)
}

func (m *MockSessionsClient) CreateUserSession(
	ctx context.Context,
	opts *sdk.UserSessionCreateOptions,
) (sdk.ThirdPartyAuthDetails, error) {
	return m.CreateUserSessionFn(ctx, opts)
}

func (m *MockSessionsClient) Delete(
	ctx context.Context,
	opts *sdk.SessionDeleteOptions,
) error {
	return m.DeleteFn(ctx, opts)
}
