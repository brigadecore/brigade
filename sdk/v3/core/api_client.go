package core

import "github.com/brigadecore/brigade/sdk/v3/restmachinery"

// APIClient is the root of a tree of more specialized API clients within the
// core package.
type APIClient interface {
	// Events returns a specialized client for Event management.
	Events() EventsClient
	// Projects returns a specialized client for Project management.
	Projects() ProjectsClient
	// Substrate returns a specialized client for monitoring the state of the
	// substrate.
	Substrate() SubstrateClient
}

type apiClient struct {
	// eventsClient is a specialized client for Event management.
	eventsClient EventsClient
	// projectsClient is a specialized client for Project management.
	projectsClient ProjectsClient
	// substrateClient is a specialized client for substrate monitoring.
	substrateClient SubstrateClient
}

// NewAPIClient returns an APIClient, which is the root of a tree of more
// specialized API clients within the core package. It will initialize all
// clients in the tree so they are ready for immediate use.
func NewAPIClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) APIClient {
	return &apiClient{
		eventsClient:    NewEventsClient(apiAddress, apiToken, opts),
		projectsClient:  NewProjectsClient(apiAddress, apiToken, opts),
		substrateClient: NewSubstrateClient(apiAddress, apiToken, opts),
	}
}

func (a *apiClient) Events() EventsClient {
	return a.eventsClient
}

func (a *apiClient) Projects() ProjectsClient {
	return a.projectsClient
}

func (a *apiClient) Substrate() SubstrateClient {
	return a.substrateClient
}
