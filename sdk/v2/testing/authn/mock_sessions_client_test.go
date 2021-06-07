package authn

import (
	"testing"

	"github.com/brigadecore/brigade/sdk/v2/authn"
	"github.com/stretchr/testify/require"
)

func TestMockSessionsClient(t *testing.T) {
	require.Implements(t, (*authn.SessionsClient)(nil), &MockSessionsClient{})
}
