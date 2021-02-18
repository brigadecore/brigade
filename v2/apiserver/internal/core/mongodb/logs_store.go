package mongodb

import (
	"context"
	"log"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// logsStore is a MongoDB-based implementation of the core.LogsStore interface.
type logsStore struct {
	collection mongodb.Collection
}

// NewLogsStore returns a MongoDB-based implementation of the core.LogsStore
// interface. This implementation relies on a log aggregator having forwarded
// and stored log entries-- a process which necessarily introduces some latency.
// Callers should favor another implementation of the core.LogsStore interface
// and fall back on this implementation when the other fails.
func NewLogsStore(database *mongo.Database) core.LogsStore {
	return &logsStore{
		collection: database.Collection("logs"),
	}
}

func (l *logsStore) StreamLogs(
	ctx context.Context,
	_ core.Project,
	event core.Event,
	selector core.LogsSelector,
	opts core.LogStreamOptions,
) (<-chan core.LogEntry, error) {
	criteria := criteriaFromSelector(event.ID, selector)

	logEntryCh := make(chan core.LogEntry)
	go func() {
		defer close(logEntryCh)

		cur, err := l.collection.Find(
			ctx,
			criteria,
			&options.FindOptions{},
		)

		for cur.Next(ctx) {
			logEntry := core.LogEntry{}
			err = cur.Decode(&logEntry)
			if err != nil {
				log.Println(
					errors.Wrapf(err, "error decoding log entry from collection"),
				)
				return
			}

			select {
			case logEntryCh <- logEntry:
			case <-ctx.Done():
				return
			}
		}
	}()

	return logEntryCh, nil
}

func criteriaFromSelector(
	eventID string,
	selector core.LogsSelector,
) bson.M {
	criteria := bson.M{
		"event": eventID,
	}
	if selector.Job == "" { // We want worker logs
		criteria["component"] = "worker"
	} else { // We want job logs
		criteria["component"] = "job"
		criteria["job"] = selector.Job
	}
	criteria["container"] = selector.Container
	return criteria
}
