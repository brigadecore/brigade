package testing

import "github.com/brigadecore/brigade/sdk/v3"

type MockAuthnClient struct {
	ServiceAccountsClient sdk.ServiceAccountsClient
	SessionsClient        sdk.SessionsClient
	UsersClient           sdk.UsersClient
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
