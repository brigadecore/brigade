package mongodb

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authz"
	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	mongoTesting "github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb/testing" // nolint: lll
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestRoleAssignmentsStoreGrant(t *testing.T) {
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

func TestRoleAssignmentsStoreList(t *testing.T) {
	testRoleAssignment := libAuthz.RoleAssignment{
		Principal: libAuthz.PrincipalReference{
			Type: authz.PrincipalTypeUser,
			ID:   "tony@starkindustries.com",
		},
		Role:  libAuthz.Role("ceo"),
		Scope: "corporate",
	}

	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(authz.RoleAssignmentList, error)
	}{

		{
			name: "error finding role assignments",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(_ authz.RoleAssignmentList, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error finding role assignments")
			},
		},

		{
			name: "role assignments found; no more pages of results exist",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cursor, err := mongoTesting.MockCursor(testRoleAssignment)
					require.NoError(t, err)
					return cursor, nil
				},
				CountDocumentsFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.CountOptions,
				) (int64, error) {
					return 0, nil
				},
			},
			assertions: func(roleAssignments authz.RoleAssignmentList, err error) {
				require.NoError(t, err)
				require.Empty(t, roleAssignments.Continue)
				require.Zero(t, roleAssignments.RemainingItemCount)
			},
		},

		{
			name: "role assignments found; more pages of results exist",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cursor, err := mongoTesting.MockCursor(testRoleAssignment)
					require.NoError(t, err)
					return cursor, nil
				},
				CountDocumentsFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.CountOptions,
				) (int64, error) {
					return 5, nil
				},
			},
			assertions: func(roleAssignments authz.RoleAssignmentList, err error) {
				require.NoError(t, err)
				require.Equal(
					t,
					"USER:tony@starkindustries.com:ceo:corporate",
					roleAssignments.Continue,
				)
				require.Equal(t, int64(5), roleAssignments.RemainingItemCount)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &roleAssignmentsStore{
				collection: testCase.collection,
			}
			roleAssignments, err := store.List(
				context.Background(),
				authz.RoleAssignmentsSelector{},
				meta.ListOptions{
					Limit: 1,
				},
			)
			testCase.assertions(roleAssignments, err)
		})
	}
}

func TestRoleAssignmentsStoreRevoke(t *testing.T) {
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
