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

// workersStore is a MongoDB-based implementation of the api.WorkersStore
// interface.
type workersStore struct {
	collection mongodb.Collection
}

// NewWorkersStore returns a MongoDB-based implementation of the
// api.WorkersStore interface.
func NewWorkersStore(database *mongo.Database) (api.WorkersStore, error) {
	return &workersStore{
		collection: database.Collection("events"),
	}, nil
}

func (w *workersStore) UpdateStatus(
	ctx context.Context,
	eventID string,
	status api.WorkerStatus,
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
			Type: api.EventKind,
			ID:   eventID,
		}
	}
	return nil
}

func (w *workersStore) UpdateHashedToken(
	ctx context.Context,
	eventID string,
	hashedToken string,
) error {
	res, err := w.collection.UpdateOne(
		ctx,
		bson.M{"id": eventID},
		bson.M{
			"$set": bson.M{
				"worker.hashedToken": hashedToken,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error updating event %q worker hashed token",
			eventID,
		)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: api.EventKind,
			ID:   eventID,
		}
	}
	return nil
}

func (w *workersStore) Timeout(
	ctx context.Context,
	eventID string,
) error {
	timedOutTime := time.Now().UTC()

	res, err := w.collection.UpdateOne(
		ctx,
		bson.M{
			"id": eventID,
			"worker.status.phase": bson.M{
				"$in": []api.WorkerPhase{
					api.WorkerPhaseStarting,
					api.WorkerPhaseRunning,
				},
			},
		},
		bson.M{
			"$set": bson.M{
				"worker.status.ended":                           timedOutTime,
				"worker.status.phase":                           api.WorkerPhaseTimedOut, // nolint: lll
				"worker.jobs.$[pending].status.phase":           api.JobPhaseCanceled,
				"worker.jobs.$[startingOrRunning].status.phase": api.JobPhaseAborted,
			},
		},
		&options.UpdateOptions{
			ArrayFilters: &options.ArrayFilters{
				Filters: []interface{}{
					bson.M{
						"pending.status.phase": api.JobPhasePending,
					},
					bson.M{
						"startingOrRunning.status.phase": bson.M{
							"$in": []api.JobPhase{
								api.JobPhaseStarting,
								api.JobPhaseRunning,
							},
						},
					},
				},
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
		return &meta.ErrConflict{
			Type: api.EventKind,
			ID:   eventID,
			Reason: fmt.Sprintf(
				"Event %q was not timed out "+
					"because it was not in a starting or running state.",
				eventID,
			),
		}
	}

	return nil
}
