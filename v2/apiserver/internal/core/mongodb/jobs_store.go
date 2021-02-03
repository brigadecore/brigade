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

func (j *jobsStore) Cancel(ctx context.Context, id, job string) error {
	status, err := j.GetStatus(ctx, id, job)
	if err != nil {
		return errors.Wrapf(err, "unable to get status of job %q", job)
	}

	now := time.Now()
	updatedStatus := core.JobStatus{
		Ended: &now,
	}

	// TODO: need to check if job phase already terminal?
	// Or does existing logic in events store prevent this?
	switch status.Phase {
	case core.JobPhasePending:
		updatedStatus.Phase = core.JobPhaseCanceled
	case core.JobPhaseRunning, core.JobPhaseStarting:
		updatedStatus.Phase = core.JobPhaseAborted
	}

	if err := j.UpdateStatus(
		ctx,
		id,
		job,
		updatedStatus,
	); err != nil {
		return errors.Wrapf(
			err,
			"error updating status of worker job %q for event %q",
			job,
			id,
		)
	}

	return nil
}

func (j *jobsStore) Create(
	ctx context.Context,
	eventID string,
	jobName string,
	job core.Job,
) error {
	res, err := j.collection.UpdateOne(
		ctx,
		bson.M{"id": eventID},
		bson.M{
			"$set": bson.M{
				fmt.Sprintf("worker.jobs.%s", jobName): job,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error updating spec of event %q job %q",
			eventID,
			jobName,
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

func (j *jobsStore) GetStatus(
	ctx context.Context,
	eventID string,
	jobName string,
) (core.JobStatus, error) {
	// The jobsStore's collection is the events collection
	// So, find the event and then inspect the decoded object
	event := core.Event{}
	res := j.collection.FindOne(ctx, bson.M{"id": eventID})
	err := res.Decode(&event)
	if err == mongo.ErrNoDocuments {
		return core.JobStatus{}, &meta.ErrNotFound{
			Type: "Event",
			ID:   eventID,
		}
	}
	if err != nil {
		return core.JobStatus{},
			errors.Wrapf(
				res.Err(),
				"error finding/decoding event %q for worker job %q",
				eventID,
				jobName,
			)
	}

	job, ok := event.Worker.Jobs[jobName]
	if !ok {
		return core.JobStatus{},
			&meta.ErrNotFound{
				Type: "Job",
				ID:   jobName,
			}
	}

	return *job.Status, nil
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
			"id": eventID,
		},
		bson.M{
			"$set": bson.M{
				fmt.Sprintf("worker.jobs.%s.status", jobName): status,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error updating status of event %q worker job %q",
			eventID,
			jobName,
		)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: "Job",
			ID:   eventID,
		}
	}
	return nil
}
