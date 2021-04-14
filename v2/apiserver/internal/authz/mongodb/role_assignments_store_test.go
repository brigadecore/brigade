package mongodb

import (
	"context"
	"errors"
	"testing"

	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	mongoTesting "github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb/testing" // nolint: lll
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestGrant(t *testing.T) {
	testRoleAssignment := libAuthz.RoleAssignment{}
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
	testRoleAssignment := libAuthz.RoleAssignment{}
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

func TestRevokeMany(t *testing.T) {
	testRoleAssignment := libAuthz.RoleAssignment{}
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(error)
	}{
		{
			name: "error",
			collection: &mongoTesting.MockCollection{
				DeleteManyFn: func(
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
				require.Contains(t, err.Error(), "error deleting role assignments")
			},
		},
		{
			name: "success",
			collection: &mongoTesting.MockCollection{
				DeleteManyFn: func(
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
			err := store.RevokeMany(context.Background(), testRoleAssignment)
			testCase.assertions(err)
		})
	}
}

func TestExists(t *testing.T) {
	testRoleAssignment := libAuthz.RoleAssignment{}
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(exists bool, err error)
	}{
		{
			name: "not found",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(mongo.ErrNoDocuments)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(exists bool, err error) {
				require.False(t, exists)
				require.NoError(t, err)
			},
		},
		{
			name: "error",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err :=
						mongoTesting.MockSingleResult(errors.New("something went wrong"))
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(exists bool, err error) {
				require.False(t, exists)
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error finding role assignment")
			},
		},
		{
			name: "success",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(testRoleAssignment)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(exists bool, err error) {
				require.True(t, exists)
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &roleAssignmentsStore{
				collection: testCase.collection,
			}
			exists, err := store.Exists(context.Background(), testRoleAssignment)
			testCase.assertions(exists, err)
		})
	}
}
