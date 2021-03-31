package mongodb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// projectsStore is a MongoDB-based implementation of the core.ProjectsStore
// interface.
type projectsStore struct {
	collection mongodb.Collection
}

// NewProjectsStore returns a MongoDB-based implementation of the
// core.ProjectsStore interface.
func NewProjectsStore(database *mongo.Database) (core.ProjectsStore, error) {
	ctx, cancel :=
		context.WithTimeout(context.Background(), createIndexTimeout)
	defer cancel()
	unique := true
	collection := database.Collection("projects")
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
		return nil, errors.Wrap(
			err,
			"error adding indexes to projects collection",
		)
	}
	return &projectsStore{
		collection: collection,
	}, nil
}

func (p *projectsStore) Create(
	ctx context.Context,
	project core.Project,
) error {
	if _, err := p.collection.InsertOne(ctx, project); err != nil {
		if mongodb.IsDuplicateKeyError(err) {
			return &meta.ErrConflict{
				Type: core.ProjectKind,
				ID:   project.ID,
				Reason: fmt.Sprintf(
					"A project with the ID %q already exists.",
					project.ID,
				),
			}
		}
		return errors.Wrapf(err, "error inserting new project %q", project.ID)
	}
	return nil
}

func (p *projectsStore) List(
	ctx context.Context,
	opts meta.ListOptions,
) (core.ProjectList, error) {
	projects := core.ProjectList{}

	criteria := bson.M{}
	if opts.Continue != "" {
		criteria["id"] = bson.M{"$gt": opts.Continue}
	}

	findOptions := options.Find()
	findOptions.SetSort(
		// bson.D preserves order so we use this wherever we sort so that if
		// additional sort criteria are added in the future, they will be applied
		// in the specified order.
		bson.D{
			{Key: "id", Value: 1},
		},
	)
	findOptions.SetLimit(opts.Limit)
	cur, err := p.collection.Find(ctx, criteria, findOptions)
	if err != nil {
		return projects, errors.Wrap(err, "error finding projects")
	}
	if err := cur.All(ctx, &projects.Items); err != nil {
		return projects, errors.Wrap(err, "error decoding projects")
	}

	if int64(len(projects.Items)) == opts.Limit {
		continueID := projects.Items[opts.Limit-1].ID
		criteria["id"] = bson.M{"$gt": continueID}
		remaining, err := p.collection.CountDocuments(ctx, criteria)
		if err != nil {
			return projects, errors.Wrap(err, "error counting remaining projects")
		}
		if remaining > 0 {
			projects.Continue = continueID
			projects.RemainingItemCount = remaining
		}
	}

	return projects, nil
}

func (p *projectsStore) ListSubscribers(
	ctx context.Context,
	event core.Event,
) (core.ProjectList, error) {
	projects := core.ProjectList{}
	// Finding all Projects that are subscribed to a given Event is rather tricky.
	// Matching on the basis of source, type, and qualifiers is easy enough, but
	// LABELS account for the difficulty. If we were selecting Events, it's easy
	// enough to say "select all events that have (at least) these labels," but
	// when selecting PROJECTS, we're effectively working in reverse. We want to
	// select all Projects having an EventSubscription that matches the Event's
	// source, type, and qualifiers, AND doesn't have any labels that PRECLUDE a
	// match. The only practical way of expressing those criteria in a MongoDB
	// query is to provide a JavaScript function as a "where" clause. That
	// function contains the logic for determining whether a SINGLE document
	// (Project) is a match for the Event. For a large number of Projects, such a
	// query is wildly inefficient, since the function needs to be applied to
	// every single document (Project). To compensate, we build PRELIMINARY match
	// criteria based on matching source, type, and qualifiers. MongoDB sensibly
	// applies these criteria FIRST, then iterates over the preliminary results
	// only, applying the provided "where" function to each in turn.
	preliminaryMatchCriteria := bson.M{
		"source": event.Source,
		"types": bson.M{
			"$in": []string{event.Type, "*"},
		},
	}
	if len(event.Qualifiers) > 0 {
		preliminaryMatchCriteria["qualifiers"] = event.Qualifiers
	}
	eventJSON, err := json.Marshal(
		struct {
			Source     string            `json:"source"`
			Type       string            `json:"type"`
			Qualifiers core.Qualifiers   `json:"qualifiers"`
			Labels     map[string]string `json:"labels"`
		}{
			Source:     event.Source,
			Type:       event.Type,
			Qualifiers: event.Qualifiers,
			Labels:     event.Labels,
		},
	)
	// nolint: lll
	where := fmt.Sprintf(`
function() {
	const project = this;
	if (!project.spec.eventSubscriptions) {
		return false;
	}
	const event = %s;
	event.qualifiers = event.qualifiers || {};
	event.labels = event.labels || {};
	loop:
	for (const subscription of project.spec.eventSubscriptions) {
		subscription.qualifiers = subscription.qualifiers || {};
		subscription.labels = subscription.labels || {};
		if (subscription.source != event.source) continue;
		if (!subscription.types.includes(event.type) && !subscription.types.includes("*")) continue; 
		for (const key in event.qualifiers) {
			if (subscription.qualifiers[key] != event.qualifiers[key]) continue loop;
		}
		for (const key in subscription.qualifiers) {
			if (subscription.qualifiers[key] != event.qualifiers[key]) continue loop;
		}
		const eventLabelKeys = Object.keys(event.labels);
		for (const labelKey in subscription.labels) {
			if (!eventLabelKeys.includes(labelKey)) {
				continue loop;
			}
			if (subscription.labels[labelKey] != event.labels[labelKey]) {
				continue loop;
			}
		}
		return true;
	}
	return false;
}`,
		eventJSON,
	)
	findOptions := options.Find()
	findOptions.SetSort(
		// bson.D preserves order so we use this wherever we sort so that if
		// additional sort criteria are added in the future, they will be applied
		// in the specified order.
		bson.D{
			{Key: "id", Value: 1},
		},
	)
	cur, err := p.collection.Find(
		ctx,
		bson.M{
			"spec.eventSubscriptions": bson.M{
				"$elemMatch": preliminaryMatchCriteria,
			},
			"$where": where,
		},
		findOptions,
	)
	if err != nil {
		return projects, errors.Wrap(err, "error finding projects")
	}
	if err := cur.All(ctx, &projects.Items); err != nil {
		return projects, errors.Wrap(err, "error decoding projects")
	}
	return projects, nil
}

func (p *projectsStore) Get(
	ctx context.Context,
	id string,
) (core.Project, error) {
	project := core.Project{}
	res := p.collection.FindOne(ctx, bson.M{"id": id})
	err := res.Decode(&project)
	if err == mongo.ErrNoDocuments {
		return project, &meta.ErrNotFound{
			Type: core.ProjectKind,
			ID:   id,
		}
	}
	if err != nil {
		return project,
			errors.Wrapf(res.Err(), "error finding/decoding project %q", id)
	}
	return project, nil
}

func (p *projectsStore) Update(
	ctx context.Context, project core.Project,
) error {
	res, err := p.collection.UpdateOne(
		ctx,
		bson.M{
			"id": project.ID,
		},
		bson.M{
			"$set": bson.M{
				"description": project.Description,
				"spec":        project.Spec,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating project %q", project.ID)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: core.ProjectKind,
			ID:   project.ID,
		}
	}
	return nil
}

func (p *projectsStore) Delete(ctx context.Context, id string) error {
	res, err := p.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return errors.Wrapf(err, "error deleting project %q", id)
	}
	if res.DeletedCount == 0 {
		return &meta.ErrNotFound{
			Type: core.ProjectKind,
			ID:   id,
		}
	}
	return nil
}
