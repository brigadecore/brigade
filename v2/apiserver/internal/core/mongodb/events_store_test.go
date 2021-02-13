package mongodb

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	mongoTesting "github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb/testing" // nolint: lll
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestEventsStoreCreate(t *testing.T) {
	testEvent := core.Event{
		ObjectMeta: meta.ObjectMeta{
			ID: "foo",
		},
	}
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(err error)
	}{

		{
			name: "error inserting event",
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
				require.Contains(t, err.Error(), "error inserting new event")
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
			store := &eventsStore{
				collection: testCase.collection,
			}
			err := store.Create(context.Background(), testEvent)
			testCase.assertions(err)
		})
	}
}

func TestEventsStoreList(t *testing.T) {
	const testProjectID = "blue-book"
	now := time.Now().UTC()
	testEvent := core.Event{
		ObjectMeta: meta.ObjectMeta{
			ID:      "foo",
			Created: &now,
		},
	}
	testCases := []struct {
		name        string
		listOptions meta.ListOptions
		collection  mongodb.Collection
		assertions  func(events core.EventList, err error)
	}{
		{
			name: "unparsable continue value",
			listOptions: meta.ListOptions{
				Continue: "invalid time",
			},
			assertions: func(events core.EventList, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error parsing continue time")
			},
		},

		{
			name: "error finding events",
			listOptions: meta.ListOptions{
				// This continue value is formatted correctly, but its value is totally
				// made up and that is totally ok because we're going to force the
				// results we want to be returned regardless of what this value is.
				Continue: fmt.Sprintf("%d:foo", time.Now().UTC().UnixNano()),
			},
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					context.Context,
					interface{},
					...*options.FindOptions,
				) (*mongo.Cursor, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(events core.EventList, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error finding events")
			},
		},

		{
			name: "events found; no more pages of results exist",
			listOptions: meta.ListOptions{
				Limit: 1,
			},
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cursor, err := mongoTesting.MockCursor(testEvent)
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
			assertions: func(events core.EventList, err error) {
				require.NoError(t, err)
				require.Len(t, events.Items, 1)
				require.Equal(t, testEvent.ID, events.Items[0].ID)
				require.Empty(t, events.Continue)
				require.Zero(t, events.RemainingItemCount)
			},
		},

		{
			name: "events found; more pages of results exist",
			listOptions: meta.ListOptions{
				Limit: 1,
			},
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cursor, err := mongoTesting.MockCursor(testEvent)
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
			assertions: func(events core.EventList, err error) {
				require.NoError(t, err)
				require.Len(t, events.Items, 1)
				require.Equal(t, testEvent.ID, events.Items[0].ID)
				require.Equal(
					t,
					fmt.Sprintf(
						"%d:%s",
						events.Items[0].Created.UnixNano(),
						testEvent.ID,
					),
					events.Continue,
				)
				require.Equal(t, int64(5), events.RemainingItemCount)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &eventsStore{
				collection: testCase.collection,
			}
			events, err := store.List(
				context.Background(),
				core.EventsSelector{
					ProjectID: testProjectID,
				},
				testCase.listOptions,
			)
			testCase.assertions(events, err)
		})
	}
}

func TestEventsStoreGet(t *testing.T) {
	const testEventID = "123456789"
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(event core.Event, err error)
	}{
		{
			name: "event not found",
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
			assertions: func(_ core.Event, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
				require.Equal(t, "Event", err.(*meta.ErrNotFound).Type)
				require.Equal(t, testEventID, err.(*meta.ErrNotFound).ID)
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
			assertions: func(_ core.Event, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error finding/decoding event")
			},
		},

		{
			name: "event found",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(
						core.Event{
							ObjectMeta: meta.ObjectMeta{
								ID: testEventID,
							},
						},
					)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(event core.Event, err error) {
				require.NoError(t, err)
				require.Equal(t, testEventID, event.ID)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &eventsStore{
				collection: testCase.collection,
			}
			event, err := store.Get(context.Background(), testEventID)
			testCase.assertions(event, err)
		})
	}
}

func TestEventsStoreGetByHashedToken(t *testing.T) {
	const testEventID = "123456789"
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(event core.Event, err error)
	}{
		{
			name: "event not found",
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
			assertions: func(_ core.Event, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
				require.Equal(t, "Event", err.(*meta.ErrNotFound).Type)
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
			assertions: func(_ core.Event, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error finding/decoding event")
			},
		},

		{
			name: "event found",
			collection: &mongoTesting.MockCollection{
				FindOneFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mongoTesting.MockSingleResult(
						core.Event{
							ObjectMeta: meta.ObjectMeta{
								ID: testEventID,
							},
						},
					)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(event core.Event, err error) {
				require.NoError(t, err)
				require.Equal(t, testEventID, event.ID)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &eventsStore{
				collection: testCase.collection,
			}
			event, err :=
				store.GetByHashedWorkerToken(context.Background(), "abcdefg")
			testCase.assertions(event, err)
		})
	}
}

func TestEventsStoreCancel(t *testing.T) {
	const testEventID = "abcedfg"
	testCases := []struct {
		name       string
		setup      func() mongodb.Collection
		assertions func(err error)
	}{
		{
			name: "error updating event with pending worker",
			setup: func() mongodb.Collection {
				return &mongoTesting.MockCollection{
					UpdateOneFn: func(
						context.Context,
						interface{},
						interface{},
						...*options.UpdateOptions,
					) (*mongo.UpdateResult, error) {
						return nil, errors.New("something went wrong")
					},
				}
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error updating status of event")
				require.Contains(t, err.Error(), "something went wrong")
			},
		},

		{
			name: "one event with pending worker updated",
			setup: func() mongodb.Collection {
				return &mongoTesting.MockCollection{
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
				}
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},

		{
			name: "error updating event with running worker",
			setup: func() mongodb.Collection {
				once := &sync.Once{}
				return &mongoTesting.MockCollection{
					UpdateOneFn: func(
						context.Context,
						interface{},
						interface{},
						...*options.UpdateOptions,
					) (*mongo.UpdateResult, error) {
						var res *mongo.UpdateResult
						// This is just a fancy way to make UpdateOne update nothing the
						// first time it's called (attempting to update a pending event),
						// but have a different result (an error) on a subsequent call to
						// UpdateOne (attempting to update a running event).
						once.Do(func() {
							res = &mongo.UpdateResult{
								MatchedCount: 0,
							}
						})
						if res != nil {
							return res, nil
						}
						return nil, errors.New("something went wrong")
					},
				}
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error updating status of event")
				require.Contains(t, err.Error(), "something went wrong")
			},
		},

		{
			name: "no event updated",
			setup: func() mongodb.Collection {
				return &mongoTesting.MockCollection{
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
				}
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"was not canceled because it was already in a terminal state",
				)
			},
		},

		{
			name: "one event with running worker updated",
			setup: func() mongodb.Collection {
				once := &sync.Once{}
				return &mongoTesting.MockCollection{
					UpdateOneFn: func(
						context.Context,
						interface{},
						interface{},
						...*options.UpdateOptions,
					) (*mongo.UpdateResult, error) {
						var res *mongo.UpdateResult
						// This is just a fancy way to make UpdateOne update nothing the
						// first time it's called (attempting to update a pending event),
						// but have a different result (an error) on a subsequent call to
						// UpdateOne (attempting to update a running event).
						once.Do(func() {
							res = &mongo.UpdateResult{
								MatchedCount: 0,
							}
						})
						if res != nil {
							return res, nil
						}
						return &mongo.UpdateResult{
							MatchedCount: 1,
						}, nil
					},
				}
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &eventsStore{
				collection: testCase.setup(),
			}
			err := store.Cancel(context.Background(), testEventID)
			testCase.assertions(err)
		})
	}
}

func TestEventsStoreCancelMany(t *testing.T) {
	testCases := []struct {
		name           string
		eventsSelector core.EventsSelector
		collection     mongodb.Collection
		assertions     func(err error)
	}{

		{
			name:           "requested neither pending nor running events canceled",
			eventsSelector: core.EventsSelector{},
			collection: &mongoTesting.MockCollection{
				UpdateManyFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					require.Fail(t, "UpdateMany should NOT have been called")
					return nil, nil
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},

		{
			name: "error updating events with pending workers",
			eventsSelector: core.EventsSelector{
				WorkerPhases: []core.WorkerPhase{
					core.WorkerPhasePending,
				},
			},
			collection: &mongoTesting.MockCollection{
				UpdateManyFn: func(
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
				require.Contains(t, err.Error(), "error updating events")
				require.Contains(t, err.Error(), "something went wrong")
			},
		},

		{
			name: "error updating events with running workers",
			eventsSelector: core.EventsSelector{
				WorkerPhases: []core.WorkerPhase{
					core.WorkerPhaseRunning,
				},
			},
			collection: &mongoTesting.MockCollection{
				UpdateManyFn: func(
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
				require.Contains(t, err.Error(), "error updating events")
				require.Contains(t, err.Error(), "something went wrong")
			},
		},

		{
			name: "error counting canceled events",
			eventsSelector: core.EventsSelector{
				WorkerPhases: []core.WorkerPhase{
					core.WorkerPhasePending,
					core.WorkerPhaseRunning,
				},
			},
			collection: &mongoTesting.MockCollection{
				UpdateManyFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return &mongo.UpdateResult{}, nil
				},
				CountDocumentsFn: func(
					context.Context,
					interface{},
					...*options.CountOptions,
				) (int64, error) {
					return 0, errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error counting canceled events")
				require.Contains(t, err.Error(), "something went wrong")
			},
		},

		{
			name: "error finding canceled events",
			eventsSelector: core.EventsSelector{
				WorkerPhases: []core.WorkerPhase{
					core.WorkerPhasePending,
					core.WorkerPhaseRunning,
				},
			},
			collection: &mongoTesting.MockCollection{
				UpdateManyFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return &mongo.UpdateResult{}, nil
				},
				CountDocumentsFn: func(
					context.Context,
					interface{},
					...*options.CountOptions,
				) (int64, error) {
					return 0, nil
				},
				FindFn: func(
					context.Context,
					interface{},
					...*options.FindOptions,
				) (*mongo.Cursor, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error finding canceled events")
				require.Contains(t, err.Error(), "something went wrong")
			},
		},

		{
			name: "success",
			eventsSelector: core.EventsSelector{
				WorkerPhases: []core.WorkerPhase{
					core.WorkerPhasePending,
					core.WorkerPhaseRunning,
				},
			},
			collection: &mongoTesting.MockCollection{
				UpdateManyFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return &mongo.UpdateResult{}, nil
				},
				CountDocumentsFn: func(
					context.Context,
					interface{},
					...*options.CountOptions,
				) (int64, error) {
					return 0, nil
				},
				FindFn: func(
					context.Context,
					interface{},
					...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cursor, err := mongoTesting.MockCursor(core.Event{})
					require.NoError(t, err)
					return cursor, nil
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &eventsStore{
				collection: testCase.collection,
			}
			_, _, error :=
				store.CancelMany(context.Background(), testCase.eventsSelector)
			testCase.assertions(error)
		})
	}
}

func TestEventsStoreDelete(t *testing.T) {
	const testEventID = "qrstuvwxyz"
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(err error)
	}{

		{
			name: "error deleting event",
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
				require.Contains(t, err.Error(), "error deleting event")
				require.Contains(t, err.Error(), "something went wrong")
			},
		},

		{
			name: "event not found",
			collection: &mongoTesting.MockCollection{
				DeleteOneFn: func(
					context.Context,
					interface{},
					...*options.DeleteOptions,
				) (*mongo.DeleteResult, error) {
					return &mongo.DeleteResult{
						DeletedCount: 0,
					}, nil
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
				require.Equal(t, "Event", err.(*meta.ErrNotFound).Type)
				require.Equal(t, testEventID, err.(*meta.ErrNotFound).ID)
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
			store := &eventsStore{
				collection: testCase.collection,
			}
			err := store.Delete(context.Background(), testEventID)
			testCase.assertions(err)
		})
	}
}

func TestEventsStoreDeleteMany(t *testing.T) {
	testCases := []struct {
		name           string
		eventsSelector core.EventsSelector
		collection     mongodb.Collection
		assertions     func(error)
	}{

		{
			name: "error updating events",
			eventsSelector: core.EventsSelector{
				WorkerPhases: []core.WorkerPhase{
					core.WorkerPhasePending,
				},
			},
			collection: &mongoTesting.MockCollection{
				UpdateManyFn: func(
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
				require.Contains(t, err.Error(), "error logically deleting events")
				require.Contains(t, err.Error(), "something went wrong")
			},
		},

		{
			name: "error counting deleted events",
			eventsSelector: core.EventsSelector{
				WorkerPhases: []core.WorkerPhase{
					core.WorkerPhasePending,
					core.WorkerPhaseRunning,
				},
			},
			collection: &mongoTesting.MockCollection{
				UpdateManyFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return &mongo.UpdateResult{}, nil
				},
				CountDocumentsFn: func(
					context.Context,
					interface{},
					...*options.CountOptions,
				) (int64, error) {
					return 0, errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"error counting deleted events",
				)
				require.Contains(t, err.Error(), "something went wrong")
			},
		},

		{
			name: "error finding deleted events",
			eventsSelector: core.EventsSelector{
				WorkerPhases: []core.WorkerPhase{
					core.WorkerPhasePending,
					core.WorkerPhaseRunning,
				},
			},
			collection: &mongoTesting.MockCollection{
				UpdateManyFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return &mongo.UpdateResult{}, nil
				},
				CountDocumentsFn: func(
					context.Context,
					interface{},
					...*options.CountOptions,
				) (int64, error) {
					return 0, nil
				},
				FindFn: func(
					context.Context,
					interface{},
					...*options.FindOptions,
				) (*mongo.Cursor, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"error finding logically deleted events",
				)
				require.Contains(t, err.Error(), "something went wrong")
			},
		},

		{
			name: "success",
			eventsSelector: core.EventsSelector{
				WorkerPhases: []core.WorkerPhase{
					core.WorkerPhasePending,
					core.WorkerPhaseRunning,
				},
			},
			collection: &mongoTesting.MockCollection{
				UpdateManyFn: func(
					context.Context,
					interface{},
					interface{},
					...*options.UpdateOptions,
				) (*mongo.UpdateResult, error) {
					return &mongo.UpdateResult{}, nil
				},
				CountDocumentsFn: func(
					context.Context,
					interface{},
					...*options.CountOptions,
				) (int64, error) {
					return 0, nil
				},
				FindFn: func(
					context.Context,
					interface{},
					...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cur, err := mongoTesting.MockCursor(core.Event{})
					require.NoError(t, err)
					return cur, nil
				},
				DeleteManyFn: func(
					context.Context,
					interface{},
					...*options.DeleteOptions,
				) (*mongo.DeleteResult, error) {
					return &mongo.DeleteResult{}, nil
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &eventsStore{
				collection: testCase.collection,
			}
			_, _, err :=
				store.DeleteMany(context.Background(), testCase.eventsSelector)
			testCase.assertions(err)
		})
	}
}
