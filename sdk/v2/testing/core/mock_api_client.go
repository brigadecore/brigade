package core

import "github.com/brigadecore/brigade/sdk/v2/core"

type MockAPIClient struct {
	EventsClient    core.EventsClient
	ProjectsClient  core.ProjectsClient
	SubstrateClient core.SubstrateClient
}

func (m *MockAPIClient) Events() core.EventsClient {
	return m.EventsClient
}

func (m *MockAPIClient) Projects() core.ProjectsClient {
	return m.ProjectsClient
}

func (m *MockAPIClient) Substrate() core.SubstrateClient {
	return m.SubstrateClient
}
