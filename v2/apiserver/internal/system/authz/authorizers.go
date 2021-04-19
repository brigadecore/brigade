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

// roleAssignmentsHolder is an interface for any sort of security principal that
// can directly return its own RoleAssignments from a function call without
// making a database call.
type roleAssignmentsHolder interface {
	RoleAssignments() []libAuthz.RoleAssignment
}

// Authorizer is the public interface for the component returned by the
// NewAuthorizer function.
type Authorizer interface {
	// Authorize retrieves a principal from the provided Context and asserts that
	// it has the specified Role with the specified scope. If it does not,
	// implementations MUST return a *meta.ErrAuthorization error.
	Authorize(ctx context.Context, roles libAuthz.Role, scope string) error
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
	role libAuthz.Role,
	scope string,
) error {
	principal := libAuthn.PrincipalFromContext(ctx)
	if principal == nil {
		return &meta.ErrAuthorization{}
	}
	roleAssignment := libAuthz.RoleAssignment{
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
	case *authn.User:
		roleAssignment.Principal = libAuthz.PrincipalReference{
			Type: authz.PrincipalTypeUser,
			ID:   p.ID,
		}
	case *authn.ServiceAccount:
		roleAssignment.Principal = libAuthz.PrincipalReference{
			Type: authz.PrincipalTypeServiceAccount,
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
