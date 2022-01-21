package authz

import (
	"testing"

	"github.com/brigadecore/brigade/sdk/v3/authz"
	"github.com/stretchr/testify/require"
)

func TestMockRoleAssignmentsClient(t *testing.T) {
	require.Implements(
		t,
		(*authz.RoleAssignmentsClient)(nil),
		&MockRoleAssignmentsClient{},
	)
}
