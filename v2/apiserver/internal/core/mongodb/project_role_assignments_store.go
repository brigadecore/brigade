package mongodb

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// projectRoleAssignmentsStore is a MongoDB-based implementation of the
// core.ProjectRoleAssignmentsStore interface.
type projectRoleAssignmentsStore struct {
	collection mongodb.Collection
}

// NewProjectRoleAssignmentsStore returns a MongoDB-based implementation of the
// core.ProjectRoleAssignmentsStore interface.
func NewProjectRoleAssignmentsStore(
	database *mongo.Database,
) core.ProjectRoleAssignmentsStore {
	// TODO: Add indices
	return &projectRoleAssignmentsStore{
		collection: database.Collection("project-role-assignments"),
	}
}

func (p *projectRoleAssignmentsStore) Grant(
	ctx context.Context,
	projectRoleAssignment core.ProjectRoleAssignment,
) error {
	tru := true
	criteria := bson.M{
		"projectID":      projectRoleAssignment.ProjectID,
		"role.name":      projectRoleAssignment.Role.Name,
		"principal.type": projectRoleAssignment.Principal.Type,
		"principal.id":   projectRoleAssignment.Principal.ID,
	}
	if res := p.collection.FindOneAndReplace(
		ctx,
		criteria,
		projectRoleAssignment,
		&options.FindOneAndReplaceOptions{
			Upsert: &tru,
		},
	); res.Err() != nil && res.Err() != mongo.ErrNoDocuments {
		return errors.Wrapf(
			res.Err(),
			"error upserting project role assignment %v",
			projectRoleAssignment,
		)
	}
	return nil
}

func (p *projectRoleAssignmentsStore) Revoke(
	ctx context.Context,
	projectRoleAssignment core.ProjectRoleAssignment,
) error {
	criteria := bson.M{
		"projectID":      projectRoleAssignment.ProjectID,
		"role.name":      projectRoleAssignment.Role.Name,
		"principal.type": projectRoleAssignment.Principal.Type,
		"principal.id":   projectRoleAssignment.Principal.ID,
	}
	if _, err := p.collection.DeleteOne(ctx, criteria); err != nil {
		return errors.Wrapf(
			err,
			"error deleting project role assignment %v",
			projectRoleAssignment,
		)
	}
	return nil
}

func (p *projectRoleAssignmentsStore) RevokeByProjectID(
	ctx context.Context,
	projectID string,
) error {
	if _, err := p.collection.DeleteMany(
		ctx,
		bson.M{"projectID": projectID},
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting project role assignments for project %q",
			projectID,
		)
	}
	return nil
}

func (p *projectRoleAssignmentsStore) Exists(
	ctx context.Context,
	projectRoleAssignment core.ProjectRoleAssignment,
) (bool, error) {
	criteria := bson.M{
		"projectID":      projectRoleAssignment.ProjectID,
		"role.name":      projectRoleAssignment.Role.Name,
		"principal.type": projectRoleAssignment.Principal.Type,
		"principal.id":   projectRoleAssignment.Principal.ID,
	}
	if err :=
		p.collection.FindOne(ctx, criteria).Err(); err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, errors.Wrap(err, "error finding project role assignment")
	}
	return true, nil
}
