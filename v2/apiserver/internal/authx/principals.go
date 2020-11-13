package authx

import "context"

var (
	// Root is a singleton that represents Brigade's "root" user.
	Root = &root{}
	// Scheduler is a singleton that represents Brigade's scheduler component.
	Scheduler = &scheduler{}
)

// Principal is an interface for any sort of security principal (human user,
// service account, etc.)
type Principal interface{}

// root is an implementation of the Principal interface for the "root" user.
type root struct{}

type scheduler struct{}

type principalContextKey struct{}

// ContextWithPrincipal returns a context.Context that has been augmented with
// the provided Principal.
func ContextWithPrincipal(
	ctx context.Context,
	principal Principal,
) context.Context {
	return context.WithValue(
		ctx,
		principalContextKey{},
		principal,
	)
}

// PrincipalFromContext extracts a Principal from the provided context.Context
// and returns it.
func PrincipalFromContext(ctx context.Context) Principal {
	return ctx.Value(principalContextKey{}).(Principal)
}
