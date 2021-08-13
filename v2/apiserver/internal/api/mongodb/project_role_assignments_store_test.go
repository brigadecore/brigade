package mongodb

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	mongoTesting "github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb/testing" // nolint: lll
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestProjectRoleAssignmentStoreGrant(t *testing.T) {
	testProjectRoleAssignment := api.ProjectRoleAssignment{}
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
				require.Contains(
					t,
					err.Error(),
					"error upserting project role assignment",
				)
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
					res, err := mongoTesting.MockSingleResult(testProjectRoleAssignment)
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
			store := &projectRoleAssignmentsStore{
				collection: testCase.collection,
			}
			err := store.Grant(context.Background(), testProjectRoleAssignment)
			testCase.assertions(err)
		})
	}
}

func TestProjectRoleAssignmentsStoreList(t *testing.T) {
	testProjectRoleAssignment := api.ProjectRoleAssignment{
		ProjectID: "avengers-initiative",
		Principal: api.PrincipalReference{
			Type: api.PrincipalTypeUser,
			ID:   "tony@starkindustries.com",
		},
		Role: api.Role("ceo"),
	}

	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(api.ProjectRoleAssignmentList, error)
	}{

		{
			name: "error finding project role assignments",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(_ api.ProjectRoleAssignmentList, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(
					t,
					err.Error(),
					"error finding project role assignments",
				)
			},
		},

		{
			name: "project role assignments found; no more pages of results exist",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cursor, err := mongoTesting.MockCursor(testProjectRoleAssignment)
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
			assertions: func(
				projectRoleAssignments api.ProjectRoleAssignmentList,
				err error,
			) {
				require.NoError(t, err)
				require.Empty(t, projectRoleAssignments.Continue)
				require.Zero(t, projectRoleAssignments.RemainingItemCount)
			},
		},

		{
			name: "project role assignments found; more pages of results exist",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cursor, err := mongoTesting.MockCursor(testProjectRoleAssignment)
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
			assertions: func(
				projectRoleAssignments api.ProjectRoleAssignmentList,
				err error,
			) {
				require.NoError(t, err)
				require.Equal(
					t,
					"avengers-initiative:USER:tony@starkindustries.com:ceo",
					projectRoleAssignments.Continue,
				)
				require.Equal(t, int64(5), projectRoleAssignments.RemainingItemCount)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &projectRoleAssignmentsStore{
				collection: testCase.collection,
			}
			roleAssignments, err := store.List(
				context.Background(),
				api.ProjectRoleAssignmentsSelector{},
				meta.ListOptions{
					Limit: 1,
				},
			)
			testCase.assertions(roleAssignments, err)
		})
	}
}

func TestProjectRoleAssignmentStoreRevoke(t *testing.T) {
	testProjectRoleAssignment := api.ProjectRoleAssignment{}
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
				require.Contains(
					t,
					err.Error(),
					"error deleting project role assignment",
				)
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
			store := &projectRoleAssignmentsStore{
				collection: testCase.collection,
			}
			err := store.Revoke(context.Background(), testProjectRoleAssignment)
			testCase.assertions(err)
		})
	}
}

func TestProjectRoleAssignmentStoreRevokeByProjectID(t *testing.T) {
	const testProjectID = "bluebook"
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
				require.Contains(
					t,
					err.Error(),
					"error deleting project role assignments",
				)
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
			store := &projectRoleAssignmentsStore{
				collection: testCase.collection,
			}
			err := store.RevokeByProjectID(context.Background(), testProjectID)
			testCase.assertions(err)
		})
	}
}

func TestProjectRoleAssignmentStoreRevokeByPrincipal(t *testing.T) {
	testPrincipalReference := api.PrincipalReference{
		Type: api.PrincipalTypeUser,
		ID:   "tony@starkindustries.com",
	}
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
				require.Contains(
					t,
					err.Error(),
					"error deleting project role assignments",
				)
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
			store := &projectRoleAssignmentsStore{
				collection: testCase.collection,
			}
			err := store.RevokeByPrincipal(
				context.Background(),
				testPrincipalReference,
			)
			testCase.assertions(err)
		})
	}
}

func TestProjectRoleAssignmentStoreExists(t *testing.T) {
	testProjectRoleAssignment := api.ProjectRoleAssignment{}
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
				require.Contains(
					t,
					err.Error(),
					"error finding project role assignment",
				)
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
					res, err := mongoTesting.MockSingleResult(testProjectRoleAssignment)
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
			store := &projectRoleAssignmentsStore{
				collection: testCase.collection,
			}
			exists, err :=
				store.Exists(context.Background(), testProjectRoleAssignment)
			testCase.assertions(exists, err)
		})
	}
}
