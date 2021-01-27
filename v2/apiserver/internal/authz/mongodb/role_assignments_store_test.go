package mongodb

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	mongoTesting "github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb/testing" // nolint: lll
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestGrant(t *testing.T) {
	testRoleAssignment := authz.RoleAssignment{}
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(error)
	}{
		{
			name: "error",
			collection: &mongoTesting.MockCollection{
				FindOneAndReplaceFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.FindOneAndReplaceOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(
						errors.New("something went wrong"),
					)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error upserting role assignment")
			},
		},
		{
			name: "success",
			collection: &mongoTesting.MockCollection{
				FindOneAndReplaceFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.FindOneAndReplaceOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(testRoleAssignment)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &roleAssignmentsStore{
				collection: testCase.collection,
			}
			err := store.Grant(context.Background(), testRoleAssignment)
			testCase.assertions(err)
		})
	}
}

func TestRevoke(t *testing.T) {
	testRoleAssignment := authz.RoleAssignment{}
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(error)
	}{
		{
			name: "error",
			collection: &mongoTesting.MockCollection{
				DeleteOneFn: func(
					context.Context,
					interface{},
					...*options.DeleteOptions,
				) (*mongo.DeleteResult, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error deleting role assignment")
			},
		},
		{
			name: "success",
			collection: &mongoTesting.MockCollection{
				DeleteOneFn: func(
					context.Context,
					interface{},
					...*options.DeleteOptions,
				) (*mongo.DeleteResult, error) {
					return &mongo.DeleteResult{
						DeletedCount: 1,
					}, nil
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &roleAssignmentsStore{
				collection: testCase.collection,
			}
			err := store.Revoke(context.Background(), testRoleAssignment)
			testCase.assertions(err)
		})
	}
}
