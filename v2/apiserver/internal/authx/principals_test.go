package authx

import (
	"context"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
)

func TestContextWithPrincipal(t *testing.T) {
	testUser := User{
		ObjectMeta: meta.ObjectMeta{
			ID: "tony@starkindustries.com",
		},
	}
	ctx := ContextWithPrincipal(context.Background(), testUser)
	val := ctx.Value(principalContextKey{})
	require.Equal(t, testUser, val)
}

func TestPrincipalFromContext(t *testing.T) {
	testUser := User{
		ObjectMeta: meta.ObjectMeta{
			ID: "tony@starkindustries.com",
		},
	}
	ctx :=
		context.WithValue(context.Background(), principalContextKey{}, testUser)
	principal := PrincipalFromContext(ctx)
	require.Equal(t, testUser, principal)
}
