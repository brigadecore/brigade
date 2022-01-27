package testing

import (
	"github.com/brigadecore/brigade/sdk/v3"
)

type MockProjectAuthzClient struct {
	RoleAssignmentsClient sdk.ProjectRoleAssignmentsClient
}

func (
	m *MockProjectAuthzClient,
) RoleAssignments() sdk.ProjectRoleAssignmentsClient {
	return m.RoleAssignmentsClient
}
