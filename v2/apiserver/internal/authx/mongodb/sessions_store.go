package mongodb

import (
	"context"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// sessionsStore is a MongoDB-based implementation of the authx.SessionsStore
// interface.
type sessionsStore struct {
	collection mongodb.Collection
}

// NewSessionsStore returns a MongoDB-based implementation of the
// authx.SessionsStore interface.
func NewSessionsStore(database *mongo.Database) authx.SessionsStore {
	return &sessionsStore{
		collection: database.Collection("sessions"),
	}
}

func (s *sessionsStore) Create(
	ctx context.Context,
	session authx.Session,
) error {
	if _, err := s.collection.InsertOne(ctx, session); err != nil {
		return errors.Wrapf(err, "error inserting new session %q", session.ID)
	}
	return nil
}

func (s *sessionsStore) GetByHashedOAuth2State(
	ctx context.Context,
	hashedOAuth2State string,
) (authx.Session, error) {
	session := authx.Session{}
	res := s.collection.FindOne(
		ctx,
		bson.M{"hashedOAuth2State": hashedOAuth2State},
	)
	err := res.Decode(&session)
	if err == mongo.ErrNoDocuments {
		return session, &meta.ErrNotFound{
			Type: "Session",
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
) (authx.Session, error) {
	session := authx.Session{}
	res := s.collection.FindOne(ctx, bson.M{"hashedToken": hashedToken})
	err := res.Decode(&session)
	if err == mongo.ErrNoDocuments {
		return session, &meta.ErrNotFound{
			Type: "Session",
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
			Type: "Session",
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
			Type: "Session",
			ID:   id,
		}
	}
	return nil
}
