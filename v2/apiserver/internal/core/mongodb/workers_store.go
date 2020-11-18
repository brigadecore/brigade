package mongodb

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// workersStore is a MongoDB-based implementation of the core.WorkersStore
// interface.
type workersStore struct {
	collection mongodb.Collection
}

// NewWorkersStore returns a MongoDB-based implementation of the
// core.WorkersStore interface.
func NewWorkersStore(database *mongo.Database) (core.WorkersStore, error) {
	return &workersStore{
		collection: database.Collection("events"),
	}, nil
}

func (w *workersStore) UpdateStatus(
	ctx context.Context,
	eventID string,
	status core.WorkerStatus,
) error {
	res, err := w.collection.UpdateOne(
		ctx,
		bson.M{"id": eventID},
		bson.M{
			"$set": bson.M{
				"worker.status": status,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error updating status of event %q worker",
			eventID,
		)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: "Event",
			ID:   eventID,
		}
	}
	return nil
}
