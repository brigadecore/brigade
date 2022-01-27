package testing

import (
	"testing"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/stretchr/testify/require"
)

func TestMockLogsClient(t *testing.T) {
	require.Implements(t, (*sdk.LogsClient)(nil), &MockLogsClient{})
}
