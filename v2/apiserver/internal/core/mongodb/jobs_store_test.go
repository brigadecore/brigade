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

func TestJobsStoreCancel(t *testing.T) {
	const testEvent = "123456789"
	const testJobName = "italian"
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(err error)
	}{
		// TODO: it'd be great to test the one bit of logic
		// around updating to canceled if pending and aborted if starting/running
		{
			name: "error getting job status",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					context.Context,
					interface{},
					...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(
						mongo.ErrNoDocuments,
					)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "unable to get status of job")
			},
		},
		{
			name: "error updating job status",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					context.Context,
					interface{},
					...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(
						core.Event{
							ObjectMeta: meta.ObjectMeta{
								ID: testEvent,
							},
							Worker: core.Worker{
								Jobs: map[string]core.Job{
									testJobName: {
										Status: &core.JobStatus{
											Phase: core.JobPhaseSucceeded,
										},
									},
								},
							},
						},
					)
					require.NoError(t, err)
					return res
				},
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
				require.Contains(t, err.Error(), "error updating status of worker job")
			},
		},
		{
			name: "success",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					context.Context,
					interface{},
					...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(
						core.Event{
							ObjectMeta: meta.ObjectMeta{
								ID: testEvent,
							},
							Worker: core.Worker{
								Jobs: map[string]core.Job{
									testJobName: {
										Status: &core.JobStatus{
											Phase: core.JobPhaseSucceeded,
										},
									},
								},
							},
						},
					)
					require.NoError(t, err)
					return res
				},
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
				store.Cancel(
					context.Background(),
					testEvent,
					testJobName,
				),
			)
		})
	}
}

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
				require.Contains(t, err.Error(), "error updating spec of event")
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
					testJobName,
					core.Job{},
				),
			)
		})
	}
}

func TestJobsStoreGetStatus(t *testing.T) {
	const testEvent = "123456789"
	const testJobName = "italian"
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(status core.JobStatus, err error)
	}{
		{
			name: "unanticipated error",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					context.Context,
					interface{},
					...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(
						errors.New("something went wrong"),
					)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(status core.JobStatus, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error finding/decoding event")
			},
		},
		{
			name: "event not found",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					context.Context,
					interface{},
					...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(
						mongo.ErrNoDocuments,
					)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(status core.JobStatus, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
			},
		},
		{
			name: "job not found",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					context.Context,
					interface{},
					...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(
						core.Event{
							ObjectMeta: meta.ObjectMeta{
								ID: testEvent,
							},
						},
					)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(status core.JobStatus, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
			},
		},
		{
			name: "success",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					context.Context,
					interface{},
					...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(
						core.Event{
							ObjectMeta: meta.ObjectMeta{
								ID: testEvent,
							},
							Worker: core.Worker{
								Jobs: map[string]core.Job{
									testJobName: {
										Status: &core.JobStatus{
											Phase: core.JobPhaseSucceeded,
										},
									},
								},
							},
						},
					)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(status core.JobStatus, err error) {
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
				store.GetStatus(
					context.Background(),
					testEvent,
					testJobName,
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
