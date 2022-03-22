package api

import (
	"context"
	"encoding/json"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

const (
	// ProjectRoleAssignmentKind represents the canonical ProjectRoleAssignment
	// kind string
	ProjectRoleAssignmentKind = "ProjectRoleAssignment"

	// ProjectRoleAssignmentListKind represents the canonical
	// ProjectRoleAssignmentList kind string
	ProjectRoleAssignmentListKind = "ProjectRoleAssignmentList"
)

// ProjectRoleAssignment represents the assignment of a project-level Role to a
// principal such as a User or ServiceAccount.
type ProjectRoleAssignment struct {
	// ProjectID qualifies the scope of the Role.
	ProjectID string `json:"projectID" bson:"projectID,omitempty"`
	// Role assigns a Role to the specified principal.
	Role Role `json:"role" bson:"role"`
	// Principal specifies the principal to whom the Role is assigned.
	Principal PrincipalReference `json:"principal" bson:"principal"`
}

// Matches determines if this ProjectRoleAssignment matches the projectID and
// role arguments.
func (p ProjectRoleAssignment) Matches(
	projectID string,
	role Role,
) bool {
	return p.Role == role &&
		(p.ProjectID == projectID || p.ProjectID == ProjectRoleScopeGlobal)
}

// MarshalJSON amends ProjectRoleAssignment instances with type metadata.
func (p ProjectRoleAssignment) MarshalJSON() ([]byte, error) {
	type Alias ProjectRoleAssignment
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       ProjectRoleAssignmentKind,
			},
			Alias: (Alias)(p),
		},
	)
}

// ProjectRoleAssignmentsSelector represents useful filter criteria when
// selecting multiple ProjectRoleAssignments for API group operations like list.
type ProjectRoleAssignmentsSelector struct {
	// Principal specifies that only ProjectRoleAssignments for the specified
	// principal should be selected.
	Principal *PrincipalReference
	// ProjectID specifies that only ProjectRoleAssignments for the specified
	// Project should be selected.
	ProjectID string
	// Role specifies that only ProjectRoleAssignments for the specified
	// Role should be selected.
	Role Role
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

	// List returns a ProjectRoleAssignmentList, with its Items
	// (ProjectRoleAssignments) ordered by projectID, principal type, principalID,
	// and role. Criteria for which ProjectRoleAssignments should be retrieved can
	// be specified using the ProjectRoleAssignmentsSelector parameter.
	List(
		context.Context,
		ProjectRoleAssignmentsSelector,
		meta.ListOptions,
	) (ProjectRoleAssignmentList, error)

	// Revoke revokes the project-level Role specified by the
	// ProjectRoleAssignment for the principal also specified by the
	// ProjectRoleAssignment. If the specified principal does not exist,
	// implementations must return a *meta.ErrNotFound error.
	Revoke(context.Context, ProjectRoleAssignment) error
}

type projectRoleAssignmentsService struct {
	authorize                   AuthorizeFn
	projectAuthorize            ProjectAuthorizeFn
	projectsStore               ProjectsStore
	usersStore                  UsersStore
	serviceAccountsStore        ServiceAccountsStore
	projectRoleAssignmentsStore ProjectRoleAssignmentsStore
}

// NewProjectRoleAssignmentsService returns a specialized interface for managing
// project-level RoleAssignments.
func NewProjectRoleAssignmentsService(
	authorize AuthorizeFn,
	projectAuthorize ProjectAuthorizeFn,
	projectsStore ProjectsStore,
	usersStore UsersStore,
	serviceAccountsStore ServiceAccountsStore,
	projectRoleAssignmentsStore ProjectRoleAssignmentsStore,
) ProjectRoleAssignmentsService {
	return &projectRoleAssignmentsService{
		authorize:                   authorize,
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

	if projectRoleAssignment.Principal.Type == PrincipalTypeUser {
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
	} else if projectRoleAssignment.Principal.Type == PrincipalTypeServiceAccount { // nolint: lll
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

func (p *projectRoleAssignmentsService) List(
	ctx context.Context,
	selector ProjectRoleAssignmentsSelector,
	opts meta.ListOptions,
) (ProjectRoleAssignmentList, error) {
	if err := p.authorize(ctx, RoleReader, ""); err != nil {
		return ProjectRoleAssignmentList{}, err
	}

	if opts.Limit == 0 {
		opts.Limit = 20
	}

	roleAssignments, err :=
		p.projectRoleAssignmentsStore.List(ctx, selector, opts)
	return roleAssignments,
		errors.Wrap(err, "error retrieving role assignments from store")
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

	if projectRoleAssignment.Principal.Type == PrincipalTypeUser {
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
	} else if projectRoleAssignment.Principal.Type == PrincipalTypeServiceAccount { // nolint: lll
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
	// List returns a ProjectRoleAssignmentsList, with its Items
	// (ProjectRoleAssignments) ordered by projectID, principal type, principalID,
	// and role. Criteria for which RoleAssignments should be retrieved can be
	// specified using the RoleAssignmentsSelector parameter.
	List(
		context.Context,
		ProjectRoleAssignmentsSelector,
		meta.ListOptions,
	) (ProjectRoleAssignmentList, error)
	// Revoke the Project specified by the ProjectRoleAssignment for the principal
	// specified by the ProjectRoleAssignment.
	Revoke(context.Context, ProjectRoleAssignment) error
	// RevokeByProjectID revokes all ProjectRoleAssignments for the specified
	// Project.
	RevokeByProjectID(ctx context.Context, projectID string) error
	// RevokeByPrincipal revokes all project roles for the principal specified by
	// the PrincipalReference.
	RevokeByPrincipal(context.Context, PrincipalReference) error
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
