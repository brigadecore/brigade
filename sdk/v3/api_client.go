package sdk

import (
	"github.com/brigadecore/brigade/sdk/v3/restmachinery"
)

// APIClient is the general interface for the Brigade API. It does little more
// than expose functions for obtaining more specialized clients for different
// areas of concern, like User management or Project management.
type APIClient interface {
	Authn() AuthnClient
	Authz() SystemAuthzClient
	Core() CoreClient
	System() SystemClient
}

type apiClient struct {
	authnClient  AuthnClient
	authzClient  SystemAuthzClient
	coreClient   CoreClient
	systemClient SystemClient
}

// NewAPIClient returns a Brigade client.
func NewAPIClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) APIClient {
	return &apiClient{
		authnClient:  NewAuthnClient(apiAddress, apiToken, opts),
		authzClient:  NewSystemAuthzClient(apiAddress, apiToken, opts),
		coreClient:   NewCoreClient(apiAddress, apiToken, opts),
		systemClient: NewSystemClient(apiAddress, apiToken, opts),
	}
}

func (a *apiClient) Authn() AuthnClient {
	return a.authnClient
}

func (a *apiClient) Authz() SystemAuthzClient {
	return a.authzClient
}

func (a *apiClient) Core() CoreClient {
	return a.coreClient
}

func (a *apiClient) System() SystemClient {
	return a.systemClient
}
