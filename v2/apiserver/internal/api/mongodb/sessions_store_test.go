package mongodb

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	mongoTesting "github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb/testing" // nolint: lll
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestSessionsStoreCreate(t *testing.T) {
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(err error)
	}{

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
				require.Contains(t, err.Error(), "error inserting new session")
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
			store := &sessionsStore{
				collection: testCase.collection,
			}
			err := store.Create(context.Background(), api.Session{})
			testCase.assertions(err)
		})
	}
}

func TestSessionsStoreGetByHashedOAut2State(t *testing.T) {
	const testSessionID = "12345"
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(session api.Session, err error)
	}{

		{
			name: "session not found",
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
			assertions: func(session api.Session, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
				require.Equal(t, "Session", err.(*meta.ErrNotFound).Type)
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
			assertions: func(session api.Session, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error finding/decoding session")
			},
		},

		{
			name: "session found",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(
						api.Session{
							ObjectMeta: meta.ObjectMeta{
								ID: testSessionID,
							},
						},
					)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(session api.Session, err error) {
				require.NoError(t, err)
				require.Equal(t, testSessionID, session.ID)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &sessionsStore{
				collection: testCase.collection,
			}
			session, err := store.GetByHashedOAuth2State(
				context.Background(),
				"thisisafakeoaut2state",
			)
			testCase.assertions(session, err)
		})
	}
}

func TestSessionsStoreGetByHashedToken(t *testing.T) {
	const testSessionID = "12345"
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(session api.Session, err error)
	}{

		{
			name: "session not found",
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
			assertions: func(session api.Session, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
				require.Equal(t, "Session", err.(*meta.ErrNotFound).Type)
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
			assertions: func(session api.Session, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error finding/decoding session")
			},
		},

		{
			name: "session found",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(
						api.Session{
							ObjectMeta: meta.ObjectMeta{
								ID: testSessionID,
							},
						},
					)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(session api.Session, err error) {
				require.NoError(t, err)
				require.Equal(t, testSessionID, session.ID)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &sessionsStore{
				collection: testCase.collection,
			}
			session, err := store.GetByHashedToken(
				context.Background(),
				"thisisafakehashedtoken",
			)
			testCase.assertions(session, err)
		})
	}
}

func TestSessionsStoreAuthenticate(t *testing.T) {
	const (
		testSessionID = "12345"
		testUserID    = "tony@starkindustries.com"
	)
	var testExpiryTime = time.Now().UTC().Add(time.Hour)

	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(err error)
	}{

		{
			name: "session not found",
			collection: &mongoTesting.MockCollection{
				UpdateOneFn: func(
					ctx context.Context,
					filter interface{},
					update interface{},
					opts ...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return &mongo.UpdateResult{
						MatchedCount: 0,
					}, nil
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
				require.Equal(t, "Session", err.(*meta.ErrNotFound).Type)
				require.Equal(t, testSessionID, err.(*meta.ErrNotFound).ID)
			},
		},

		{
			name: "unanticipated error",
			collection: &mongoTesting.MockCollection{
				UpdateOneFn: func(
					ctx context.Context,
					filter interface{},
					update interface{},
					opts ...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error updating session")
			},
		},

		{
			name: "session found",
			collection: &mongoTesting.MockCollection{
				UpdateOneFn: func(
					ctx context.Context,
					filter interface{},
					update interface{},
					opts ...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return &mongo.UpdateResult{
						MatchedCount: 1,
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
			store := &sessionsStore{
				collection: testCase.collection,
			}
			err := store.Authenticate(
				context.Background(),
				testSessionID,
				testUserID,
				testExpiryTime,
			)
			testCase.assertions(err)
		})
	}
}

func TestSessionsStoreDelete(t *testing.T) {
	const testSessionID = "12345"

	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(err error)
	}{

		{
			name: "session not found",
			collection: &mongoTesting.MockCollection{
				DeleteOneFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.DeleteOptions,
				) (*mongo.DeleteResult, error) {
					return &mongo.DeleteResult{
						DeletedCount: 0,
					}, nil
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
				require.Equal(t, "Session", err.(*meta.ErrNotFound).Type)
				require.Equal(t, testSessionID, err.(*meta.ErrNotFound).ID)
			},
		},

		{
			name: "unanticipated error",
			collection: &mongoTesting.MockCollection{
				DeleteOneFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.DeleteOptions,
				) (*mongo.DeleteResult, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error deleting session")
			},
		},

		{
			name: "session found",
			collection: &mongoTesting.MockCollection{
				DeleteOneFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.DeleteOptions,
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
			store := &sessionsStore{
				collection: testCase.collection,
			}
			err := store.Delete(context.Background(), testSessionID)
			testCase.assertions(err)
		})
	}
}

func TestSessionsStoreDeleteByUser(t *testing.T) {
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(err error)
	}{

		{
			name: "unanticipated error",
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
				require.Contains(t, err.Error(), "error deleting sessions for user")
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
			store := &sessionsStore{
				collection: testCase.collection,
			}
			err := store.DeleteByUser(
				context.Background(),
				"tony@starkindustries.com",
			)
			testCase.assertions(err)
		})
	}
}
