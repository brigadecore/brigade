package sdk

import (
	"github.com/brigadecore/brigade/sdk/v2/authn"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// APIClient is the general interface for the Brigade API. It does little more
// than expose functions for obtaining more specialized clients for different
// areas of concern, like User management or Project management.
type APIClient interface {
	Authn() authn.APIClient
	Core() core.APIClient
}

type apiClient struct {
	authnClient authn.APIClient
	coreClient  core.APIClient
}

// NewAPIClient returns a Brigade client.
func NewAPIClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) APIClient {
	return &apiClient{
		authnClient: authn.NewAPIClient(apiAddress, apiToken, opts),
		coreClient:  core.NewAPIClient(apiAddress, apiToken, opts),
	}
}

func (a *apiClient) Authn() authn.APIClient {
	return a.authnClient
}

func (a *apiClient) Core() core.APIClient {
	return a.coreClient
}
