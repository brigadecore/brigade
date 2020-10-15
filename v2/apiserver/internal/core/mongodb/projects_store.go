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
	"go.mongodb.org/mongo-driver/mongo/options"
)

// projectsStore is a a MongoDB-based implementation of the core.ProjectsStore
// interface.
type projectsStore struct {
	collection mongodb.Collection
}

// NewProjectsStore returns a MongoDB-based implementation of the
// core.ProjectsStore interface.
func NewProjectsStore(database *mongo.Database) core.ProjectsStore {
	return &projectsStore{
		collection: database.Collection("projects"),
	}
}

func (p *projectsStore) Create(
	ctx context.Context,
	project core.Project,
) error {
	if _, err := p.collection.InsertOne(ctx, project); err != nil {
		if writeException, ok := err.(mongo.WriteException); ok {
			if len(writeException.WriteErrors) == 1 &&
				writeException.WriteErrors[0].Code == 11000 {
				return &meta.ErrConflict{
					Type: "Project",
					ID:   project.ID,
					Reason: fmt.Sprintf(
						"A project with the ID %q already exists.",
						project.ID,
					),
				}
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
	findOptions.SetSort(bson.M{"id": 1})
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

func (p *projectsStore) Get(
	ctx context.Context,
	id string,
) (core.Project, error) {
	project := core.Project{}
	res := p.collection.FindOne(ctx, bson.M{"id": id})
	err := res.Decode(&project)
	if err == mongo.ErrNoDocuments {
		return project, &meta.ErrNotFound{
			Type: "Project",
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
			Type: "Project",
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
			Type: "Project",
			ID:   id,
		}
	}
	return nil
}
