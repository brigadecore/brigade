package api

import (
	"context"
	"encoding/json"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

// UserKind represents the canonical User kind string
const UserKind = "User"

// User represents a (human) Brigade user.
type User struct {
	// ObjectMeta encapsulates User metadata.
	meta.ObjectMeta `json:"metadata" bson:",inline"`
	// Name is the given name and surname of the User.
	Name string `json:"name" bson:"name"`
	// Locked indicates when the User has been locked out of the system by an
	// administrator. If this field's value is nil, the User is not locked.
	Locked *time.Time `json:"locked" bson:"locked"`
}

// MarshalJSON amends User instances with type metadata.
func (u User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       UserKind,
			},
			Alias: (Alias)(u),
		},
	)
}

// UsersServiceConfig encapsulates several configuration options for the
// UsersService.
type UsersServiceConfig struct {
	// ThirdPartyAuthEnabled indicates whether authentication using a third-party
	// identity provider is supported by the Sessions service.
	ThirdPartyAuthEnabled bool
}

// UsersService is the specialized interface for managing Users. It's decoupled
// from underlying technology choices (e.g. data store) to keep business logic
// reusable and consistent while the underlying tech stack remains free to
// change.
type UsersService interface {
	// List returns a UserList.
	List(context.Context, meta.ListOptions) (UserList, error)
	// Get retrieves a single User specified by their identifier.
	Get(context.Context, string) (User, error)

	// Lock removes access to the API for a single User specified by their
	// identifier.
	Lock(context.Context, string) error
	// Unlock restores access to the API for a single User specified by their
	// identifier.
	Unlock(context.Context, string) error

	// Delete removes a single User specified by their identifier.
	Delete(context.Context, string) error
}

// usersService is an implementation of the UsersService interface.
type usersService struct {
	authorize                   AuthorizeFn
	usersStore                  UsersStore
	sessionsStore               SessionsStore
	roleAssignmentsStore        RoleAssignmentsStore
	projectRoleAssignmentsStore ProjectRoleAssignmentsStore
	config                      UsersServiceConfig
}

// NewUsersService returns a specialized interface for managing Users.
func NewUsersService(
	authorizeFn AuthorizeFn,
	usersStore UsersStore,
	sessionsStore SessionsStore,
	roleAssignmentsStore RoleAssignmentsStore,
	projectRoleAssignmentsStore ProjectRoleAssignmentsStore,
	config UsersServiceConfig,
) UsersService {
	return &usersService{
		authorize:                   authorizeFn,
		usersStore:                  usersStore,
		sessionsStore:               sessionsStore,
		roleAssignmentsStore:        roleAssignmentsStore,
		projectRoleAssignmentsStore: projectRoleAssignmentsStore,
		config:                      config,
	}
}

func (u *usersService) List(
	ctx context.Context,
	opts meta.ListOptions,
) (UserList, error) {
	if err := u.authorize(ctx, RoleReader, ""); err != nil {
		return UserList{}, err
	}

	if !u.config.ThirdPartyAuthEnabled {
		return UserList{}, errUserManagementDisabled()
	}

	if opts.Limit == 0 {
		opts.Limit = 20
	}
	users, err := u.usersStore.List(ctx, opts)
	if err != nil {
		return users, errors.Wrap(err, "error retrieving users from store")
	}
	return users, nil
}

func (u *usersService) Get(ctx context.Context, id string) (User, error) {
	if err := u.authorize(ctx, RoleReader, ""); err != nil {
		return User{}, err
	}

	if !u.config.ThirdPartyAuthEnabled {
		return User{}, errUserManagementDisabled()
	}

	user, err := u.usersStore.Get(ctx, id)
	if err != nil {
		return user, errors.Wrapf(
			err,
			"error retrieving user %q from store",
			id,
		)
	}
	return user, nil
}

func (u *usersService) Lock(ctx context.Context, id string) error {
	if err := u.authorize(ctx, RoleAdmin, ""); err != nil {
		return err
	}

	if !u.config.ThirdPartyAuthEnabled {
		return errUserManagementDisabled()
	}

	if err := u.usersStore.Lock(ctx, id); err != nil {
		return errors.Wrapf(err, "error locking user %q in store", id)
	}
	if err := u.sessionsStore.DeleteByUser(ctx, id); err != nil {
		return errors.Wrapf(err, "error deleting user %q sessions from store", id)
	}
	return nil
}

func (u *usersService) Unlock(ctx context.Context, id string) error {
	if err := u.authorize(ctx, RoleAdmin, ""); err != nil {
		return err
	}

	if !u.config.ThirdPartyAuthEnabled {
		return errUserManagementDisabled()
	}

	if err := u.usersStore.Unlock(ctx, id); err != nil {
		return errors.Wrapf(err, "error unlocking user %q in store", id)
	}
	return nil
}

func (u *usersService) Delete(ctx context.Context, id string) error {
	if err := u.authorize(ctx, RoleAdmin, ""); err != nil {
		return err
	}

	principalReference := PrincipalReference{
		Type: PrincipalTypeUser,
		ID:   id,
	}
	if err := u.roleAssignmentsStore.RevokeByPrincipal(
		ctx,
		principalReference,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting user %q role assignments from store",
			id,
		)
	}
	if err := u.projectRoleAssignmentsStore.RevokeByPrincipal(
		ctx,
		principalReference,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting user %q project role assignments from store",
			id,
		)
	}
	if err := u.usersStore.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "error deleting user %q from store", id)
	}
	if err := u.sessionsStore.DeleteByUser(ctx, id); err != nil {
		return errors.Wrapf(err, "error deleting user %q sessions from store", id)
	}
	return nil
}

// UsersStore is an interface for User persistence operations.
type UsersStore interface {
	// Create persists a new User in the underlying data store. If a User having
	// the same ID already exists, implementations MUST return a *meta.ErrConflict
	// error.
	Create(context.Context, User) error
	// List retrieves a UserList from the underlying data store, with its Items
	// (Users) ordered by ID.
	List(context.Context, meta.ListOptions) (UserList, error)
	// Get retrieves a single User from the underlying data store. Implementations
	// MUST use a case insensitive query for this operation. If the specified User
	// does not exist, implementations MUST return a *meta.ErrNotFound error.
	Get(context.Context, string) (User, error)

	// Lock updates the specified User in the underlying data store to reflect
	// that it has been locked out of the system. Implementations MUST use a case
	// insensitive update statement for this operation. If the specified User does
	// not exist, implementations MUST return a *meta.ErrNotFound error.
	Lock(context.Context, string) error
	// Unlock updates the specified User in the underlying data store to reflect
	// that its system access (after presumably having been revoked) has been
	// restored. Implementations MUST use a case insensitive update statement for
	// this operation. If the specified User does not exist, implementations MUST
	// return a *meta.ErrNotFound error.
	Unlock(ctx context.Context, id string) error

	// Delete deletes the specified User. If no User having the given identifier
	// is found, implementations MUST return a *meta.ErrNotFound error.
	Delete(context.Context, string) error
}

func errUserManagementDisabled() *meta.ErrNotSupported {
	return &meta.ErrNotSupported{
		Details: "Authentication using a third party identity provider is not " +
			"supported by this server. User management functions are unavailable " +
			"when authentication via a third party identity provider is not enabled.",
	}
}
