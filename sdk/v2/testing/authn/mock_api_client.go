package authn

import "github.com/brigadecore/brigade/sdk/v2/authn"

type MockAPIClient struct {
	ServiceAccountsClient authn.ServiceAccountsClient
	SessionsClient        authn.SessionsClient
	UsersClient           authn.UsersClient
}

func (m *MockAPIClient) ServiceAccounts() authn.ServiceAccountsClient {
	return m.ServiceAccountsClient
}

func (m *MockAPIClient) Sessions() authn.SessionsClient {
	return m.SessionsClient
}

func (m *MockAPIClient) Users() authn.UsersClient {
	return m.UsersClient
}
