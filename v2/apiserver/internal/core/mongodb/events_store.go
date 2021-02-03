package mongodb

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// eventsStore is a MongoDB-based implementation of the core.EventsStore
// interface.
type eventsStore struct {
	collection mongodb.Collection
}

// NewEventsStore returns a MongoDB-based implementation of the core.EventsStore
// interface.
func NewEventsStore(database *mongo.Database) (core.EventsStore, error) {
	ctx, cancel :=
		context.WithTimeout(context.Background(), createIndexTimeout)
	defer cancel()
	unique := true
	collection := database.Collection("events")
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
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to events collection")
	}
	return &eventsStore{
		collection: collection,
	}, nil
}

func (e *eventsStore) Create(ctx context.Context, event core.Event) error {
	if _, err := e.collection.InsertOne(ctx, event); err != nil {
		return errors.Wrapf(err, "error inserting new event %q", event.ID)
	}
	return nil
}

func (e *eventsStore) List(
	ctx context.Context,
	selector core.EventsSelector,
	opts meta.ListOptions,
) (core.EventList, error) {
	events := core.EventList{}

	criteria := bson.M{
		"worker.status.phase": bson.M{
			"$in": selector.WorkerPhases,
		},
		"deleted": bson.M{
			"$exists": false, // Don't grab logically deleted events
		},
	}
	if selector.ProjectID != "" {
		criteria["projectID"] = selector.ProjectID
	}
	if opts.Continue != "" {
		tokens := strings.Split(opts.Continue, ":")
		if len(tokens) != 2 {
			return events, errors.New("error parsing continue time")
		}
		continueTimeNano, err := strconv.ParseInt(tokens[0], 10, 64)
		if err != nil {
			return events, errors.Wrap(err, "error parsing continue time")
		}
		continueTime := time.Unix(0, continueTimeNano).UTC()
		continueID := tokens[1]
		criteria["$or"] = []bson.M{
			{"created": continueTime, "id": bson.M{"$gt": continueID}},
			{"created": bson.M{"$lt": continueTime}},
		}
	}

	findOptions := options.Find()
	findOptions.SetSort(
		// bson.D preserves order, and we want to sort by created date/time FIRST
		// and id SECOND
		bson.D{
			{Key: "created", Value: -1},
			{Key: "id", Value: 1},
		},
	)
	findOptions.SetLimit(opts.Limit)
	cur, err := e.collection.Find(ctx, criteria, findOptions)
	if err != nil {
		return events, errors.Wrap(err, "error finding events")
	}
	if err := cur.All(ctx, &events.Items); err != nil {
		return events, errors.Wrap(err, "error decoding events")
	}

	if int64(len(events.Items)) == opts.Limit {
		continueTime := events.Items[opts.Limit-1].Created
		continueID := events.Items[opts.Limit-1].ID
		criteria["$or"] = []bson.M{
			{"created": continueTime, "id": bson.M{"$gt": continueID}},
			{"created": bson.M{"$lt": continueTime}},
		}
		remaining, err := e.collection.CountDocuments(ctx, criteria)
		if err != nil {
			return events, errors.Wrap(err, "error counting remaining events")
		}
		if remaining > 0 {
			events.Continue =
				fmt.Sprintf("%d:%s", continueTime.UnixNano(), continueID)
			events.RemainingItemCount = remaining
		}
	}

	return events, nil
}

func (e *eventsStore) Get(
	ctx context.Context,
	id string,
) (core.Event, error) {
	event := core.Event{}
	res := e.collection.FindOne(ctx, bson.M{"id": id})
	err := res.Decode(&event)
	if err == mongo.ErrNoDocuments {
		return event, &meta.ErrNotFound{
			Type: "Event",
			ID:   id,
		}
	}
	if err != nil {
		return event,
			errors.Wrapf(res.Err(), "error finding/decoding event %q", id)
	}
	return event, nil
}

func (e *eventsStore) GetByHashedWorkerToken(
	ctx context.Context,
	hashedWorkerToken string,
) (core.Event, error) {
	event := core.Event{}
	res := e.collection.FindOne(
		ctx,
		bson.M{
			"worker.hashedToken": hashedWorkerToken,
		},
	)
	err := res.Decode(&event)
	if res.Err() == mongo.ErrNoDocuments {
		return event, &meta.ErrNotFound{
			Type: "Event",
		}
	}
	if err != nil {
		return event, errors.Wrap(err, "error finding/decoding event")
	}
	return event, nil
}

func (e *eventsStore) Cancel(ctx context.Context, id string) error {
	now := time.Now().UTC()

	res, err := e.collection.UpdateOne(
		ctx,
		bson.M{
			"id":                  id,
			"worker.status.phase": core.WorkerPhasePending,
		},
		bson.M{
			"$set": bson.M{
				"canceled":            now,
				"worker.status.phase": core.WorkerPhaseCanceled,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating status of event %q worker", id)
	}
	if res.MatchedCount == 1 {
		return nil
	}

	res, err = e.collection.UpdateOne(
		ctx,
		bson.M{
			"id": id,
			"worker.status.phase": bson.M{
				"$in": []core.WorkerPhase{
					core.WorkerPhaseStarting,
					core.WorkerPhaseRunning,
				},
			},
		},
		bson.M{
			"$set": bson.M{
				"canceled":            now,
				"worker.status.phase": core.WorkerPhaseAborted,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating status of event %q worker", id)
	}

	if res.MatchedCount == 0 {
		return &meta.ErrConflict{
			Type: "Event",
			ID:   id,
			Reason: fmt.Sprintf(
				"Event %q was not canceled because it was already in a terminal state.",
				id,
			),
		}
	}

	return nil
}

func (e *eventsStore) CancelMany(
	ctx context.Context,
	selector core.EventsSelector,
) (core.EventList, error) {
	events := core.EventList{}
	// It only makes sense to cancel events that are in a pending, starting, or
	// running state. We can ignore anything else.
	var cancelPending bool
	var cancelStarting bool
	var cancelRunning bool
	for _, workerPhase := range selector.WorkerPhases {
		if workerPhase == core.WorkerPhasePending {
			cancelPending = true
		}
		if workerPhase == core.WorkerPhaseStarting {
			cancelStarting = true
		}
		if workerPhase == core.WorkerPhaseRunning {
			cancelRunning = true
		}
	}

	// Bail if we're not canceling pending, starting, or running events
	if !cancelPending && !cancelStarting && !cancelRunning {
		return events, nil
	}

	// The MongoDB driver for Go doesn't expose findAndModify(), which could be
	// used to select events and cancel them at the same time. As a workaround,
	// we'll cancel first, then select events based on cancellation time.

	cancellationTime := time.Now().UTC()

	criteria := bson.M{
		"projectID": selector.ProjectID,
	}

	if cancelPending {
		criteria["worker.status.phase"] = core.WorkerPhasePending
		if _, err := e.collection.UpdateMany(
			ctx,
			criteria,
			bson.M{
				"$set": bson.M{
					"canceled":            cancellationTime,
					"worker.status.phase": core.WorkerPhaseCanceled,
				},
			},
		); err != nil {
			return events, errors.Wrap(err, "error updating events")
		}
	}

	if cancelStarting {
		criteria["worker.status.phase"] = core.WorkerPhaseStarting
		if _, err := e.collection.UpdateMany(
			ctx,
			criteria,
			bson.M{
				"$set": bson.M{
					"canceled":            cancellationTime,
					"worker.status.phase": core.WorkerPhaseAborted,
				},
			},
		); err != nil {
			return events, errors.Wrap(err, "error updating events")
		}
	}

	if cancelRunning {
		criteria["worker.status.phase"] = core.WorkerPhaseRunning
		if _, err := e.collection.UpdateMany(
			ctx,
			criteria,
			bson.M{
				"$set": bson.M{
					"canceled":            cancellationTime,
					"worker.status.phase": core.WorkerPhaseAborted,
				},
			},
		); err != nil {
			return events, errors.Wrap(err, "error updating events")
		}
	}

	delete(criteria, "worker.status.phase")
	criteria["canceled"] = cancellationTime
	findOptions := options.Find()
	findOptions.SetSort(
		// bson.D preserves order so we use this wherever we sort so that if
		// additional sort criteria are added in the future, they will be applied
		// in the specified order.
		bson.D{
			{Key: "created", Value: -1},
		},
	)
	cur, err := e.collection.Find(ctx, criteria, findOptions)
	if err != nil {
		return events, errors.Wrapf(err, "error finding canceled events")
	}
	if err := cur.All(ctx, &events.Items); err != nil {
		return events, errors.Wrap(err, "error decoding canceled events")
	}

	return events, nil
}

func (e *eventsStore) Delete(ctx context.Context, id string) error {
	res, err := e.collection.DeleteOne(
		ctx,
		bson.M{
			"id": id,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error deleting event %q", id)
	}
	if res.DeletedCount != 1 {
		return &meta.ErrNotFound{
			Type: "Event",
			ID:   id,
		}
	}
	return nil
}

func (e *eventsStore) DeleteMany(
	ctx context.Context,
	selector core.EventsSelector,
) (core.EventList, error) {
	events := core.EventList{}

	// The MongoDB driver for Go doesn't expose findAndModify(), which could be
	// used to select events and delete them at the same time. As a workaround,
	// we'll perform a logical delete first, select the logically deleted events,
	// and then perform a real delete.

	deletedTime := time.Now().UTC()

	// Logical delete...
	criteria := bson.M{
		"projectID": selector.ProjectID,
		"worker.status.phase": bson.M{
			"$in": selector.WorkerPhases,
		},
		"deleted": bson.M{
			"$exists": false,
		},
	}
	if _, err := e.collection.UpdateMany(
		ctx,
		criteria,
		bson.M{
			"$set": bson.M{
				"deleted": deletedTime,
			},
		},
	); err != nil {
		return events, errors.Wrap(err, "error logically deleting events")
	}

	// Select the logically deleted documents...
	criteria["deleted"] = deletedTime
	findOptions := options.Find()
	findOptions.SetSort(
		// bson.D preserves order so we use this wherever we sort so that if
		// additional sort criteria are added in the future, they will be applied
		// in the specified order.
		bson.D{
			{Key: "created", Value: -1},
		},
	)
	cur, err := e.collection.Find(ctx, criteria, findOptions)
	if err != nil {
		return events, errors.Wrapf(
			err,
			"error finding logically deleted events",
		)
	}
	if err := cur.All(ctx, &events.Items); err != nil {
		return events, errors.Wrap(
			err,
			"error decoding logically deleted events",
		)
	}

	// Final deletion
	if _, err := e.collection.DeleteMany(ctx, criteria); err != nil {
		return events, errors.Wrap(err, "error deleting events")
	}

	return events, nil
}
