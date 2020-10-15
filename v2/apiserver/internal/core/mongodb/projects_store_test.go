package mongodb

import (
	"context"
	"errors"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestProjectsStoreCreate(t *testing.T) {
	testProject := core.Project{
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
			collection: &mockCollection{
				InsertOneFn: func(
					ctx context.Context,
					document interface{},
					opts ...*options.InsertOneOptions,
				) (*mongo.InsertOneResult, error) {
					return nil, mockWriteException
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrConflict{}, err)
				require.Equal(t, "Project", err.(*meta.ErrConflict).Type)
				require.Equal(t, testProject.ID, err.(*meta.ErrConflict).ID)
				require.Contains(t, err.(*meta.ErrConflict).Reason, "already exists")
			},
		},

		{
			name: "unanticipated error",
			collection: &mockCollection{
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
			collection: &mockCollection{
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
	testProject := core.Project{
		ObjectMeta: meta.ObjectMeta{
			ID: "blue-book",
		},
	}

	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(projects core.ProjectList, err error)
	}{

		{
			name: "error finding projects",
			collection: &mockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(projects core.ProjectList, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error finding projects")
			},
		},

		{
			name: "projects found; no more pages of results exist",
			collection: &mockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cursor, err := mockCursor(testProject)
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
			assertions: func(projects core.ProjectList, err error) {
				require.NoError(t, err)
				require.Empty(t, projects.Continue)
				require.Zero(t, projects.RemainingItemCount)
			},
		},

		{
			name: "projects found; more pages of results exist",
			collection: &mockCollection{
				FindFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOptions,
				) (*mongo.Cursor, error) {
					cursor, err := mockCursor(testProject)
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
			assertions: func(projects core.ProjectList, err error) {
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

func TestProjectsStoreGet(t *testing.T) {
	const testProjectID = "blue-book"
	testCases := []struct {
		name       string
		collection mongodb.Collection
		assertions func(project core.Project, err error)
	}{

		{
			name: "project not found",
			collection: &mockCollection{
				FindOneFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mockSingleResult(mongo.ErrNoDocuments)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(project core.Project, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
				require.Equal(t, "Project", err.(*meta.ErrNotFound).Type)
				require.Equal(t, testProjectID, err.(*meta.ErrNotFound).ID)
			},
		},

		{
			name: "unanticipated error",
			collection: &mockCollection{
				FindOneFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mockSingleResult(errors.New("something went wrong"))
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(project core.Project, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error finding/decoding project")
			},
		},

		{
			name: "user found",
			collection: &mockCollection{
				FindOneFn: func(
					ctx context.Context,
					filter interface{},
					opts ...*options.FindOneOptions,
				) *mongo.SingleResult {
					res, err := mockSingleResult(
						core.Project{
							ObjectMeta: meta.ObjectMeta{
								ID: testProjectID,
							},
						},
					)
					require.NoError(t, err)
					return res
				},
			},
			assertions: func(project core.Project, err error) {
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
	testProject := core.Project{
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
			collection: &mockCollection{
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
				require.Equal(t, "Project", err.(*meta.ErrNotFound).Type)
				require.Equal(t, testProject.ID, err.(*meta.ErrNotFound).ID)
			},
		},

		{
			name: "unanticipated error",
			collection: &mockCollection{
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
			collection: &mockCollection{
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
			collection: &mockCollection{
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
				require.Equal(t, "Project", err.(*meta.ErrNotFound).Type)
				require.Equal(t, testProjectID, err.(*meta.ErrNotFound).ID)
			},
		},

		{
			name: "unanticipated error",
			collection: &mockCollection{
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
			collection: &mockCollection{
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
