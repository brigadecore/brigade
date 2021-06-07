package authn

import (
	"testing"

	"github.com/brigadecore/brigade/sdk/v2/authn"
	"github.com/stretchr/testify/require"
)

func TestMockUsersClient(t *testing.T) {
	require.Implements(t, (*authn.UsersClient)(nil), &MockUsersClient{})
}
