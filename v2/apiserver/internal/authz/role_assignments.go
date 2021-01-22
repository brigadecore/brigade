package authz

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/pkg/errors"
)

// PrincipalType is a type whose values can be used to disambiguate one type of
// principal from another. For instance, when assigning a Role to a principal
// via a RoleAssignment, a PrincipalType field is used to indicate whether the
// value of the PrincipalID field reflects a User ID or a ServiceAccount ID.
type PrincipalType string

const (
	// PrincipalTypeServiceAccount represents a principal that is a
	// ServiceAccount.
	PrincipalTypeServiceAccount PrincipalType = "SERVICE_ACCOUNT"
	// PrincipalTypeUser represents a principal that is a User.
	PrincipalTypeUser PrincipalType = "USER"
)

// PrincipalReference is a reference to any sort of security principal (human
// user, service account, etc.)
type PrincipalReference struct {
	// Type qualifies what kind of principal is referenced by the ID field-- for
	// instance, a User or a ServiceAccount.
	Type PrincipalType `json:"type,omitempty" bson:"type,omitempty"`
	// ID references a principal. The Type qualifies what type of principal that
	// is-- for instance, a User or a ServiceAccount.
	ID string `json:"id,omitempty" bson:"id,omitempty"`
}

// RoleAssignment represents the assignment of a Role to a principal such as a
// User or ServiceAccount.
type RoleAssignment struct {
	// Role assigns a Role to the specified principal.
	Role libAuthz.Role `json:"role" bson:"role"`
	// Principal specifies the principal to whom the Role is assigned.
	Principal PrincipalReference `json:"principal" bson:"principal"`
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
	usersStore           authn.UsersStore
	serviceAccountsStore authn.ServiceAccountsStore
	roleAssignmentsStore RoleAssignmentsStore
}

// NewRoleAssignmentsService returns a specialized interface for managing
// RoleAssignments.
func NewRoleAssignmentsService(
	usersStore authn.UsersStore,
	serviceAccountsStore authn.ServiceAccountsStore,
	roleAssignmentsStore RoleAssignmentsStore,
) RoleAssignmentsService {
	return &roleAssignmentsService{
		usersStore:           usersStore,
		serviceAccountsStore: serviceAccountsStore,
		roleAssignmentsStore: roleAssignmentsStore,
	}
}

func (r *roleAssignmentsService) Grant(
	ctx context.Context,
	roleAssignment RoleAssignment,
) error {
	switch roleAssignment.Principal.Type {
	case PrincipalTypeUser:
		// Make sure the User exists
		if _, err :=
			r.usersStore.Get(ctx, roleAssignment.Principal.ID); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving user %q from store",
				roleAssignment.Principal.ID,
			)
		}
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
	switch roleAssignment.Principal.Type {
	case PrincipalTypeUser:
		// Make sure the User exists
		if _, err :=
			r.usersStore.Get(ctx, roleAssignment.Principal.ID); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving user %q from store",
				roleAssignment.Principal.ID,
			)
		}
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
}
