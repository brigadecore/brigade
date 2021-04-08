package core

import (
	"context"
	"log"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	libAuthn "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authn"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
)

// ProjectAuthorizeFn is the signature for any function that can, presumably,
// retrieve a principal from the provided Context and make an access control
// decision based on the principal having (or not having) at least one of the
// provided ProjectRoles. Implementations MUST return a *meta.ErrAuthorization
// error if the principal is not authorized.
type ProjectAuthorizeFn func(context.Context, ...ProjectRole) error

// projectRoleHolder is an interface for any sort of security principal that can
// directly return its own ProjectRoles from a function call without making a
// database call.
type projectRoleHolder interface {
	ProjectRoles() []ProjectRole
}

// ProjectRoleAuthorizer is the public interface for the component returned by
// the NewAuthorizer function.
type ProjectRoleAuthorizer interface {
	// Authorize retrieves a principal from the provided Context and asserts that
	// it has at least one of the allowed ProjectRoles. If it does not,
	// implementations MUST return a *meta.ErrAuthorization error.
	Authorize(ctx context.Context, allowedRoles ...ProjectRole) error
}

// projectRoleAuthorizer is a component that can authorize a request based on
// ProjectRoles.
type projectRoleAuthorizer struct {
	projectRoleAssignmentsStore ProjectRoleAssignmentsStore
}

// NewProjectRoleAuthorizer returns a component that can authorize a request
// based on ProjectRoles.
func NewProjectRoleAuthorizer(
	projectRoleAssignmentsStore ProjectRoleAssignmentsStore,
) ProjectRoleAuthorizer {
	return &projectRoleAuthorizer{
		projectRoleAssignmentsStore: projectRoleAssignmentsStore,
	}
}

func (p *projectRoleAuthorizer) Authorize(
	ctx context.Context,
	allowedRoles ...ProjectRole,
) error {
	principal := libAuthn.PrincipalFromContext(ctx)
	if principal == nil {
		return &meta.ErrAuthorization{}
	}
	roleAssignment := ProjectRoleAssignment{}
	switch p := principal.(type) {
	case projectRoleHolder: // Any principal with hard-coded ProjectRoles
		for _, allowedRole := range allowedRoles {
			for _, principalRole := range p.ProjectRoles() {
				if principalRole.Matches(allowedRole) {
					return nil
				}
			}
		}
		return &meta.ErrAuthorization{}
	case *authn.User:
		roleAssignment.Principal = authn.PrincipalReference{
			Type: authn.PrincipalTypeUser,
			ID:   p.ID,
		}
	case *authn.ServiceAccount:
		roleAssignment.Principal = authn.PrincipalReference{
			Type: authn.PrincipalTypeServiceAccount,
			ID:   p.ID,
		}
	default: // What kind of principal is this??? This shouldn't happen.
		return &meta.ErrAuthorization{}
	}
	// We only get here if the principal was a User or ServiceAccount
	for _, roleAssignment.Role = range allowedRoles {
		if exists, err := p.projectRoleAssignmentsStore.Exists(
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
