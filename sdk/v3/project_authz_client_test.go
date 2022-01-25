package sdk

import (
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestNewProjectAuthzClient(t *testing.T) {
	client := NewProjectAuthzClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	)
	require.IsType(t, &projectAuthzClient{}, client)
	require.NotNil(t, client.(*projectAuthzClient).projectRoleAssignmentsClient)
	require.Equal(
		t,
		client.(*projectAuthzClient).projectRoleAssignmentsClient,
		client.RoleAssignments(),
	)
}
