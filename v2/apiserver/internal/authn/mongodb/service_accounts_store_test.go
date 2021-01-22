package mongodb

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	mongoTesting "github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb/testing" // nolint: lll
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestServiceAccountsStoreCreate(t *testing.T) {
	testServiceAccount := authn.ServiceAccount{
		ObjectMeta: meta.ObjectMeta{
			ID: "jarvis",
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
				require.Equal(t, "ServiceAccount", err.(*meta.ErrConflict).Type)
				require.Equal(t, testServiceAccount.ID, err.(*meta.ErrConflict).ID)
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
				require.Contains(t, err.Error(), "error inserting new service account")
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
			store := &serviceAccountsStore{
				collection: testCase.collection,
			}
			err := store.Create(context.Background(), testServiceAccount)
			testCase.assertions(err)
		})
	}
}

func TestServiceAccountsStoreList(t *testing.T) {
	testServiceAccount := authn.ServiceAccount{
		ObjectMeta: meta.ObjectMeta{
			ID: "jarvis",
		},
	}

	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(serviceAccounts authn.ServiceAccountList, err error)
	}{

		{
			name: "error finding service accounts",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(_ authn.ServiceAccountList, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error finding service accounts")
			},
		},

		{
			name: "service accounts found; no more pages of results exist",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cursor, err := mongoTesting.MockCursor(testServiceAccount)
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
			assertions: func(serviceAccounts authn.ServiceAccountList, err error) {
				require.NoError(t, err)
				require.Empty(t, serviceAccounts.Continue)
				require.Zero(t, serviceAccounts.RemainingItemCount)
			},
		},

		{
			name: "service accounts found; more pages of results exist",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cursor, err := mongoTesting.MockCursor(testServiceAccount)
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
			assertions: func(serviceAccounts authn.ServiceAccountList, err error) {
				require.NoError(t, err)
				require.Equal(t, testServiceAccount.ID, serviceAccounts.Continue)
				require.Equal(t, int64(5), serviceAccounts.RemainingItemCount)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &serviceAccountsStore{
				collection: testCase.collection,
			}
			serviceAccounts, err := store.List(
				context.Background(),
				meta.ListOptions{
					Limit:    1,
					Continue: "blue-book",
				},
			)
			testCase.assertions(serviceAccounts, err)
		})
	}
}

func TestServiceAccountsStoreGet(t *testing.T) {
	const testServiceAccountID = "jarvis"
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(authn.ServiceAccount, error)
	}{

		{
			name: "service account not found",
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
			assertions: func(_ authn.ServiceAccount, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
				require.Equal(t, "ServiceAccount", err.(*meta.ErrNotFound).Type)
				require.Equal(t, testServiceAccountID, err.(*meta.ErrNotFound).ID)
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
			assertions: func(_ authn.ServiceAccount, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(
					t,
					err.Error(),
					"error finding/decoding service account",
				)
			},
		},

		{
			name: "service account found",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(
						authn.ServiceAccount{
							ObjectMeta: meta.ObjectMeta{
								ID: testServiceAccountID,
							},
						},
					)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(serviceAccount authn.ServiceAccount, err error) {
				require.NoError(t, err)
				require.Equal(t, testServiceAccountID, serviceAccount.ID)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &serviceAccountsStore{
				collection: testCase.collection,
			}
			serviceAccount, err :=
				store.Get(context.Background(), testServiceAccountID)
			testCase.assertions(serviceAccount, err)
		})
	}
}

func TestServiceAccountsStoreGetByHashedToken(t *testing.T) {
	const testServiceAccountID = "jarvis"
	const testHashedToken = "abcdefghijklmnopqrstuvwxyz"
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(authn.ServiceAccount, error)
	}{

		{
			name: "service account not found",
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
			assertions: func(_ authn.ServiceAccount, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
				require.Equal(t, "ServiceAccount", err.(*meta.ErrNotFound).Type)
				require.Empty(t, err.(*meta.ErrNotFound).ID)
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
			assertions: func(_ authn.ServiceAccount, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(
					t,
					err.Error(),
					"error finding/decoding service account by hashed token",
				)
			},
		},

		{
			name: "service account found",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(
						authn.ServiceAccount{
							ObjectMeta: meta.ObjectMeta{
								ID: testServiceAccountID,
							},
						},
					)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(serviceAccount authn.ServiceAccount, err error) {
				require.NoError(t, err)
				require.Equal(t, testServiceAccountID, serviceAccount.ID)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &serviceAccountsStore{
				collection: testCase.collection,
			}
			serviceAccount, err :=
				store.GetByHashedToken(context.Background(), testHashedToken)
			testCase.assertions(serviceAccount, err)
		})
	}
}

func TestServiceAccountsLock(t *testing.T) {
	const testServiceAccountID = "jarvis"

	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(error)
	}{
		{
			name: "service account not found",
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
				require.Equal(t, "ServiceAccount", err.(*meta.ErrNotFound).Type)
				require.Equal(t, testServiceAccountID, err.(*meta.ErrNotFound).ID)
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
				require.Contains(t, err.Error(), "error updating service account")
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
			store := &serviceAccountsStore{
				collection: testCase.collection,
			}
			err := store.Lock(context.Background(), testServiceAccountID)
			testCase.assertions(err)
		})
	}
}

func TestServiceAccountsUnLock(t *testing.T) {
	const testServiceAccountID = "jarvis"

	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(error)
	}{
		{
			name: "service account not found",
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
				require.Equal(t, "ServiceAccount", err.(*meta.ErrNotFound).Type)
				require.Equal(t, testServiceAccountID, err.(*meta.ErrNotFound).ID)
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
				require.Contains(t, err.Error(), "error updating service account")
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
			store := &serviceAccountsStore{
				collection: testCase.collection,
			}
			err := store.Unlock(
				context.Background(),
				testServiceAccountID,
				"123456789", // New hashed-token
			)
			testCase.assertions(err)
		})
	}
}
