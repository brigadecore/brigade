package authn

import (
	"context"
)

type principalContextKey struct{}

// ContextWithPrincipal returns a context.Context that has been augmented with
// the provided principal.
func ContextWithPrincipal(
	ctx context.Context,
	principal interface{},
) context.Context {
	return context.WithValue(
		ctx,
		principalContextKey{},
		principal,
	)
}

// PrincipalFromContext extracts a principal from the provided context.Context
// and returns it.
func PrincipalFromContext(ctx context.Context) interface{} {
	return ctx.Value(principalContextKey{})
}
