package testing

import (
	"github.com/brigadecore/brigade/sdk/v3"
)

type MockAPIClient struct {
	AuthnClient  sdk.AuthnClient
	AuthzClient  sdk.SystemAuthzClient
	CoreClient   sdk.CoreClient
	SystemClient sdk.SystemClient
}

func (m *MockAPIClient) Authn() sdk.AuthnClient {
	return m.AuthnClient
}

func (m *MockAPIClient) Authz() sdk.SystemAuthzClient {
	return m.AuthzClient
}

func (m *MockAPIClient) Core() sdk.CoreClient {
	return m.CoreClient
}

func (m *MockAPIClient) System() sdk.SystemClient {
	return m.SystemClient
}
