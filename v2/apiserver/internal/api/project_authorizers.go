package api

import (
	"context"
	"log"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
)

// ProjectAuthorizeFn is the signature for any function that can, presumably,
// retrieve a principal from the provided Context and make an access control
// decision based on the principal having (or not having) the specified Role for
// the specified Project. Implementations MUST return a *meta.ErrAuthorization
// error if the principal is not authorized.
type ProjectAuthorizeFn func(
	ctx context.Context,
	projectID string,
	role Role,
) error

// projectRoleAssignmentsHolder is an interface for any sort of security
// principal that can directly return its own ProjectRoleAssignments from a
// function call without making a database call.
type projectRoleAssignmentsHolder interface {
	ProjectRoleAssignments() []ProjectRoleAssignment
}

// ProjectAuthorizer is the public interface for the component returned by the
// NewProjectAuthorizer function.
type ProjectAuthorizer interface {
	// Authorize retrieves a principal from the provided Context and asserts that
	// it has the specified Role for the specified Project. If it does not,
	// implementations MUST return a *meta.ErrAuthorization error.
	Authorize(ctx context.Context, projectID string, role Role) error
}

// projectAuthorizer is a component that can authorize a request.
type projectAuthorizer struct {
	projectRoleAssignmentsStore ProjectRoleAssignmentsStore
}

// NewProjectAuthorizer returns a component that can authorize a request.
func NewProjectAuthorizer(
	projectRoleAssignmentsStore ProjectRoleAssignmentsStore,
) ProjectAuthorizer {
	return &projectAuthorizer{
		projectRoleAssignmentsStore: projectRoleAssignmentsStore,
	}
}
func (p *projectAuthorizer) Authorize(
	ctx context.Context,
	projectID string,
	role Role,
) error {
	principal := PrincipalFromContext(ctx)
	if principal == nil {
		return &meta.ErrAuthorization{}
	}
	projectRoleAssignment := ProjectRoleAssignment{
		ProjectID: projectID,
		Role:      role,
	}
	switch p := principal.(type) {
	case projectRoleAssignmentsHolder:
		// A principal with hard-coded RoleAssignments
		for _, projectRoleAssignment = range p.ProjectRoleAssignments() {
			if projectRoleAssignment.Matches(projectID, role) {
				return nil
			}
		}
		return &meta.ErrAuthorization{}
	case *User:
		projectRoleAssignment.Principal = PrincipalReference{
			Type: PrincipalTypeUser,
			ID:   p.ID,
		}
	case *ServiceAccount:
		projectRoleAssignment.Principal = PrincipalReference{
			Type: PrincipalTypeServiceAccount,
			ID:   p.ID,
		}
	default:
		// This case might occur for a specialized principal like the scheduler or
		// observer that is neither a User or ServiceAccount nor implements the
		// roleAssignmentHolder interface.
		return &meta.ErrAuthorization{}
	}
	// We only get here if the principal was a User or ServiceAccount
	if exists, err := p.projectRoleAssignmentsStore.Exists(
		ctx,
		projectRoleAssignment,
	); err != nil {
		// We encountered an unexpected error when looking for a
		// ProjectRoleAssignment in the store. We're going to treat this as an authz
		// failure, but we're also going to log it for good measure.
		log.Println(err)
		return &meta.ErrAuthorization{}
	} else if exists {
		return nil
	}
	return &meta.ErrAuthorization{}
}
