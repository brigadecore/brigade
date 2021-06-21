package core

import (
	"testing"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/stretchr/testify/require"
)

func TestMockLogsClient(t *testing.T) {
	require.Implements(t, (*core.LogsClient)(nil), &MockLogsClient{})
}
