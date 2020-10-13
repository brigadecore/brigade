package authx

import (
	"context"
	"encoding/json"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

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
				Kind:       "User",
			},
			Alias: (Alias)(u),
		},
	)
}

// UsersService is the specialized interface for managing Users. It's decoupled
// from underlying technology choices (e.g. data store) to keep business logic
// reusable and consistent while the underlying tech stack remains free to
// change.
type UsersService interface {
	// Get retrieves a single User specified by their identifier.
	Get(context.Context, string) (User, error)
}

// usersService is an implementation of the UsersService interface.
type usersService struct {
	store UsersStore
}

// NewUsersService returns a specialized interface for managing Users.
func NewUsersService(store UsersStore) UsersService {
	return &usersService{
		store: store,
	}
}

func (u *usersService) Get(ctx context.Context, id string) (User, error) {
	user, err := u.store.Get(ctx, id)
	if err != nil {
		return user, errors.Wrapf(
			err,
			"error retrieving user %q from store",
			id,
		)
	}
	return user, nil
}

// UsersStore is an interface for User persistence operations.
type UsersStore interface {
	// Create stores the provided User. If a User having the same ID already
	// exists, implementations MUST return a *meta.ErrConflict error.
	Create(context.Context, User) error
	// Get retrieves an existing User by ID. If no such User exists,
	// implementations MUST return a *meta.ErrNotFound error.
	Get(context.Context, string) (User, error)
}
