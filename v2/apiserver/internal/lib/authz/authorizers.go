package authz

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
)

// AuthorizeFn is the signature for any function that can, presumably, retrieve
// a principal from the provided Context and make an access control decision
// based on the principal having (or not having) at least on of the provided
// Roles. Implementations MUST return a *meta.ErrAuthorization error if the
// principal is not authorized.
type AuthorizeFn func(context.Context, ...Role) error

// AlwaysAuthorize is an implementation of the AuthorizeFn function signature
// that unconditionally passes authorization requests by returning nil. This is
// used only for testing purposes.
func AlwaysAuthorize(context.Context, ...Role) error {
	return nil
}

// NeverAuthorize is an implementation of the AuthorizeFn function signature
// that unconditionally fails authorization requests by returning a
// *meta.ErrAuthorization error. This is used only for testing purposes.
func NeverAuthorize(context.Context, ...Role) error {
	return &meta.ErrAuthorization{}
}
