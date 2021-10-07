package mongodb

import (
	"context"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	mongoTesting "github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb/testing" // nolint: lll
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestWorkersStoreUpdateStatus(t *testing.T) {
	const testEvent = "123456789"
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(err error)
	}{
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
				require.Contains(t, err.Error(), "error updating status of event")
			},
		},

		{
			name: "event not found",
			collection: &mongoTesting.MockCollection{
				UpdateOneFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return &mongo.UpdateResult{
						MatchedCount: 0,
					}, nil
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
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
			store := &workersStore{
				collection: testCase.collection,
			}
			err :=
				store.UpdateStatus(context.Background(), testEvent, api.WorkerStatus{})
			testCase.assertions(err)
		})
	}
}

func TestWorkersStoreUpdateHashedToken(t *testing.T) {
	const testEvent = "123456789"
	const testHashedToken = "a fake hashed token"
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(err error)
	}{
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
				require.Contains(t, err.Error(), "error updating event")
			},
		},

		{
			name: "event not found",
			collection: &mongoTesting.MockCollection{
				UpdateOneFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return &mongo.UpdateResult{
						MatchedCount: 0,
					}, nil
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
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
			store := &workersStore{
				collection: testCase.collection,
			}
			err := store.UpdateHashedToken(
				context.Background(),
				testEvent,
				testHashedToken,
			)
			testCase.assertions(err)
		})
	}
}

func TestWorkersStoreTimeout(t *testing.T) {
	const testEvent = "123456789"
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(err error)
	}{
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
				require.Contains(t, err.Error(), "error updating status of event")
			},
		},

		{
			name: "worker already in a terminal state",
			collection: &mongoTesting.MockCollection{
				UpdateOneFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return &mongo.UpdateResult{
						MatchedCount: 0,
					}, nil
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrConflict{}, err)
				require.Contains(t, err.Error(), "was not timed out")
				require.Contains(t, err.Error(), "not in a starting or running state")
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
			store := &workersStore{
				collection: testCase.collection,
			}
			err := store.Timeout(context.Background(), testEvent)
			testCase.assertions(err)
		})
	}
}
