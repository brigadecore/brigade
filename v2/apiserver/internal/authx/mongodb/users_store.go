package mongodb

import (
	"context"
	"fmt"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// usersStore is a MongoDB-based implementation of the authx.UsersStore
// interface.
type usersStore struct {
	collection mongodb.Collection
}

// NewUsersStore returns a MongoDB-based implementation of the authx.UsersStore
// interface.
func NewUsersStore(database *mongo.Database) authx.UsersStore {
	return &usersStore{
		collection: database.Collection("users"),
	}
}

func (u *usersStore) Create(ctx context.Context, user authx.User) error {
	if _, err :=
		u.collection.InsertOne(ctx, user); err != nil {
		if writeException, ok := err.(mongo.WriteException); ok {
			if len(writeException.WriteErrors) == 1 &&
				writeException.WriteErrors[0].Code == 11000 {
				return &meta.ErrConflict{
					Type:   "User",
					ID:     user.ID,
					Reason: fmt.Sprintf("A user with the ID %q already exists.", user.ID),
				}
			}
		}
		return errors.Wrapf(err, "error inserting new user %q", user.ID)
	}
	return nil
}

func (u *usersStore) Get(
	ctx context.Context,
	id string,
) (authx.User, error) {
	user := authx.User{}
	res := u.collection.FindOne(ctx, bson.M{"id": id})
	err := res.Decode(&user)
	if err == mongo.ErrNoDocuments {
		return user, &meta.ErrNotFound{
			Type: "User",
			ID:   id,
		}
	}
	if err != nil {
		return user, errors.Wrapf(err, "error finding/decoding user %q", id)
	}
	return user, nil
}
