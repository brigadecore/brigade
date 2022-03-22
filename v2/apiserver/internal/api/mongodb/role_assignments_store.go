package mongodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// rolesAssignmentsStore is a MongoDB-based implementation of the
// api.RoleAssignmentsStore interface.
type roleAssignmentsStore struct {
	collection mongodb.Collection
}

// NewRoleAssignmentsStore returns a MongoDB-based implementation of the
// api.RoleAssignmentsStore interface.
func NewRoleAssignmentsStore(
	database *mongo.Database,
) api.RoleAssignmentsStore {
	// TODO: Add indices
	return &roleAssignmentsStore{
		collection: database.Collection("role-assignments"),
	}
}

func (r *roleAssignmentsStore) Grant(
	ctx context.Context,
	roleAssignment api.RoleAssignment,
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

func (r *roleAssignmentsStore) List(
	ctx context.Context,
	selector api.RoleAssignmentsSelector,
	opts meta.ListOptions,
) (meta.List[api.RoleAssignment], error) {
	roleAssignments := meta.List[api.RoleAssignment]{}

	criteria := bson.M{}
	if selector.Principal != nil {
		criteria["principal.type"] = selector.Principal.Type
		criteria["principal.id"] = selector.Principal.ID
	}
	if selector.Role != "" {
		criteria["role"] = selector.Role
	}
	if opts.Continue != "" {
		tokens := strings.Split(opts.Continue, ":")
		if len(tokens) != 4 {
			return roleAssignments, errors.New("error parsing continue value")
		}
		continuePrincipalType := tokens[0]
		continuePrincipalID := tokens[1]
		continueRole := tokens[2]
		continueScope := tokens[3]
		criteria["$or"] = []bson.M{
			// Same principal type, principal id, and role, but scopes we didn't list
			// yet
			{
				"principal.type": continuePrincipalType,
				"principal.id":   continuePrincipalID,
				"role":           continueRole,
				"scope":          bson.M{"$gt": continueScope},
			},
			// Same principal type and principal id, but roles and scopes we didn't
			// list yet
			{
				"principal.type": continuePrincipalType,
				"principal.id":   continuePrincipalID,
				"role":           bson.M{"$gt": continueRole},
			},
			// Same principal type, but principal ids, roles, and scopes we didn't
			// list yet
			{
				"principal.type": continuePrincipalType,
				"principal.id":   bson.M{"$gt": continuePrincipalID},
			},
			// Anything remaining that we didn't list yet
			{
				"principal.type": bson.M{"$gt": continuePrincipalType},
			},
		}
	}

	findOptions := options.Find()
	findOptions.SetSort(
		// bson.D preserves order, and we want to sort in this specific order
		bson.D{
			{Key: "principal.type", Value: 1},
			{Key: "principal.id", Value: 1},
			{Key: "role", Value: 1},
			{Key: "scope", Value: 1},
		},
	)
	findOptions.SetLimit(opts.Limit)
	cur, err := r.collection.Find(ctx, criteria, findOptions)
	if err != nil {
		return roleAssignments, errors.Wrap(err, "error finding role assignments")
	}
	if err := cur.All(ctx, &roleAssignments.Items); err != nil {
		return roleAssignments, errors.Wrap(err, "error decoding role assignments")
	}

	if roleAssignments.Len() == opts.Limit {
		continuePrincipalType := roleAssignments.Items[opts.Limit-1].Principal.Type
		continuePrincipalID := roleAssignments.Items[opts.Limit-1].Principal.ID
		continueRole := roleAssignments.Items[opts.Limit-1].Role
		continueScope := roleAssignments.Items[opts.Limit-1].Scope
		var continueValue = fmt.Sprintf(
			"%s:%s:%s:%s",
			continuePrincipalType,
			continuePrincipalID,
			continueRole,
			continueScope,
		)
		criteria["$or"] = []bson.M{
			// Same principal type, principal id, and role, but scopes we didn't list
			// yet
			{
				"principal.type": continuePrincipalType,
				"principal.id":   continuePrincipalID,
				"role":           continueRole,
				"scope":          bson.M{"$gt": continueScope},
			},
			// Same principal type and principal id, but roles and scopes we didn't
			// list yet
			{
				"principal.type": continuePrincipalType,
				"principal.id":   continuePrincipalID,
				"role":           bson.M{"$gt": continueRole},
			},
			// Same principal type, but principal ids, roles, and scopes we didn't
			// list yet
			{
				"principal.type": continuePrincipalType,
				"principal.id":   bson.M{"$gt": continuePrincipalID},
			},
			// Anything remaining that we didn't list yet
			{
				"principal.type": bson.M{"$gt": continuePrincipalType},
			},
		}
		remaining, err := r.collection.CountDocuments(ctx, criteria)
		if err != nil {
			return roleAssignments,
				errors.Wrap(err, "error counting remaining role assignments")
		}
		if remaining > 0 {
			roleAssignments.Continue = continueValue
			roleAssignments.RemainingItemCount = remaining
		}
	}

	return roleAssignments, nil
}

func (r *roleAssignmentsStore) Revoke(
	ctx context.Context,
	roleAssignment api.RoleAssignment,
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

func (r *roleAssignmentsStore) RevokeByPrincipal(
	ctx context.Context,
	principalReference api.PrincipalReference,
) error {
	if _, err := r.collection.DeleteMany(
		ctx,
		bson.M{"principal": principalReference},
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting role assignments for %s %q",
			principalReference.Type,
			principalReference.ID,
		)
	}
	return nil
}

func (r *roleAssignmentsStore) Exists(
	ctx context.Context,
	roleAssignment api.RoleAssignment,
) (bool, error) {
	criteria := bson.M{
		"role":           roleAssignment.Role,
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
