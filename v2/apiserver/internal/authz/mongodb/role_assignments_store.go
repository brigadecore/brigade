package mongodb

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authz"
	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// rolesAssignmentsStore is a MongoDB-based implementation of the
// authz.RoleAssignmentsStore interface.
type roleAssignmentsStore struct {
	collection mongodb.Collection
}

// NewRoleAssignmentsStore returns a MongoDB-based implementation of the
// authz.RoleAssignmentsStore interface.
func NewRoleAssignmentsStore(
	database *mongo.Database,
) authz.RoleAssignmentsStore {
	// TODO: Add indices
	return &roleAssignmentsStore{
		collection: database.Collection("role-assignments"),
	}
}

func (r *roleAssignmentsStore) Grant(
	ctx context.Context,
	roleAssignment libAuthz.RoleAssignment,
) error {
	tru := true
	if res := r.collection.FindOneAndReplace(
		ctx,
		roleAssignment,
		roleAssignment,
		&options.FindOneAndReplaceOptions{
			Upsert: &tru,
		},
	); res.Err() != nil && res.Err() != mongo.ErrNoDocuments {
		return errors.Wrapf(
			res.Err(),
			"error upserting role assignment %v",
			roleAssignment,
		)
	}
	return nil
}

func (r *roleAssignmentsStore) Revoke(
	ctx context.Context,
	roleAssignment libAuthz.RoleAssignment,
) error {
	if _, err := r.collection.DeleteOne(ctx, roleAssignment); err != nil {
		return errors.Wrapf(
			err,
			"error deleting role assignment %v",
			roleAssignment,
		)
	}
	return nil
}

func (r *roleAssignmentsStore) Exists(
	ctx context.Context,
	roleAssignment libAuthz.RoleAssignment,
) (bool, error) {
	criteria := bson.M{
		"role.name":      roleAssignment.Role.Name,
		"principal.type": roleAssignment.Principal.Type,
		"principal.id":   roleAssignment.Principal.ID,
	}
	if roleAssignment.Scope == "" {
		criteria["scope"] = bson.M{
			"$exists": false,
		}
	} else {
		criteria["scope"] = bson.M{
			"$in": []string{roleAssignment.Scope, "*"},
		}
	}
	if err :=
		r.collection.FindOne(ctx, criteria).Err(); err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, errors.Wrap(err, "error finding role assignment")
	}
	return true, nil
}
