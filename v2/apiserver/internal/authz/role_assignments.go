package authz

import (
	"context"
	"encoding/json"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/brigadecore/brigade/v2/apiserver/internal/system"
	"github.com/pkg/errors"
)

const (
	// RoleAssignmentListKind represents the canonical RoleAssignmentList kind
	// string
	RoleAssignmentListKind = "RoleAssignmentList"

	// PrincipalTypeServiceAccount represents a principal that is a
	// ServiceAccount.
	PrincipalTypeServiceAccount libAuthz.PrincipalType = "SERVICE_ACCOUNT"
	// PrincipalTypeUser represents a principal that is a User.
	PrincipalTypeUser libAuthz.PrincipalType = "USER"
)

// RoleAssignmentList is an ordered and pageable list of system-level
// RoleAssignments.
type RoleAssignmentList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of RoleAssignments.
	Items []libAuthz.RoleAssignment `json:"items,omitempty"`
}

// MarshalJSON amends RoleAssignmentList instances with type metadata.
func (r RoleAssignmentList) MarshalJSON() ([]byte, error) {
	type Alias RoleAssignmentList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       RoleAssignmentListKind,
			},
			Alias: (Alias)(r),
		},
	)
}

// RoleAssignmentsSelector represents useful filter criteria when selecting
// multiple RoleAssignments for API group operations like list.
type RoleAssignmentsSelector struct {
	// Principal specifies that only RoleAssignments for the specified principal
	// should be selected.
	Principal *libAuthz.PrincipalReference
	// Role specifies that only RoleAssignments for the specified Role should be
	// selected.
	Role libAuthz.Role
}

// RoleAssignmentsService is the specialized interface for managing
// RoleAssignments. It's decoupled from underlying technology choices (e.g. data
// store, message bus, etc.) to keep business logic reusable and consistent
// while the underlying tech stack remains free to change.
type RoleAssignmentsService interface {
	// Grant grants the Role specified by the RoleAssignment to the principal also
	// specified by the RoleAssignment. If the specified principal does not exist,
	// implementations must return a *meta.ErrNotFound error.
	Grant(ctx context.Context, roleAssignment libAuthz.RoleAssignment) error

	// List returns a RoleAssignmentsList, with its Items (RoleAssignments)
	// ordered by principal type, principalID, role, and scope. Criteria for which
	// RoleAssignments should be retrieved can be specified using the
	// RoleAssignmentsSelector parameter.
	List(
		context.Context,
		RoleAssignmentsSelector,
		meta.ListOptions,
	) (RoleAssignmentList, error)

	// Revoke revokes the Role specified by the RoleAssignment for the principal
	// also specified by the RoleAssignment. If the specified principal does not
	// exist, implementations must return a *meta.ErrNotFound error.
	Revoke(ctx context.Context, roleAssignment libAuthz.RoleAssignment) error
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
	roleAssignment libAuthz.RoleAssignment,
) error {
	if err := r.authorize(ctx, system.RoleAdmin, ""); err != nil {
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
			roleAssignment.Role,
			roleAssignment.Scope,
			roleAssignment.Principal.Type,
			roleAssignment.Principal.ID,
		)
	}

	return nil
}

func (r *roleAssignmentsService) List(
	ctx context.Context,
	selector RoleAssignmentsSelector,
	opts meta.ListOptions,
) (RoleAssignmentList, error) {
	if err := r.authorize(ctx, system.RoleReader, ""); err != nil {
		return RoleAssignmentList{}, err
	}

	if opts.Limit == 0 {
		opts.Limit = 20
	}

	roleAssignments, err := r.roleAssignmentsStore.List(ctx, selector, opts)
	return roleAssignments,
		errors.Wrap(err, "error retrieving role assignments from store")
}

func (r *roleAssignmentsService) Revoke(
	ctx context.Context,
	roleAssignment libAuthz.RoleAssignment,
) error {
	if err := r.authorize(ctx, system.RoleAdmin, ""); err != nil {
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
			roleAssignment.Role,
			roleAssignment.Scope,
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
	Grant(context.Context, libAuthz.RoleAssignment) error
	// List returns a RoleAssignmentsList, with its Items (system-level
	// RoleAssignments) ordered by principal type, principalID, role name, and
	// scope. Criteria for which RoleAssignments should be retrieved can be
	// specified using the RoleAssignmentsSelector parameter.
	List(
		context.Context,
		RoleAssignmentsSelector,
		meta.ListOptions,
	) (RoleAssignmentList, error)
	// Revoke the role specified by the RoleAssignment for the principal specified
	// by the RoleAssignment.
	Revoke(context.Context, libAuthz.RoleAssignment) error
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
	Exists(context.Context, libAuthz.RoleAssignment) (bool, error)
}
