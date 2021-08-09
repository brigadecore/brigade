package api

import (
	"context"
	"log"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
)

// AuthorizeFn is the signature for any function that can, presumably, retrieve
// a principal from the provided Context and make an access control decision
// based on the principal having (or not having) the specified Role with the
// specified scope. Implementations MUST return a *meta.ErrAuthorization error
// if the principal is not authorized.
type AuthorizeFn func(ctx context.Context, role Role, scope string) error

// roleAssignmentsHolder is an interface for any sort of security principal that
// can directly return its own RoleAssignments from a function call without
// making a database call.
type roleAssignmentsHolder interface {
	RoleAssignments() []RoleAssignment
}

// Authorizer is the public interface for the component returned by the
// NewAuthorizer function.
type Authorizer interface {
	// Authorize retrieves a principal from the provided Context and asserts that
	// it has the specified Role with the specified scope. If it does not,
	// implementations MUST return a *meta.ErrAuthorization error.
	Authorize(ctx context.Context, roles Role, scope string) error
}

// authorizer is a component that can authorize a request.
type authorizer struct {
	roleAssignmentsStore RoleAssignmentsStore
}

// NewAuthorizer returns a component that can authorize a request.
func NewAuthorizer(roleAssignmentsStore RoleAssignmentsStore) Authorizer {
	return &authorizer{
		roleAssignmentsStore: roleAssignmentsStore,
	}
}

func (a *authorizer) Authorize(
	ctx context.Context,
	role Role,
	scope string,
) error {
	principal := PrincipalFromContext(ctx)
	if principal == nil {
		return &meta.ErrAuthorization{}
	}
	roleAssignment := RoleAssignment{
		Role:  role,
		Scope: scope,
	}
	switch p := principal.(type) {
	case roleAssignmentsHolder: // Any principal with hard-coded RoleAssignments
		for _, principalRoleAssignment := range p.RoleAssignments() {
			if principalRoleAssignment.Matches(role, scope) {
				return nil
			}
		}
		return &meta.ErrAuthorization{}
	case *User:
		roleAssignment.Principal = PrincipalReference{
			Type: PrincipalTypeUser,
			ID:   p.ID,
		}
	case *ServiceAccount:
		roleAssignment.Principal = PrincipalReference{
			Type: PrincipalTypeServiceAccount,
			ID:   p.ID,
		}
	default: // What kind of principal is this??? This shouldn't happen.
		return &meta.ErrAuthorization{}
	}
	// We only get here if the principal was a User or ServiceAccount
	if exists, err := a.roleAssignmentsStore.Exists(
		ctx,
		roleAssignment,
	); err != nil {
		// We encountered an unexpected error when looking for a RoleAssignment
		// in the store. We're going to treat this as an authz failure, but we're
		// also going to log it for good measure.
		log.Println(err)
		return &meta.ErrAuthorization{}
	} else if exists {
		return nil
	}
	return &meta.ErrAuthorization{}
}
