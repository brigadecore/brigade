package authx

import "testing"

func TestTokenMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, Token{}, "Token")
}
