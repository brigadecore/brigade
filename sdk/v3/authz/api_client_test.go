package authz

import (
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient(rmTesting.TestAPIAddress, rmTesting.TestAPIToken, nil)
	require.NotNil(t, client.(*apiClient).roleAssignmentsClient)
	require.Equal(
		t,
		client.(*apiClient).roleAssignmentsClient,
		client.RoleAssignments(),
	)
}
