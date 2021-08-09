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

func TestUsersStoreCreate(t *testing.T) {
	testUser := api.User{
		ObjectMeta: meta.ObjectMeta{
			ID: "tony@starkindustries.com",
		},
	}
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(err error)
	}{

		{
			name: "id already exists",
			collection: &mongoTesting.MockCollection{
				InsertOneFn: func(
					ctx context.Context,
					document interface{},
					opts ...*options.InsertOneOptions,
				) (*mongo.InsertOneResult, error) {
					return nil, mongoTesting.MockWriteException
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrConflict{}, err)
				require.Equal(t, "User", err.(*meta.ErrConflict).Type)
				require.Equal(t, testUser.ID, err.(*meta.ErrConflict).ID)
				require.Contains(t, err.(*meta.ErrConflict).Reason, "already exists")
			},
		},

		{
			name: "unanticipated error",
			collection: &mongoTesting.MockCollection{
				InsertOneFn: func(
					ctx context.Context,
					document interface{},
					opts ...*options.InsertOneOptions,
				) (*mongo.InsertOneResult, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error inserting new user")
			},
		},

		{
			name: "successful creation",
			collection: &mongoTesting.MockCollection{
				InsertOneFn: func(
					ctx context.Context,
					document interface{},
					opts ...*options.InsertOneOptions,
				) (*mongo.InsertOneResult, error) {
					return nil, nil
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &usersStore{
				collection: testCase.collection,
			}
			err := store.Create(context.Background(), testUser)
			testCase.assertions(err)
		})
	}
}

func TestUsersStoreList(t *testing.T) {
	testUser := api.User{
		ObjectMeta: meta.ObjectMeta{
			ID: "tony@starkindustries.com",
		},
	}

	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(api.UserList, error)
	}{

		{
			name: "error finding users",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(_ api.UserList, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error finding users")
			},
		},

		{
			name: "users found; no more pages of results exist",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cursor, err := mongoTesting.MockCursor(testUser)
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
			assertions: func(users api.UserList, err error) {
				require.NoError(t, err)
				require.Empty(t, users.Continue)
				require.Zero(t, users.RemainingItemCount)
			},
		},

		{
			name: "users found; more pages of results exist",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cursor, err := mongoTesting.MockCursor(testUser)
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
			assertions: func(users api.UserList, err error) {
				require.NoError(t, err)
				require.Equal(t, testUser.ID, users.Continue)
				require.Equal(t, int64(5), users.RemainingItemCount)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &usersStore{
				collection: testCase.collection,
			}
			users, err := store.List(
				context.Background(),
				meta.ListOptions{
					Limit:    1,
					Continue: "tony@starkindustries.com",
				},
			)
			testCase.assertions(users, err)
		})
	}
}

func TestUsersStoreGet(t *testing.T) {
	const testUserID = "tony@starkindustries.com"
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(user api.User, err error)
	}{

		{
			name: "user not found",
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
			assertions: func(user api.User, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
				require.Equal(t, "User", err.(*meta.ErrNotFound).Type)
				require.Equal(t, testUserID, err.(*meta.ErrNotFound).ID)
			},
		},

		{
			name: "unanticipated error",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(
						errors.New("something went wrong"),
					)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(user api.User, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error finding/decoding user")
			},
		},

		{
			name: "user found",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(
						api.User{
							ObjectMeta: meta.ObjectMeta{
								ID: testUserID,
							},
						},
					)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(user api.User, err error) {
				require.NoError(t, err)
				require.Equal(t, testUserID, user.ID)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &usersStore{
				collection: testCase.collection,
			}
			user, err := store.Get(context.Background(), testUserID)
			testCase.assertions(user, err)
		})
	}
}

func TestUsersStoreLock(t *testing.T) {
	const testUserID = "tony@starkindustries.com"

	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(error)
	}{
		{
			name: "user not found",
			collection: &mongoTesting.MockCollection{
				UpdateOneFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return &mongo.UpdateResult{MatchedCount: 0}, nil
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
				require.Equal(t, "User", err.(*meta.ErrNotFound).Type)
				require.Equal(t, testUserID, err.(*meta.ErrNotFound).ID)
			},
		},

		{
			name: "unanticipated error",
			collection: &mongoTesting.MockCollection{
				UpdateOneFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error updating user")
			},
		},

		{
			name: "success",
			collection: &mongoTesting.MockCollection{
				UpdateOneFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return &mongo.UpdateResult{MatchedCount: 1}, nil
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &usersStore{
				collection: testCase.collection,
			}
			err := store.Lock(context.Background(), testUserID)
			testCase.assertions(err)
		})
	}
}

func TestUsersStoreUnLock(t *testing.T) {
	const testUserID = "tony@starkindustries.com"

	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(error)
	}{
		{
			name: "user not found",
			collection: &mongoTesting.MockCollection{
				UpdateOneFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return &mongo.UpdateResult{MatchedCount: 0}, nil
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
				require.Equal(t, "User", err.(*meta.ErrNotFound).Type)
				require.Equal(t, testUserID, err.(*meta.ErrNotFound).ID)
			},
		},

		{
			name: "unanticipated error",
			collection: &mongoTesting.MockCollection{
				UpdateOneFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error updating user")
			},
		},

		{
			name: "success",
			collection: &mongoTesting.MockCollection{
				UpdateOneFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return &mongo.UpdateResult{MatchedCount: 1}, nil
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &usersStore{
				collection: testCase.collection,
			}
			err := store.Unlock(context.Background(), testUserID)
			testCase.assertions(err)
		})
	}
}
