package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// serviceAccountsStore is a MongoDB-based implementation of the
// api.ServiceAccountsStore interface.
type serviceAccountsStore struct {
	collection mongodb.Collection
}

// NewServiceAccountsStore returns a MongoDB-based implementation of the
// api.ServiceAccountsStore interface.
func NewServiceAccountsStore(
	database *mongo.Database,
) (api.ServiceAccountsStore, error) {
	ctx, cancel :=
		context.WithTimeout(context.Background(), createIndexTimeout)
	defer cancel()
	unique := true
	collection := database.Collection("service-accounts")
	if _, err := collection.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.M{
					"id": 1,
				},
				Options: &options.IndexOptions{
					Unique: &unique,
				},
			},
			{
				Keys: bson.M{
					"hashedToken": 1,
				},
				Options: &options.IndexOptions{
					Unique: &unique,
				},
			},
		},
	); err != nil {
		return nil, errors.Wrap(
			err,
			"error adding indexes to service accounts collection",
		)
	}
	return &serviceAccountsStore{
		collection: collection,
	}, nil
}

func (s *serviceAccountsStore) Create(
	ctx context.Context,
	serviceAccount api.ServiceAccount,
) error {
	if _, err := s.collection.InsertOne(ctx, serviceAccount); err != nil {
		if mongodb.IsDuplicateKeyError(err) {
			return &meta.ErrConflict{
				Type: api.ServiceAccountKind,
				ID:   serviceAccount.ID,
				Reason: fmt.Sprintf(
					"A service account with the ID %q already exists.",
					serviceAccount.ID,
				),
			}
		}
		return errors.Wrapf(
			err,
			"error inserting new service account %q",
			serviceAccount.ID,
		)
	}
	return nil
}

func (s *serviceAccountsStore) List(
	ctx context.Context,
	opts meta.ListOptions,
) (api.ServiceAccountList, error) {
	serviceAccounts := api.ServiceAccountList{}

	criteria := bson.M{}
	if opts.Continue != "" {
		criteria["id"] = bson.M{"$gt": opts.Continue}
	}

	findOptions := options.Find()
	findOptions.SetSort(
		// bson.D preserves order so we use this wherever we sort so that if
		// additional sort criteria are added in the future, they will be applied
		// in the specified order.
		bson.D{
			{Key: "id", Value: 1},
		},
	)
	findOptions.SetLimit(opts.Limit)
	cur, err := s.collection.Find(ctx, criteria, findOptions)
	if err != nil {
		return serviceAccounts,
			errors.Wrap(err, "error finding service accounts")
	}
	if err := cur.All(ctx, &serviceAccounts.Items); err != nil {
		return serviceAccounts,
			errors.Wrap(err, "error decoding service accounts")
	}

	if int64(len(serviceAccounts.Items)) == opts.Limit {
		continueID := serviceAccounts.Items[opts.Limit-1].ID
		criteria["id"] = bson.M{"$gt": continueID}
		remaining, err := s.collection.CountDocuments(ctx, criteria)
		if err != nil {
			return serviceAccounts,
				errors.Wrap(err, "error counting remaining service accounts")
		}
		if remaining > 0 {
			serviceAccounts.Continue = continueID
			serviceAccounts.RemainingItemCount = remaining
		}
	}

	return serviceAccounts, nil
}

func (s *serviceAccountsStore) Get(
	ctx context.Context,
	id string,
) (api.ServiceAccount, error) {
	serviceAccount := api.ServiceAccount{}
	res := s.collection.FindOne(ctx, bson.M{"id": id})
	err := res.Decode(&serviceAccount)
	if err == mongo.ErrNoDocuments {
		return serviceAccount, &meta.ErrNotFound{
			Type: api.ServiceAccountKind,
			ID:   id,
		}
	}
	if err != nil {
		return serviceAccount,
			errors.Wrapf(res.Err(), "error finding/decoding service account %q", id)
	}
	return serviceAccount, nil
}

func (s *serviceAccountsStore) GetByHashedToken(
	ctx context.Context,
	hashedToken string,
) (api.ServiceAccount, error) {
	serviceAccount := api.ServiceAccount{}
	res :=
		s.collection.FindOne(ctx, bson.M{"hashedToken": hashedToken})
	err := res.Decode(&serviceAccount)
	if err == mongo.ErrNoDocuments {
		return serviceAccount, &meta.ErrNotFound{
			Type: api.ServiceAccountKind,
		}
	}
	if err != nil {
		return serviceAccount,
			errors.Wrap(err, "error finding/decoding service account by hashed token")
	}
	return serviceAccount, nil
}

func (s *serviceAccountsStore) Lock(ctx context.Context, id string) error {
	res, err := s.collection.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.M{
			"$set": bson.M{
				"locked": time.Now().UTC(),
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating service account %q", id)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: api.ServiceAccountKind,
			ID:   id,
		}
	}
	// Note, there are no sessions to delete because service accounts use
	// non-expiring, sessionless tokens.
	return nil
}

func (s *serviceAccountsStore) Unlock(
	ctx context.Context,
	id string,
	newHashedToken string,
) error {
	res, err := s.collection.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.M{
			"$set": bson.M{
				"locked":      nil,
				"hashedToken": newHashedToken,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating service account %q", id)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: api.ServiceAccountKind,
			ID:   id,
		}
	}
	return nil
}
