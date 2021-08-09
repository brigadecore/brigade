package api

import (
	"context"
)

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

// PrincipalReference is a reference to any sort of security principal (human
// user, service account, etc.)
type PrincipalReference struct {
	// Type qualifies what kind of principal is referenced by the ID field-- for
	// instance, a User or a ServiceAccount.
	Type PrincipalType `json:"type,omitempty"`
	// ID references a principal. The Type qualifies what type of principal that
	// is-- for instance, a User or a ServiceAccount.
	ID string `json:"id,omitempty"`
}

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
