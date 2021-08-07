package mongodb

import (
	"context"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// sessionsStore is a MongoDB-based implementation of the api.SessionsStore
// interface.
type sessionsStore struct {
	collection mongodb.Collection
}

// NewSessionsStore returns a MongoDB-based implementation of the
// api.SessionsStore interface.
func NewSessionsStore(database *mongo.Database) (api.SessionsStore, error) {
	ctx, cancel :=
		context.WithTimeout(context.Background(), createIndexTimeout)
	defer cancel()
	unique := true
	// Set the default expiration value to 0 so that each record's 'expires'
	// field is used to determine actual expiration
	defaultExpiry := int32(0)
	collection := database.Collection("sessions")
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
			{
				Keys: bson.M{
					"expires": 1,
				},
				Options: &options.IndexOptions{
					ExpireAfterSeconds: &defaultExpiry,
				},
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to sessions collection")
	}
	return &sessionsStore{
		collection: collection,
	}, nil
}

func (s *sessionsStore) Create(
	ctx context.Context,
	session api.Session,
) error {
	if _, err := s.collection.InsertOne(ctx, session); err != nil {
		return errors.Wrapf(err, "error inserting new session %q", session.ID)
	}
	return nil
}

func (s *sessionsStore) GetByHashedOAuth2State(
	ctx context.Context,
	hashedOAuth2State string,
) (api.Session, error) {
	session := api.Session{}
	res := s.collection.FindOne(
		ctx,
		bson.M{"hashedOAuth2State": hashedOAuth2State},
	)
	err := res.Decode(&session)
	if err == mongo.ErrNoDocuments {
		return session, &meta.ErrNotFound{
			Type: api.SessionKind,
		}
	}
	if err != nil {
		return session, errors.Wrap(err, "error finding/decoding session")
	}
	return session, nil
}

func (s *sessionsStore) GetByHashedToken(
	ctx context.Context,
	hashedToken string,
) (api.Session, error) {
	session := api.Session{}
	res := s.collection.FindOne(ctx, bson.M{"hashedToken": hashedToken})
	err := res.Decode(&session)
	if err == mongo.ErrNoDocuments {
		return session, &meta.ErrNotFound{
			Type: api.SessionKind,
		}
	}
	if err != nil {
		return session, errors.Wrap(err, "error finding/decoding session")
	}
	return session, nil
}

func (s *sessionsStore) Authenticate(
	ctx context.Context,
	sessionID string,
	userID string,
	expires time.Time,
) error {
	res, err := s.collection.UpdateOne(
		ctx,
		bson.M{
			"id": sessionID,
		},
		bson.M{
			"$set": bson.M{
				"userID":        userID,
				"authenticated": time.Now().UTC(),
				"expires":       expires,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating session %q", sessionID)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: api.SessionKind,
			ID:   sessionID,
		}
	}
	return nil
}

func (s *sessionsStore) Delete(ctx context.Context, id string) error {
	res, err := s.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return errors.Wrapf(err, "error deleting session %q", id)
	}
	if res.DeletedCount == 0 {
		return &meta.ErrNotFound{
			Type: api.SessionKind,
			ID:   id,
		}
	}
	return nil
}

func (s *sessionsStore) DeleteByUser(ctx context.Context, userID string) error {
	if _, err :=
		s.collection.DeleteMany(ctx, bson.M{"userID": userID}); err != nil {
		return errors.Wrapf(err, "error deleting sessions for user %q", userID)
	}
	return nil
}
