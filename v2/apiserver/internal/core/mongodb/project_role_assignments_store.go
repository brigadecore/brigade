package mongodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
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
		"role":           projectRoleAssignment.Role,
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

func (p *projectRoleAssignmentsStore) List(
	ctx context.Context,
	selector core.ProjectRoleAssignmentsSelector,
	opts meta.ListOptions,
) (core.ProjectRoleAssignmentList, error) {
	projectRoleAssignments := core.ProjectRoleAssignmentList{}

	criteria := bson.M{
		"projectID": selector.ProjectID,
	}
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
			return projectRoleAssignments, errors.New("error parsing continue value")
		}
		continueProjectID := tokens[0]
		continuePrincipalType := tokens[1]
		continuePrincipalID := tokens[2]
		continueRole := tokens[3]
		criteria["$or"] = []bson.M{
			// Same project id, principal type, and principal id, but roles we didn't
			// list yet
			{
				"projectID":      continueProjectID,
				"principal.type": continuePrincipalType,
				"principal.id":   continuePrincipalID,
				"role":           bson.M{"$gt": continueRole},
			},
			// Same project id and principal type, but principal ids and roles we
			// didn't list yet
			{
				"projectID":      continueProjectID,
				"principal.type": continuePrincipalType,
				"principal.id":   bson.M{"$gt": continuePrincipalID},
			},
			// Same project id, but principal types, principal ids, and roles we
			// didn't list yet
			{
				"projectID":      continueProjectID,
				"principal.type": bson.M{"$gt": continuePrincipalType},
			},
			// Anything remaining that we didn't list yet
			{
				"projectID": bson.M{"$gt": continueProjectID},
			},
		}
	}

	findOptions := options.Find()
	findOptions.SetSort(
		// bson.D preserves order, and we want to sort in this specific order
		bson.D{
			{Key: "projectID", Value: 1},
			{Key: "principal.type", Value: 1},
			{Key: "principal.id", Value: 1},
			{Key: "role", Value: 1},
		},
	)
	findOptions.SetLimit(opts.Limit)
	cur, err := p.collection.Find(ctx, criteria, findOptions)
	if err != nil {
		return projectRoleAssignments,
			errors.Wrap(err, "error finding project role assignments")
	}
	if err := cur.All(ctx, &projectRoleAssignments.Items); err != nil {
		return projectRoleAssignments,
			errors.Wrap(err, "error decoding project role assignments")
	}

	if int64(len(projectRoleAssignments.Items)) == opts.Limit {
		continueProjectID := projectRoleAssignments.Items[opts.Limit-1].ProjectID
		continuePrincipalType :=
			projectRoleAssignments.Items[opts.Limit-1].Principal.Type
		continuePrincipalID :=
			projectRoleAssignments.Items[opts.Limit-1].Principal.ID
		continueRole := projectRoleAssignments.Items[opts.Limit-1].Role
		var continueValue = fmt.Sprintf(
			"%s:%s:%s:%s",
			continueProjectID,
			continuePrincipalType,
			continuePrincipalID,
			continueRole,
		)
		criteria["$or"] = []bson.M{
			// Same project id, principal type, and principal id, but roles we didn't
			// list yet
			{
				"projectID":      continueProjectID,
				"principal.type": continuePrincipalType,
				"principal.id":   continuePrincipalID,
				"role":           bson.M{"$gt": continueRole},
			},
			// Same project id and principal type, but principal ids and roles we
			// didn't list yet
			{
				"projectID":      continueProjectID,
				"principal.type": continuePrincipalType,
				"principal.id":   bson.M{"$gt": continuePrincipalID},
			},
			// Same project id, but principal types, principal ids, and roles we
			// didn't list yet
			{
				"projectID":      continueProjectID,
				"principal.type": bson.M{"$gt": continuePrincipalType},
			},
			// Anything remaining that we didn't list yet
			{
				"projectID": bson.M{"$gt": continuePrincipalID},
			},
		}
		remaining, err := p.collection.CountDocuments(ctx, criteria)
		if err != nil {
			return projectRoleAssignments,
				errors.Wrap(err, "error counting remaining role assignments")
		}
		if remaining > 0 {
			projectRoleAssignments.Continue = continueValue
			projectRoleAssignments.RemainingItemCount = remaining
		}
	}

	return projectRoleAssignments, nil
}

func (p *projectRoleAssignmentsStore) Revoke(
	ctx context.Context,
	projectRoleAssignment core.ProjectRoleAssignment,
) error {
	criteria := bson.M{
		"projectID":      projectRoleAssignment.ProjectID,
		"role":           projectRoleAssignment.Role,
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
		"role":           projectRoleAssignment.Role,
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
