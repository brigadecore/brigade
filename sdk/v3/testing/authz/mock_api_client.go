package authz

import "github.com/brigadecore/brigade/sdk/v3/authz"

type MockAPIClient struct {
	RoleAssignmentsClient authz.RoleAssignmentsClient
}

func (m *MockAPIClient) RoleAssignments() authz.RoleAssignmentsClient {
	return m.RoleAssignmentsClient
}
