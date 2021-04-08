package core

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	"github.com/pkg/errors"
)

// ProjectRoleAssignment represents the assignment of a ProjectRole to a
// principal such as a User or ServiceAccount.
type ProjectRoleAssignment struct {
	// Role assigns a ProjectRole to the specified principal.
	Role ProjectRole `json:"role" bson:"role"`
	// Principal specifies the principal to whom the ProjectRole is assigned.
	Principal authn.PrincipalReference `json:"principal" bson:"principal"`
}

// ProjectRoleAssignmentsService is the specialized interface for managing
// ProjectRoleAssignments. It's decoupled from underlying technology choices
// (e.g. data store, message bus, etc.) to keep business logic reusable and
// consistent while the underlying tech stack remains free to change.
type ProjectRoleAssignmentsService interface {
	// Grant grants the ProjectRole specified by the ProjectRoleAssignment to the
	// principal also specified by the ProjectRoleAssignment. If the specified
	// principal does not exist, implementations must return a *meta.ErrNotFound
	// error.
	Grant(ctx context.Context, roleAssignment ProjectRoleAssignment) error
	// Revoke revokes the ProjectRole specified by the ProjectRoleAssignment for
	// the principal also specified by the ProjectRoleAssignment. If the specified
	// principal does not exist, implementations must return a *meta.ErrNotFound
	// error.
	Revoke(ctx context.Context, roleAssignment ProjectRoleAssignment) error
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
	projectID := projectRoleAssignment.Role.ProjectID
	if err := p.projectAuthorize(ctx, RoleProjectAdmin(projectID)); err != nil {
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

	if projectRoleAssignment.Principal.Type == authn.PrincipalTypeUser {
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
		// creating a role assignment, we'd like to respect case. So we DON'T use
		// the ID from the inbound RoleAssignment-- which may have incorrect case.
		// Instead we replace it with the ID (with correct case) from the User we
		// found.
		projectRoleAssignment.Principal.ID = user.ID
	} else if projectRoleAssignment.Principal.Type == authn.PrincipalTypeServiceAccount { // nolint: lll
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
		ctx, projectRoleAssignment,
	); err != nil {
		return errors.Wrapf(
			err,
			"error granting project %q role %q to %s %q in store",
			projectID,
			projectRoleAssignment.Role.Name,
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
	projectID := projectRoleAssignment.Role.ProjectID
	if err := p.projectAuthorize(ctx, RoleProjectAdmin(projectID)); err != nil {
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

	if projectRoleAssignment.Principal.Type == authn.PrincipalTypeUser {
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
		// creating a role assignment, we'd like to respect case. So we DON'T use
		// the ID from the inbound RoleAssignment-- which may have incorrect case.
		// Instead we replace it with the ID (with correct case) from the User we
		// found.
		projectRoleAssignment.Principal.ID = user.ID
	} else if projectRoleAssignment.Principal.Type == authn.PrincipalTypeServiceAccount { // nolint: lll
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
			projectRoleAssignment.Role.Name,
			projectRoleAssignment.Principal.Type,
			projectRoleAssignment.Principal.ID,
		)
	}
	return nil
}

// ProjectRoleAssignmentsStore is an interface for components that implement
// ProjectRoleAssignment persistence concerns.
type ProjectRoleAssignmentsStore interface {
	// Grant the ProjectRole specified by the ProjectRoleAssignment to the
	// principal specified by the RoleAssignment.
	Grant(context.Context, ProjectRoleAssignment) error
	// Revoke the ProjectRole specified by the RoleAssignment for the principal
	// specified by the RoleAssignment.
	Revoke(context.Context, ProjectRoleAssignment) error
	// RevokeMany revokes all ProjectRoleAssignments for the given project ID.
	RevokeMany(ctx context.Context, projectID string) error
	// Exists returns a bool indicating whether the specified
	// ProjectRoleAssignment exists within the store. Implementations MUST also
	// return true if a ProjectRoleAssignment exists in the store that logically
	// "overlaps" the specified ProjectRoleAssignment. For instance, when seeking
	// to determine whether a ProjectRoleAssignment exists that endows some
	// principal P with Role X for Project Y, and such a ProjectRoleAssignment
	// does not exist, but one does that endows that principal P with Role X
	// having GLOBAL SCOPE (*), then true MUST be returned. Implementations MUST
	// also return an error if and only if anything goes wrong. i.e. Errors are
	// never used to communicate that the specified ProjectRoleAssignment does not
	// exist in the store. They are only used to convey an actual failure.
	Exists(context.Context, ProjectRoleAssignment) (bool, error)
}
