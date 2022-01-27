package testing

import "github.com/brigadecore/brigade/sdk/v3"

type MockCoreClient struct {
	EventsClient    sdk.EventsClient
	ProjectsClient  sdk.ProjectsClient
	SubstrateClient sdk.SubstrateClient
}

func (m *MockCoreClient) Events() sdk.EventsClient {
	return m.EventsClient
}

func (m *MockCoreClient) Projects() sdk.ProjectsClient {
	return m.ProjectsClient
}

func (m *MockCoreClient) Substrate() sdk.SubstrateClient {
	return m.SubstrateClient
}
