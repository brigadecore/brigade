package testing

import (
	"github.com/brigadecore/brigade/sdk/v3/authn"
	"github.com/brigadecore/brigade/sdk/v3/authz"
	"github.com/brigadecore/brigade/sdk/v3/core"
	"github.com/brigadecore/brigade/sdk/v3/system"
)

type MockAPIClient struct {
	AuthnClient  authn.APIClient
	AuthzClient  authz.APIClient
	CoreClient   core.APIClient
	SystemClient system.APIClient
}

func (m *MockAPIClient) Authn() authn.APIClient {
	return m.AuthnClient
}

func (m *MockAPIClient) Authz() authz.APIClient {
	return m.AuthzClient
}

func (m *MockAPIClient) Core() core.APIClient {
	return m.CoreClient
}

func (m *MockAPIClient) System() system.APIClient {
	return m.SystemClient
}
