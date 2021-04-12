package authz

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/system"
	"github.com/pkg/errors"
)

const (
	// PrincipalTypeServiceAccount represents a principal that is a
	// ServiceAccount.
	PrincipalTypeServiceAccount libAuthz.PrincipalType = "SERVICE_ACCOUNT"
	// PrincipalTypeUser represents a principal that is a User.
	PrincipalTypeUser libAuthz.PrincipalType = "USER"
)

// RoleAssignment represents the assignment of a Role to a principal such as a
// User or ServiceAccount.
type RoleAssignment struct {
	// Role assigns a Role to the specified principal.
	Role libAuthz.Role `json:"role" bson:"role"`
	// Principal specifies the principal to whom the Role is assigned.
	Principal libAuthz.PrincipalReference `json:"principal" bson:"principal"`
}

// RoleAssignmentsService is the specialized interface for managing
// RoleAssignments. It's decoupled from underlying technology choices (e.g. data
// store, message bus, etc.) to keep business logic reusable and consistent
// while the underlying tech stack remains free to change.
type RoleAssignmentsService interface {
	// Grant grants the Role specified by the RoleAssignment to the principal also
	// specified by the RoleAssignment. If the specified principal does not exist,
	// implementations must return a *meta.ErrNotFound error.
	Grant(ctx context.Context, roleAssignment RoleAssignment) error

	// Revoke revokes the Role specified by the RoleAssignment for the principal
	// also specified by the RoleAssignment. If the specified principal does not
	// exist, implementations must return a *meta.ErrNotFound error.
	Revoke(ctx context.Context, roleAssignment RoleAssignment) error
}

type roleAssignmentsService struct {
	authorize            libAuthz.AuthorizeFn
	usersStore           authn.UsersStore
	serviceAccountsStore authn.ServiceAccountsStore
	roleAssignmentsStore RoleAssignmentsStore
}

// NewRoleAssignmentsService returns a specialized interface for managing
// RoleAssignments.
func NewRoleAssignmentsService(
	authorizeFn libAuthz.AuthorizeFn,
	usersStore authn.UsersStore,
	serviceAccountsStore authn.ServiceAccountsStore,
	roleAssignmentsStore RoleAssignmentsStore,
) RoleAssignmentsService {
	return &roleAssignmentsService{
		authorize:            authorizeFn,
		usersStore:           usersStore,
		serviceAccountsStore: serviceAccountsStore,
		roleAssignmentsStore: roleAssignmentsStore,
	}
}

func (r *roleAssignmentsService) Grant(
	ctx context.Context,
	roleAssignment RoleAssignment,
) error {
	if err := r.authorize(ctx, system.RoleAdmin()); err != nil {
		return err
	}

	switch roleAssignment.Principal.Type {
	case PrincipalTypeUser:
		// Make sure the User exists
		user, err := r.usersStore.Get(ctx, roleAssignment.Principal.ID)
		if err != nil {
			return errors.Wrapf(
				err,
				"error retrieving user %q from store",
				roleAssignment.Principal.ID,
			)
		}
		// From an end-user's perspective, User IDs are case insensitive, but when
		// creating a role assignment, we'd like to respect case. So we DON'T use
		// the ID from the inbound RoleAssignment-- which may have incorrect case.
		// Instead we replace it with the ID (with correct case) from the User we
		// found.
		roleAssignment.Principal.ID = user.ID
	case PrincipalTypeServiceAccount:
		// Make sure the ServiceAccount exists
		if _, err :=
			r.serviceAccountsStore.Get(ctx, roleAssignment.Principal.ID); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving service account %q from store",
				roleAssignment.Principal.ID,
			)
		}
	default:
		return nil
	}

	// Give them the Role
	if err := r.roleAssignmentsStore.Grant(ctx, roleAssignment); err != nil {
		return errors.Wrapf(
			err,
			"error granting role %q with scope %q to %s %q in store",
			roleAssignment.Role.Name,
			roleAssignment.Role.Scope,
			roleAssignment.Principal.Type,
			roleAssignment.Principal.ID,
		)
	}

	return nil
}

func (r *roleAssignmentsService) Revoke(
	ctx context.Context,
	roleAssignment RoleAssignment,
) error {
	if err := r.authorize(ctx, system.RoleAdmin()); err != nil {
		return err
	}

	switch roleAssignment.Principal.Type {
	case PrincipalTypeUser:
		// Make sure the User exists
		user, err := r.usersStore.Get(ctx, roleAssignment.Principal.ID)
		if err != nil {
			return errors.Wrapf(
				err,
				"error retrieving user %q from store",
				roleAssignment.Principal.ID,
			)
		}
		// From an end-user's perspective, User IDs are case insensitive, but when
		// creating a role assignment, we'd like to respect case. So we DON'T use
		// the ID from the inbound RoleAssignment-- which may have incorrect case.
		// Instead we replace it with the ID (with correct case) from the User we
		// found.
		roleAssignment.Principal.ID = user.ID
	case PrincipalTypeServiceAccount:
		// Make sure the ServiceAccount exists
		if _, err :=
			r.serviceAccountsStore.Get(ctx, roleAssignment.Principal.ID); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving service account %q from store",
				roleAssignment.Principal.ID,
			)
		}
	default:
		return nil
	}

	// Revoke the Role
	if err := r.roleAssignmentsStore.Revoke(ctx, roleAssignment); err != nil {
		return errors.Wrapf(
			err,
			"error revoking role %q with scope %q for %s %q in store",
			roleAssignment.Role.Name,
			roleAssignment.Role.Scope,
			roleAssignment.Principal.Type,
			roleAssignment.Principal.ID,
		)
	}
	return nil
}

// RoleAssignmentsStore is an interface for components that implement
// RoleAssignment persistence concerns.
type RoleAssignmentsStore interface {
	// Grant the role specified by the RoleAssignment to the principal specified
	// by the RoleAssignment.
	Grant(context.Context, RoleAssignment) error
	// Revoke the role specified by the RoleAssignment for the principal specified
	// by the RoleAssignment.
	Revoke(context.Context, RoleAssignment) error
	// RevokeMany revokes all RoleAssignments that share ALL properties of the
	// specified RoleAssignment. Properties left unspecified are ignored, i.e.
	// not factored into the match.
	//
	// Example -- revoking all project-level RoleAssignments for a given Project:
	//
	//   err := p.roleAssignmentsStore.RevokeMany(
	// 	  ctx,
	// 	  authz.RoleAssignment{
	// 		  Role: libAuthz.Role{
	// 			  Type:  RoleTypeProject,
	// 			  Scope: projectID,
	// 		  },
	// 	  },
	//   )
	RevokeMany(ctx context.Context, roleAssignment RoleAssignment) error

	// Exists returns a bool indicating whether the specified RoleAssignment
	// exists within the store. Implementations MUST also return true if a
	// RoleAssignment exists in the store that logically "overlaps" the specified
	// RoleAssignment. For instance, when seeking to determine whether a
	// RoleAssignment exists that endows some principal P with Role X having scope
	// Y, and such a RoleAssignment does not exist, but one does that endows that
	// principal P with Role X having GLOBAL SCOPE (*), then true MUST be
	// returned. Implementations MUST also return an error if and only if anything
	// goes wrong. i.e. Errors are never used to communicate that the specified
	// RoleAssignment does not exist in the store. They are only used to convey an
	// actual failure.
	Exists(context.Context, RoleAssignment) (bool, error)
}
