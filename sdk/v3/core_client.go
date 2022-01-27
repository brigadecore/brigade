package sdk

import "github.com/brigadecore/brigade/sdk/v3/restmachinery"

// CoreClient is the root of a tree of more specialized API clients for dealing
// with Brigade's primary entities -- Projects, Events, and their constituent
// elements.
type CoreClient interface {
	// Events returns a specialized client for Event management.
	Events() EventsClient
	// Projects returns a specialized client for Project management.
	Projects() ProjectsClient
	// Substrate returns a specialized client for monitoring the state of the
	// substrate.
	Substrate() SubstrateClient
}

type coreClient struct {
	// eventsClient is a specialized client for Event management.
	eventsClient EventsClient
	// projectsClient is a specialized client for Project management.
	projectsClient ProjectsClient
	// substrateClient is a specialized client for substrate monitoring.
	substrateClient SubstrateClient
}

// NewCoreClient returns an CoreClient, which is the root of a tree of more
// specialized API clients for dealing with Brigade's primary entities --
// Projects, Events, and their constituent elements. It will initialize all
// clients in the tree so they are ready for immediate use.
func NewCoreClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) CoreClient {
	return &coreClient{
		eventsClient:    NewEventsClient(apiAddress, apiToken, opts),
		projectsClient:  NewProjectsClient(apiAddress, apiToken, opts),
		substrateClient: NewSubstrateClient(apiAddress, apiToken, opts),
	}
}

func (c *coreClient) Events() EventsClient {
	return c.eventsClient
}

func (c *coreClient) Projects() ProjectsClient {
	return c.projectsClient
}

func (c *coreClient) Substrate() SubstrateClient {
	return c.substrateClient
}
