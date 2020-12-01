package mongodb

import (
	"context"
	"log"
	"time"

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
// and stored log entries-- a process which necessary introduces some latency.
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

		cursorType := options.Tailable
		var cur *mongo.Cursor
		var err error
		// Any attempt to open a cursor that initially retrieves nothing will yield
		// a "dead" cursor which is no good for tailing the capped collection. We
		// need to retry this until we get a "live" cursor or the context is
		// canceled.
		for {
			cur, err = l.collection.Find(
				ctx,
				criteria,
				&options.FindOptions{CursorType: &cursorType},
			)
			if err != nil {
				log.Println(
					errors.Wrap(err, "error getting cursor for logs collection"),
				)
				return
			}
			if cur.ID() != 0 {
				// We got a live cursor.
				break
			}
			if !opts.Follow {
				// If we're not following, just return. We're done.
				return
			}
			select {
			case <-time.After(time.Second): // Wait before retry
			case <-ctx.Done():
				return
			}
		}

		var available bool
		for {
			available = cur.TryNext(ctx)
			if !available {
				if !opts.Follow {
					// If we're not following, just return. We're done.
					return
				}
				select {
				case <-time.After(time.Second): // Wait before retry
					continue
				case <-ctx.Done():
					return
				}
			}
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

	// If no job was specified, we want worker logs
	if selector.Job == "" {
		criteria["component"] = "worker"
		// If no container was specified, we want the "worker" container
		if selector.Container == "" {
			criteria["container"] = "worker"
		} else { // We want the one specified
			criteria["container"] = selector.Container
		}
	} else { // We want job logs
		criteria["component"] = "job"
		// If no container was specified, we want the one with the same name as the
		// job
		if selector.Container == "" {
			criteria["container"] = selector.Job
		} else { // We want the one specified
			criteria["container"] = selector.Container
		}
	}

	return criteria
}
