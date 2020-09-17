package core

// APIClient is the root of a tree of more specialized API clients within the
// core package.
type APIClient interface {
	// Events returns a specialized client for Event management.
	Events() EventsClient
	// Projects returns a specialized client for Project management.
	Projects() ProjectsClient
}

type apiClient struct {
	// eventsClient is a specialized client for Event management.
	eventsClient EventsClient
	// projectsClient is a specialized client for Project management.
	projectsClient ProjectsClient
}

// NewAPIClient returns an APIClient, which is the root of a tree of more
// specialized API clients within the core package. It will initialize all
// clients in the tree so they are ready for immediate use.
func NewAPIClient(apiAddress, apiToken string, allowInsecure bool) APIClient {
	return &apiClient{
		eventsClient:   NewEventsClient(apiAddress, apiToken, allowInsecure),
		projectsClient: NewProjectsClient(apiAddress, apiToken, allowInsecure),
	}
}

func (a *apiClient) Events() EventsClient {
	return a.eventsClient
}

func (a *apiClient) Projects() ProjectsClient {
	return a.projectsClient
}
