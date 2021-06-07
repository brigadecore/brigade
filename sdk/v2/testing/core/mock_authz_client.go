package core

import "github.com/brigadecore/brigade/sdk/v2/core"

type MockAuthzClient struct {
	RoleAssignmentsClient core.ProjectRoleAssignmentsClient
}

func (m *MockAuthzClient) RoleAssignments() core.ProjectRoleAssignmentsClient {
	return m.RoleAssignmentsClient
}
