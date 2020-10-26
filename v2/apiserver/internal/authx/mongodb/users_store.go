package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var caseInsensitiveIDCollation = &options.Collation{
	Locale:   "en",
	Strength: 1,
}

// usersStore is a MongoDB-based implementation of the authx.UsersStore
// interface.
type usersStore struct {
	collection mongodb.Collection
}

// NewUsersStore returns a MongoDB-based implementation of the authx.UsersStore
// interface.
func NewUsersStore(database *mongo.Database) (authx.UsersStore, error) {
	ctx, cancel :=
		context.WithTimeout(context.Background(), createIndexTimeout)
	defer cancel()
	unique := true
	collection := database.Collection("users")
	if _, err := collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"id": 1,
			},
			Options: &options.IndexOptions{
				Collation: caseInsensitiveIDCollation,
				Unique:    &unique,
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to users collection")
	}
	return &usersStore{
		collection: collection,
	}, nil
}

func (u *usersStore) Create(ctx context.Context, user authx.User) error {
	if _, err := u.collection.InsertOne(ctx, user); err != nil {
		if mongodb.IsDuplicateKeyError(err) {
			return &meta.ErrConflict{
				Type:   "User",
				ID:     user.ID,
				Reason: fmt.Sprintf("A user with the ID %q already exists.", user.ID),
			}
		}
		return errors.Wrapf(err, "error inserting new user %q", user.ID)
	}
	return nil
}

func (u *usersStore) List(
	ctx context.Context,
	opts meta.ListOptions,
) (authx.UserList, error) {
	users := authx.UserList{}

	criteria := bson.M{}
	if opts.Continue != "" {
		criteria["id"] = bson.M{"$gt": opts.Continue}
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"id": 1})
	findOptions.SetLimit(opts.Limit)
	cur, err := u.collection.Find(ctx, criteria, findOptions)
	if err != nil {
		return users, errors.Wrap(err, "error finding users")
	}
	if err := cur.All(ctx, &users.Items); err != nil {
		return users, errors.Wrap(err, "error decoding users")
	}

	if int64(len(users.Items)) == opts.Limit {
		continueID := users.Items[opts.Limit-1].ID
		criteria["id"] = bson.M{"$gt": continueID}
		remaining, err := u.collection.CountDocuments(ctx, criteria)
		if err != nil {
			return users, errors.Wrap(err, "error counting remaining users")
		}
		if remaining > 0 {
			users.Continue = continueID
			users.RemainingItemCount = remaining
		}
	}

	return users, nil
}

func (u *usersStore) Get(
	ctx context.Context,
	id string,
) (authx.User, error) {
	user := authx.User{}
	res := u.collection.FindOne(
		ctx,
		bson.M{"id": id},
		&options.FindOneOptions{
			Collation: caseInsensitiveIDCollation,
		},
	)
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

func (u *usersStore) Lock(ctx context.Context, id string) error {
	res, err := u.collection.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.M{
			"$set": bson.M{
				"locked": time.Now().UTC(),
			},
		},
		&options.UpdateOptions{
			Collation: caseInsensitiveIDCollation,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating user %q", id)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: "User",
			ID:   id,
		}
	}
	return nil
}

func (u *usersStore) Unlock(ctx context.Context, id string) error {
	res, err := u.collection.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.M{
			"$set": bson.M{
				"locked": nil,
			},
		},
		&options.UpdateOptions{
			Collation: caseInsensitiveIDCollation,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating user %q", id)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: "User",
			ID:   id,
		}
	}
	return nil
}
