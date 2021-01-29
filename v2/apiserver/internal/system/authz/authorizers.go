package authz

import (
	"context"
	"log"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	"github.com/brigadecore/brigade/v2/apiserver/internal/authz"
	libAuthn "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authn"
	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
)

// roleHolder is an interface for any sort of security principal that can
// directly return its own Roles from a function call without making a database
// call.
type roleHolder interface {
	Roles() []libAuthz.Role
}

// Authorizer is the public interface for the component returned by the
// NewAuthorizer function.
type Authorizer interface {
	// Authorize retrieves a principal from the provided Context and asserts that
	// it has at least one of the allowed Roles. If it does not, implementations
	// MUST return a *meta.ErrAuthorization error.
	Authorize(ctx context.Context, allowedRoles ...libAuthz.Role) error
}

// authorizer is a component that can authorize a request.
type authorizer struct {
	roleAssignmentsStore authz.RoleAssignmentsStore
}

// NewAuthorizer returns a component that can authorize a request.
func NewAuthorizer(roleAssignmentsStore authz.RoleAssignmentsStore) Authorizer {
	return &authorizer{
		roleAssignmentsStore: roleAssignmentsStore,
	}
}

func (a *authorizer) Authorize(
	ctx context.Context,
	allowedRoles ...libAuthz.Role,
) error {
	principal := libAuthn.PrincipalFromContext(ctx)
	if principal == nil {
		return &meta.ErrAuthorization{}
	}
	roleAssignment := authz.RoleAssignment{}
	switch p := principal.(type) {
	case roleHolder: // Any principal with hard-coded roles
		for _, allowedRole := range allowedRoles {
			for _, principalRole := range p.Roles() {
				if principalRole.Matches(allowedRole) {
					return nil
				}
			}
		}
		return &meta.ErrAuthorization{}
	case *authn.User:
		roleAssignment.Principal = authz.PrincipalReference{
			Type: authz.PrincipalTypeUser,
			ID:   p.ID,
		}
	case *authn.ServiceAccount:
		roleAssignment.Principal = authz.PrincipalReference{
			Type: authz.PrincipalTypeServiceAccount,
			ID:   p.ID,
		}
	default: // What kind of principal is this??? This shouldn't happen.
		return &meta.ErrAuthorization{}
	}
	// We only get here if the principal was a User or ServiceAccount
	for _, roleAssignment.Role = range allowedRoles {
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
	}
	return &meta.ErrAuthorization{}
}
