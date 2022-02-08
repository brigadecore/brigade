package sdk

import (
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestNewProjectAuthzClient(t *testing.T) {
	client, ok := NewProjectAuthzClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	).(*projectAuthzClient)
	require.True(t, ok)
	require.NotNil(t, client.projectRoleAssignmentsClient)
	require.Equal(
		t,
		client.projectRoleAssignmentsClient,
		client.RoleAssignments(),
	)
}
