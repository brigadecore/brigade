package core

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	"github.com/brigadecore/brigade/v2/apiserver/internal/authz"
	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/pkg/errors"
)

// ProjectRoleAssignment represents the assignment of a project-level Role to a
// principal such as a User or ServiceAccount.
type ProjectRoleAssignment struct {
	// ProjectID qualifies the scope of the Role.
	ProjectID string `json:"projectID,omitempty" bson:"projectID,omitempty"`
	// Role assigns a Role to the specified principal.
	Role libAuthz.Role `json:"role" bson:"role"`
	// Principal specifies the principal to whom the Role is assigned.
	Principal libAuthz.PrincipalReference `json:"principal" bson:"principal"`
}

// Matches determines if this ProjectRoleAssignment matches the projectID and
// role arguments.
func (p ProjectRoleAssignment) Matches(
	projectID string,
	role libAuthz.Role,
) bool {
	return p.Role == role &&
		(p.ProjectID == projectID || p.ProjectID == ProjectRoleScopeGlobal)
}

// ProjectRoleAssignmentsService is the specialized interface for managing
// ProjectRoleAssignments. It's decoupled from underlying technology choices
// (e.g. data store, message bus, etc.) to keep business logic reusable and
// consistent while the underlying tech stack remains free to change.
type ProjectRoleAssignmentsService interface {
	// Grant grants the project-level Role specified by the ProjectRoleAssignment
	// to the principal also specified by the ProjectRoleAssignment. If the
	// specified Project or principal does not exist, implementations must return
	// a *meta.ErrNotFound error.
	Grant(context.Context, ProjectRoleAssignment) error

	// Revoke revokes the project-level Role specified by the
	// ProjectRoleAssignment for the principal also specified by the
	// ProjectRoleAssignment. If the specified principal does not exist,
	// implementations must return a *meta.ErrNotFound error.
	Revoke(context.Context, ProjectRoleAssignment) error
}

type projectRoleAssignmentsService struct {
	projectAuthorize            ProjectAuthorizeFn
	projectsStore               ProjectsStore
	usersStore                  authn.UsersStore
	serviceAccountsStore        authn.ServiceAccountsStore
	projectRoleAssignmentsStore ProjectRoleAssignmentsStore
}

// NewProjectRoleAssignmentsService returns a specialized interface for managing
// project-level RoleAssignments.
func NewProjectRoleAssignmentsService(
	projectAuthorize ProjectAuthorizeFn,
	projectsStore ProjectsStore,
	usersStore authn.UsersStore,
	serviceAccountsStore authn.ServiceAccountsStore,
	projectRoleAssignmentsStore ProjectRoleAssignmentsStore,
) ProjectRoleAssignmentsService {
	return &projectRoleAssignmentsService{
		projectAuthorize:            projectAuthorize,
		projectsStore:               projectsStore,
		usersStore:                  usersStore,
		serviceAccountsStore:        serviceAccountsStore,
		projectRoleAssignmentsStore: projectRoleAssignmentsStore,
	}
}

func (p *projectRoleAssignmentsService) Grant(
	ctx context.Context,
	projectRoleAssignment ProjectRoleAssignment,
) error {
	projectID := projectRoleAssignment.ProjectID
	if err := p.projectAuthorize(ctx, projectID, RoleProjectAdmin); err != nil {
		return err
	}

	// Make sure the project exists
	_, err := p.projectsStore.Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}

	if projectRoleAssignment.Principal.Type == authz.PrincipalTypeUser {
		// Make sure the User exists
		user, err := p.usersStore.Get(ctx, projectRoleAssignment.Principal.ID)
		if err != nil {
			return errors.Wrapf(
				err,
				"error retrieving user %q from store",
				projectRoleAssignment.Principal.ID,
			)
		}
		// From an end-user's perspective, User IDs are case insensitive, but when
		// creating a ProjectRoleAssignment, we'd like to respect case. So we DON'T
		// use the ID from the inbound ProjectRoleAssignment-- which may have
		// incorrect case. Instead we replace it with the ID (with correct case)
		// from the User we found.
		projectRoleAssignment.Principal.ID = user.ID
	} else if projectRoleAssignment.Principal.Type == authz.PrincipalTypeServiceAccount { // nolint: lll
		// Make sure the ServiceAccount exists
		if _, err := p.serviceAccountsStore.Get(
			ctx,
			projectRoleAssignment.Principal.ID,
		); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving service account %q from store",
				projectRoleAssignment.Principal.ID,
			)
		}
	} else {
		return nil
	}

	// Give them the Role
	if err := p.projectRoleAssignmentsStore.Grant(
		ctx,
		projectRoleAssignment,
	); err != nil {
		return errors.Wrapf(
			err,
			"error granting project %q role %q to %s %q in store",
			projectID,
			projectRoleAssignment.Role,
			projectRoleAssignment.Principal.Type,
			projectRoleAssignment.Principal.ID,
		)
	}

	return nil
}

func (p *projectRoleAssignmentsService) Revoke(
	ctx context.Context,
	projectRoleAssignment ProjectRoleAssignment,
) error {
	projectID := projectRoleAssignment.ProjectID
	if err := p.projectAuthorize(ctx, projectID, RoleProjectAdmin); err != nil {
		return err
	}

	// Make sure the project exists
	_, err := p.projectsStore.Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}

	if projectRoleAssignment.Principal.Type == authz.PrincipalTypeUser {
		// Make sure the User exists
		user, err := p.usersStore.Get(ctx, projectRoleAssignment.Principal.ID)
		if err != nil {
			return errors.Wrapf(
				err,
				"error retrieving user %q from store",
				projectRoleAssignment.Principal.ID,
			)
		}
		// From an end-user's perspective, User IDs are case insensitive, but when
		// creating a ProjectRoleAssignment, we'd like to respect case. So we DON'T
		// use the ID from the inbound ProjectRoleAssignment-- which may have
		// incorrect case. Instead we replace it with the ID (with correct case)
		// from the User we found.
		projectRoleAssignment.Principal.ID = user.ID
	} else if projectRoleAssignment.Principal.Type == authz.PrincipalTypeServiceAccount { // nolint: lll
		// Make sure the ServiceAccount exists
		if _, err := p.serviceAccountsStore.Get(
			ctx,
			projectRoleAssignment.Principal.ID,
		); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving service account %q from store",
				projectRoleAssignment.Principal.ID,
			)
		}
	} else {
		return nil
	}

	// Revoke the Role
	if err := p.projectRoleAssignmentsStore.Revoke(
		ctx,
		projectRoleAssignment,
	); err != nil {
		return errors.Wrapf(
			err,
			"error revoking project %q role %q for %s %q in store",
			projectID,
			projectRoleAssignment.Role,
			projectRoleAssignment.Principal.Type,
			projectRoleAssignment.Principal.ID,
		)
	}
	return nil
}

// ProjectRoleAssignmentsStore is an interface for components that implement
// ProjectRoleAssignment persistence concerns.
type ProjectRoleAssignmentsStore interface {
	// Grant the project-level Role specified by the ProjectRoleAssignment to the
	// principal specified by the ProjectRoleAssignment.
	Grant(context.Context, ProjectRoleAssignment) error
	// Revoke the Project specified by the ProjectRoleAssignment for the principal
	// specified by the ProjectRoleAssignment.
	Revoke(context.Context, ProjectRoleAssignment) error
	// RevokeByProjectID revokes all ProjectRoleAssignments for the specified
	// Project.
	RevokeByProjectID(ctx context.Context, projectID string) error

	// Exists returns a bool indicating whether the specified
	// ProjectRoleAssignment exists within the store. Implementations MUST also
	// return true if a ProjectRoleAssignment exists in the store that logically
	// "overlaps" the specified ProjectRoleAssignment. For instance, when seeking
	// to determine whether a ProjectRoleAssignment exists that endows some
	// principal P with Role X for Project Y, and such a ProjectRoleAssignment
	// does not exist, but one does that endows that principal P with Role X
	// having GLOBAL PROJECT SCOPE (*), then true MUST be returned.
	// Implementations MUST also return an error if and only if anything goes
	// wrong. i.e. Errors are never used to communicate that the specified
	// ProjectRoleAssignment does not exist in the store. They are only used to
	// convey an actual failure.
	Exists(context.Context, ProjectRoleAssignment) (bool, error)
}
