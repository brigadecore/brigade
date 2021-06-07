package testing

import (
	"testing"

	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/stretchr/testify/require"
)

func TestMockAPIClient(t *testing.T) {
	require.Implements(t, (*sdk.APIClient)(nil), &MockAPIClient{})
}
