package mongodb

import (
	"context"
	"fmt"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// jobsStore is a MongoDB-based implementation of the core.JobsStore interface.
type jobsStore struct {
	collection mongodb.Collection
}

// NewJobsStore returns a MongoDB-based implementation of the core.JobsStore
// interface.
func NewJobsStore(database *mongo.Database) (core.JobsStore, error) {
	return &jobsStore{
		collection: database.Collection("events"),
	}, nil
}

func (j *jobsStore) Create(
	ctx context.Context,
	eventID string,
	job core.Job,
) error {
	res, err := j.collection.UpdateOne(
		ctx,
		bson.M{"id": eventID},
		bson.M{
			"$push": bson.M{
				"worker.jobs": job,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error creating event %q job %q",
			eventID,
			job.Name,
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

func (j *jobsStore) UpdateStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status core.JobStatus,
) error {
	res, err := j.collection.UpdateOne(
		ctx,
		bson.M{
			"id":               eventID,
			"worker.jobs.name": jobName,
		},
		bson.M{
			"$set": bson.M{
				"worker.jobs.$.status": status,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error updating status of event %q job %q",
			eventID,
			jobName,
		)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: core.JobKind,
			ID:   fmt.Sprintf("%s:%s", eventID, jobName),
		}
	}
	return nil
}
