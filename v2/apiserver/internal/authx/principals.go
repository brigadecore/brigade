package authx

import "context"

// PrincipalType is a type whose values can be used to disambiguate one type of
// principal from another. For instance, when assigning a Role to a principal
// via a RoleAssignment, a PrincipalType field is used to indicate whether the
// value of the PrincipalID field reflects a User ID or a ServiceAccount ID.
type PrincipalType string

const (
	// PrincipalTypeServiceAccount represents a principal that is a
	// ServiceAccount.
	PrincipalTypeServiceAccount PrincipalType = "SERVICE_ACCOUNT"
	// PrincipalTypeUser represents a principal that is a User.
	PrincipalTypeUser PrincipalType = "USER"
)

var (
	// Root is a singleton that represents Brigade's "root" user.
	Root = &root{}
	// Scheduler is a singleton that represents Brigade's scheduler component.
	Scheduler = &scheduler{}
	// Observer is a singleton that represents Brigade's observer component.
	Observer = &observer{}
)

// Principal is an interface for any sort of security principal (human user,
// service account, etc.)
type Principal interface{}

// root is an implementation of the Principal interface for the "root" user.
type root struct{}

type scheduler struct{}

type observer struct{}

type worker struct {
	eventID string
}

func Worker(eventID string) Principal {
	return &worker{
		eventID: eventID,
	}
}

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
