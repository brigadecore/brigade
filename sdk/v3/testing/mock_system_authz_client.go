package testing

import (
	"github.com/brigadecore/brigade/sdk/v3"
)

type MockSystemAuthzClient struct {
	RoleAssignmentsClient sdk.RoleAssignmentsClient
}

func (m *MockSystemAuthzClient) RoleAssignments() sdk.RoleAssignmentsClient {
	return m.RoleAssignmentsClient
}
