package mongodb

import (
	"context"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	mongoTesting "github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb/testing" // nolint: lll
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestJobsStoreCreate(t *testing.T) {
	const testEvent = "123456789"
	const testJobName = "italian"
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
				require.Contains(t, err.Error(), "error creating event")
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
			store := &jobsStore{
				collection: testCase.collection,
			}
			testCase.assertions(
				store.Create(
					context.Background(),
					testEvent,
					core.Job{
						Name: testJobName,
					},
				),
			)
		})
	}
}

func TestJobsStoreDelete(t *testing.T) {
	const testEvent = "123456789"
	const testJobName = "italian"
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
				require.Contains(t, err.Error(), "error deleting event")
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
			store := &jobsStore{
				collection: testCase.collection,
			}
			testCase.assertions(
				store.Delete(
					context.Background(),
					testEvent,
					core.Job{
						Name: testJobName,
					},
				),
			)
		})
	}
}

func TestJobsStoreUpdateStatus(t *testing.T) {
	const testEvent = "123456789"
	const testJobName = "italian"
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
			store := &jobsStore{
				collection: testCase.collection,
			}
			err := store.UpdateStatus(
				context.Background(),
				testEvent,
				testJobName,
				core.JobStatus{},
			)
			testCase.assertions(err)
		})
	}
}
