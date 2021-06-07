package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
			Type: core.EventKind,
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
			"id":                  eventID,
			"worker.status.phase": core.WorkerPhasePending,
			"deleted": bson.M{
				"$exists": false, // Don't grab logically deleted events
			},
		},
		bson.M{
			"$set": bson.M{
				"canceled":            timedOutTime,
				"worker.status.phase": core.WorkerPhaseTimedOut,
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
	if res.MatchedCount == 1 {
		return nil
	}

	res, err = w.collection.UpdateOne(
		ctx,
		bson.M{
			"id": eventID,
			"worker.status.phase": bson.M{
				"$in": []core.WorkerPhase{
					core.WorkerPhaseStarting,
					core.WorkerPhaseRunning,
				},
			},
		},
		bson.M{
			"$set": bson.M{
				"worker.status.phase":                           core.WorkerPhaseTimedOut, // nolint: lll
				"worker.jobs.$[pending].status.phase":           core.JobPhaseCanceled,
				"worker.jobs.$[startingOrRunning].status.phase": core.JobPhaseAborted,
			},
		},
		&options.UpdateOptions{
			ArrayFilters: &options.ArrayFilters{
				Filters: []interface{}{
					bson.M{
						"pending.status.phase": core.JobPhasePending,
					},
					bson.M{
						"startingOrRunning.status.phase": bson.M{
							"$in": []core.JobPhase{
								core.JobPhaseStarting,
								core.JobPhaseRunning,
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
			Type: core.EventKind,
			ID:   eventID,
			Reason: fmt.Sprintf(
				"Event %q was not timed out "+
					"because it was already in a terminal state.",
				eventID,
			),
		}
	}

	return nil
}
