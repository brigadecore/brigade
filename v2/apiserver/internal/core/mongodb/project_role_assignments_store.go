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
	if res := p.collection.FindOneAndReplace(
		ctx,
		projectRoleAssignment,
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
	roleAssignment core.ProjectRoleAssignment,
) error {
	if _, err := p.collection.DeleteOne(ctx, roleAssignment); err != nil {
		return errors.Wrapf(
			err,
			"error deleting project role assignment %v",
			roleAssignment,
		)
	}
	return nil
}

func (p *projectRoleAssignmentsStore) RevokeMany(
	ctx context.Context,
	projectID string,
) error {
	criteria := bson.M{
		"role.projectID": projectID,
	}
	if _, err := p.collection.DeleteMany(ctx, criteria); err != nil {
		return errors.Wrap(err, "error deleting project role assignments")
	}
	return nil
}

func (p *projectRoleAssignmentsStore) Exists(
	ctx context.Context,
	roleAssignment core.ProjectRoleAssignment,
) (bool, error) {
	criteria := bson.M{
		"role.name": roleAssignment.Role.Name,
		"role.projectID": bson.M{
			"$in": []string{
				roleAssignment.Role.ProjectID,
				core.ProjectIDGlobal,
			},
		},
		"principal.type": roleAssignment.Principal.Type,
		"principal.id":   roleAssignment.Principal.ID,
	}
	err := p.collection.FindOne(ctx, criteria).Err()
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "error finding role assignment")
	}
	return true, nil
}
