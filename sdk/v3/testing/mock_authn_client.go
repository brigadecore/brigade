package testing

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3"
)

type MockAuthnClient struct {
	WhoAmIFn              func(context.Context) (sdk.PrincipalReference, error)
	ServiceAccountsClient sdk.ServiceAccountsClient
	SessionsClient        sdk.SessionsClient
	UsersClient           sdk.UsersClient
}

func (m *MockAuthnClient) WhoAmI(
	ctx context.Context,
) (sdk.PrincipalReference, error) {
	return m.WhoAmIFn(ctx)
}

func (m *MockAuthnClient) ServiceAccounts() sdk.ServiceAccountsClient {
	return m.ServiceAccountsClient
}

func (m *MockAuthnClient) Sessions() sdk.SessionsClient {
	return m.SessionsClient
}

func (m *MockAuthnClient) Users() sdk.UsersClient {
	return m.UsersClient
}
