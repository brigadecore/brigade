package authn

import (
	"testing"

	metaTesting "github.com/brigadecore/brigade/v2/apiserver/internal/meta/testing" // nolint: lll
)

func TestTokenMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, Token{}, "Token")
}
