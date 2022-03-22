package mongodb

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// eventsStore is a MongoDB-based implementation of the api.EventsStore
// interface.
type eventsStore struct {
	collection mongodb.Collection
}

// NewEventsStore returns a MongoDB-based implementation of the api.EventsStore
// interface.
func NewEventsStore(database *mongo.Database) (api.EventsStore, error) {
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

func (e *eventsStore) Create(ctx context.Context, event api.Event) error {
	// We need this to be non-nil before we store or various queries and
	// statements won't work correctly.
	if event.Worker.Jobs == nil {
		event.Worker.Jobs = []api.Job{}
	}
	if _, err := e.collection.InsertOne(ctx, event); err != nil {
		return errors.Wrapf(err, "error inserting new event %q", event.ID)
	}
	return nil
}

func (e *eventsStore) List(
	ctx context.Context,
	selector api.EventsSelector,
	opts meta.ListOptions,
) (meta.List[api.Event], error) {
	events := meta.List[api.Event]{}

	criteria := bson.M{
		"deleted": bson.M{
			"$exists": false, // Don't grab logically deleted events
		},
	}
	if selector.ProjectID != "" {
		criteria["projectID"] = selector.ProjectID
	}
	if selector.Source != "" {
		criteria["source"] = selector.Source
	}
	for k, v := range selector.Qualifiers {
		criteria[fmt.Sprintf("qualifiers.%s", k)] = v
	}
	for k, v := range selector.Labels {
		criteria[fmt.Sprintf("labels.%s", k)] = v
	}
	for k, v := range selector.SourceState {
		criteria[fmt.Sprintf("sourceState.state.%s", k)] = v
	}
	if selector.Type != "" {
		criteria["type"] = selector.Type
	}
	if len(selector.WorkerPhases) > 0 {
		criteria["worker.status.phase"] = bson.M{
			"$in": selector.WorkerPhases,
		}
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

	if events.Len() == opts.Limit {
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
) (api.Event, error) {
	event := api.Event{}
	res := e.collection.FindOne(
		ctx,
		bson.M{
			"id": id,
			"deleted": bson.M{
				"$exists": false, // Don't grab logically deleted events
			},
		})
	err := res.Decode(&event)
	if err == mongo.ErrNoDocuments {
		return event, &meta.ErrNotFound{
			Type: api.EventKind,
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
) (api.Event, error) {
	event := api.Event{}
	res := e.collection.FindOne(
		ctx,
		bson.M{
			"worker.hashedToken": hashedWorkerToken,
			"deleted": bson.M{
				"$exists": false, // Don't grab logically deleted events
			},
		},
	)
	err := res.Decode(&event)
	if res.Err() == mongo.ErrNoDocuments {
		return event, &meta.ErrNotFound{
			Type: api.EventKind,
		}
	}
	if err != nil {
		return event, errors.Wrap(err, "error finding/decoding event")
	}
	return event, nil
}

func (e *eventsStore) UpdateSourceState(
	ctx context.Context,
	id string,
	sourceState api.SourceState,
) error {
	res, err := e.collection.UpdateOne(
		ctx,
		bson.M{
			"id": id,
			"deleted": bson.M{
				"$exists": false, // Don't grab logically deleted events
			},
		},
		bson.M{
			"$set": bson.M{
				"sourceState": sourceState,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error updating source state of event %q",
			id,
		)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: "Event",
			ID:   id,
		}
	}
	return nil
}

func (e *eventsStore) UpdateSummary(
	ctx context.Context,
	id string,
	summary api.EventSummary,
) error {
	res, err := e.collection.UpdateOne(
		ctx,
		bson.M{
			"id": id,
			"deleted": bson.M{
				"$exists": false, // Don't grab logically deleted events
			},
		},
		bson.M{
			"$set": bson.M{
				"summary": summary.Text,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error updating summary of event %q",
			id,
		)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: "Event",
			ID:   id,
		}
	}
	return nil
}

func (e *eventsStore) Cancel(ctx context.Context, id string) error {
	cancellationTime := time.Now().UTC()

	res, err := e.collection.UpdateOne(
		ctx,
		bson.M{
			"id":                  id,
			"worker.status.phase": api.WorkerPhasePending,
			"deleted": bson.M{
				"$exists": false, // Don't grab logically deleted events
			},
		},
		bson.M{
			"$set": bson.M{
				"canceled":            cancellationTime,
				"worker.status.phase": api.WorkerPhaseCanceled,
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
				"$in": []api.WorkerPhase{
					api.WorkerPhaseStarting,
					api.WorkerPhaseRunning,
				},
			},
		},
		bson.M{
			"$set": bson.M{
				"worker.status.phase":                           api.WorkerPhaseAborted, // nolint: lll
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
		return errors.Wrapf(err, "error updating status of event %q worker", id)
	}

	if res.MatchedCount == 0 {
		return &meta.ErrConflict{
			Type: api.EventKind,
			ID:   id,
			Reason: fmt.Sprintf(
				"Event %q was not canceled because it was already in a terminal state.",
				id,
			),
		}
	}

	return nil
}

// nolint: gocyclo
func (e *eventsStore) CancelMany(
	ctx context.Context,
	selector api.EventsSelector,
) (<-chan api.Event, int64, error) {
	var affectedCount int64
	// It only makes sense to cancel events that are in a pending, starting, or
	// running state. We can ignore anything else.
	var cancelPending bool
	var cancelStarting bool
	var cancelRunning bool
	for _, workerPhase := range selector.WorkerPhases {
		if workerPhase == api.WorkerPhasePending {
			cancelPending = true
		}
		if workerPhase == api.WorkerPhaseStarting {
			cancelStarting = true
		}
		if workerPhase == api.WorkerPhaseRunning {
			cancelRunning = true
		}
	}

	// Bail if we're not canceling pending, starting, or running events
	if !cancelPending && !cancelStarting && !cancelRunning {
		return nil, 0, nil
	}

	// The MongoDB driver for Go doesn't expose findAndModify(), which could be
	// used to select events and cancel them at the same time. As a workaround,
	// we'll cancel first, then select events based on cancellation time.

	cancellationTime := time.Now().UTC()

	criteria := bson.M{
		"projectID": selector.ProjectID,
		"deleted": bson.M{
			"$exists": false, // Don't grab logically deleted events
		},
	}
	if selector.Source != "" {
		criteria["source"] = selector.Source
	}
	for k, v := range selector.SourceState {
		criteria[fmt.Sprintf("sourceState.state.%s", k)] = v
	}
	if selector.Type != "" {
		criteria["type"] = selector.Type
	}

	if cancelPending {
		criteria["worker.status.phase"] = api.WorkerPhasePending
		result, err := e.collection.UpdateMany(
			ctx,
			criteria,
			bson.M{
				"$set": bson.M{
					"canceled":            cancellationTime,
					"worker.status.phase": api.WorkerPhaseCanceled,
				},
			},
		)
		if err != nil {
			return nil, 0, errors.Wrap(err, "error updating events")
		}
		affectedCount = affectedCount + result.ModifiedCount
	}

	if cancelStarting && cancelRunning {
		criteria["worker.status.phase"] = bson.M{
			"$in": []api.WorkerPhase{
				api.WorkerPhaseStarting,
				api.WorkerPhaseRunning,
			},
		}
	} else if cancelStarting {
		criteria["worker.status.phase"] = api.WorkerPhaseStarting
	} else if cancelRunning {
		criteria["worker.status.phase"] = api.WorkerPhaseRunning
	}

	if cancelStarting || cancelRunning {
		result, err := e.collection.UpdateMany(
			ctx,
			criteria,
			bson.M{
				// nolint: lll
				"$set": bson.M{
					"canceled":                                      cancellationTime,
					"worker.status.phase":                           api.WorkerPhaseAborted,
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
			return nil, 0, errors.Wrap(err, "error updating events")
		}
		affectedCount = affectedCount + result.ModifiedCount
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
		return nil, 0, errors.Wrapf(err, "error finding canceled events")
	}

	eventCh := make(chan api.Event)
	go func() {
		defer close(eventCh)
		for cur.Next(ctx) {
			event := api.Event{}
			if err := cur.Decode(&event); err != nil {
				log.Println(errors.Wrap(err, "error decoding event"))
			}
			eventCh <- event
		}
	}()
	return eventCh, affectedCount, nil
}

func (e *eventsStore) Delete(ctx context.Context, id string) error {
	res, err := e.collection.DeleteOne(
		ctx,
		bson.M{
			"id": id,
			"deleted": bson.M{
				"$exists": false, // Don't grab logically deleted events
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error deleting event %q", id)
	}
	if res.DeletedCount != 1 {
		return &meta.ErrNotFound{
			Type: api.EventKind,
			ID:   id,
		}
	}
	return nil
}

func (e *eventsStore) DeleteMany(
	ctx context.Context,
	selector api.EventsSelector,
) (<-chan api.Event, int64, error) {
	// The MongoDB driver for Go doesn't expose findAndModify(), which could be
	// used to select events and delete them at the same time. As a workaround,
	// we'll perform a logical delete first, select the logically deleted events,
	// and then perform a real delete.

	deletedTime := time.Now().UTC()

	// Logical delete...
	criteria := bson.M{
		"projectID": selector.ProjectID,
		"deleted": bson.M{
			"$exists": false, // Don't grab logically deleted events
		},
	}
	if selector.Source != "" {
		criteria["source"] = selector.Source
	}
	for k, v := range selector.SourceState {
		criteria[fmt.Sprintf("sourceState.state.%s", k)] = v
	}
	if selector.Type != "" {
		criteria["type"] = selector.Type
	}
	if len(selector.WorkerPhases) > 0 {
		criteria["worker.status.phase"] = bson.M{
			"$in": selector.WorkerPhases,
		}
	}
	result, err := e.collection.UpdateMany(
		ctx,
		criteria,
		bson.M{
			"$set": bson.M{
				"deleted": deletedTime,
			},
		},
	)
	if err != nil {
		return nil, 0, errors.Wrap(err, "error logically deleting events")
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
		return nil, 0, errors.Wrapf(
			err,
			"error finding logically deleted events",
		)
	}

	eventCh := make(chan api.Event)
	go func() {
		defer close(eventCh)
		for cur.Next(ctx) {
			event := api.Event{}
			if err := cur.Decode(&event); err != nil {
				log.Println(errors.Wrap(err, "error decoding event"))
			}
			eventCh <- event
		}
		// Final deletion
		if _, err := e.collection.DeleteMany(
			context.Background(), // deliberately not using ctx
			criteria,
		); err != nil {
			log.Println(errors.Wrap(err, "error deleting events"))
		}
	}()

	return eventCh, result.ModifiedCount, nil
}

func (e *eventsStore) DeleteByProjectID(
	ctx context.Context,
	projectID string,
) error {
	_, err := e.collection.DeleteMany(
		ctx,
		bson.M{
			"projectID": projectID,
		},
	)
	return errors.Wrapf(err, "error deleting events for project %q", projectID)
}
