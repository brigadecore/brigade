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

func TestProjectsStoreCreate(t *testing.T) {
	testProject := api.Project{
		ObjectMeta: meta.ObjectMeta{
			ID: "blue-book",
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
				ec, ok := err.(*meta.ErrConflict)
				require.True(t, ok)
				require.Equal(t, api.ProjectKind, ec.Type)
				require.Equal(t, testProject.ID, ec.ID)
				require.Contains(t, ec.Reason, "already exists")
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
				require.Contains(t, err.Error(), "error inserting new project")
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
			store := &projectsStore{
				collection: testCase.collection,
			}
			err := store.Create(context.Background(), testProject)
			testCase.assertions(err)
		})
	}
}

func TestProjectsStoreList(t *testing.T) {
	testProject := api.Project{
		ObjectMeta: meta.ObjectMeta{
			ID: "blue-book",
		},
	}

	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(projects meta.List[api.Project], err error)
	}{

		{
			name: "error finding projects",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(projects meta.List[api.Project], err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error finding projects")
			},
		},

		{
			name: "projects found; no more pages of results exist",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cursor, err := mongoTesting.MockCursor(testProject)
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
			assertions: func(projects meta.List[api.Project], err error) {
				require.NoError(t, err)
				require.Empty(t, projects.Continue)
				require.Zero(t, projects.RemainingItemCount)
			},
		},

		{
			name: "projects found; more pages of results exist",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cursor, err := mongoTesting.MockCursor(testProject)
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
			assertions: func(projects meta.List[api.Project], err error) {
				require.NoError(t, err)
				require.Equal(t, testProject.ID, projects.Continue)
				require.Equal(t, int64(5), projects.RemainingItemCount)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &projectsStore{
				collection: testCase.collection,
			}
			projects, err := store.List(
				context.Background(),
				meta.ListOptions{
					Limit:    1,
					Continue: "blue-book",
				},
			)
			testCase.assertions(projects, err)
		})
	}
}

func TestProjectsStoreListSubscribers(t *testing.T) {
	testProject1 := api.Project{
		ObjectMeta: meta.ObjectMeta{
			ID: "project1",
		},
	}
	testProject2 := api.Project{
		ObjectMeta: meta.ObjectMeta{
			ID: "project2",
		},
	}
	testEvent := api.Event{
		Source: "github.com/krancour/fake-gateway",
		Type:   "push",
		Qualifiers: api.Qualifiers{
			"foo": "bar",
			"bat": "baz",
		},
	}
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(subscribers meta.List[api.Project], err error)
	}{
		{
			name: "error finding subscribers",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(subscribers meta.List[api.Project], err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error finding projects")
			},
		},

		{
			name: "found no subscribers",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cursor, err := mongoTesting.MockCursor()
					require.NoError(t, err)
					return cursor, nil
				},
			},
			assertions: func(subscribers meta.List[api.Project], err error) {
				require.NoError(t, err)
				require.Empty(t, subscribers.Items)
			},
		},

		{
			name: "found subscribers",
			collection: &mongoTesting.MockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cursor, err := mongoTesting.MockCursor(testProject1, testProject2)
					require.NoError(t, err)
					return cursor, nil
				},
			},
			assertions: func(subscribers meta.List[api.Project], err error) {
				require.NoError(t, err)
				require.Len(t, subscribers.Items, 2)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &projectsStore{
				collection: testCase.collection,
			}
			subscribers, err :=
				store.ListSubscribers(context.Background(), testEvent)
			testCase.assertions(subscribers, err)
		})
	}
}

func TestProjectsStoreGet(t *testing.T) {
	const testProjectID = "blue-book"
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(project api.Project, err error)
	}{

		{
			name: "project not found",
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
			assertions: func(project api.Project, err error) {
				require.Error(t, err)
				enf, ok := err.(*meta.ErrNotFound)
				require.True(t, ok)
				require.Equal(t, api.ProjectKind, enf.Type)
				require.Equal(t, testProjectID, enf.ID)
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
			assertions: func(project api.Project, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error finding/decoding project")
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
						api.Project{
							ObjectMeta: meta.ObjectMeta{
								ID: testProjectID,
							},
						},
					)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(project api.Project, err error) {
				require.NoError(t, err)
				require.Equal(t, testProjectID, project.ID)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &projectsStore{
				collection: testCase.collection,
			}
			user, err := store.Get(context.Background(), testProjectID)
			testCase.assertions(user, err)
		})
	}
}

func TestProjectsStoreUpdate(t *testing.T) {
	testProject := api.Project{
		ObjectMeta: meta.ObjectMeta{
			ID: "blue-book",
		},
	}

	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(err error)
	}{

		{
			name: "project not found",
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
				enf, ok := err.(*meta.ErrNotFound)
				require.True(t, ok)
				require.Equal(t, api.ProjectKind, enf.Type)
				require.Equal(t, testProject.ID, enf.ID)
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
				require.Contains(t, err.Error(), "error updating project")
			},
		},

		{
			name: "project found",
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
			store := &projectsStore{
				collection: testCase.collection,
			}
			err := store.Update(context.Background(), testProject)
			testCase.assertions(err)
		})
	}
}

func TestProjectsStoreDelete(t *testing.T) {
	const testProjectID = "blue-book"

	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(err error)
	}{

		{
			name: "project not found",
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
				enf, ok := err.(*meta.ErrNotFound)
				require.True(t, ok)
				require.Equal(t, api.ProjectKind, enf.Type)
				require.Equal(t, testProjectID, enf.ID)
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
				require.Contains(t, err.Error(), "error deleting project")
			},
		},

		{
			name: "project found",
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
			store := &projectsStore{
				collection: testCase.collection,
			}
			err := store.Delete(context.Background(), testProjectID)
			testCase.assertions(err)
		})
	}
}
