package testing

import (
	"testing"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/stretchr/testify/require"
)

func TestMockProjectsClient(t *testing.T) {
	require.Implements(t, (*sdk.ProjectsClient)(nil), &MockProjectsClient{})
}
