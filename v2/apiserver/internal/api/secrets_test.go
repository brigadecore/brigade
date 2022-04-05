package api

import (
	"context"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	metaTesting "github.com/brigadecore/brigade/v2/apiserver/internal/meta/testing" // nolint: lll
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestSecretMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, &Secret{}, "Secret")
}

func TestNewSecretsService(t *testing.T) {
	projectsStore := &mockProjectsStore{}
	secretsStore := &mockSecretsStore{}
	svc, ok := NewSecretsService(
		alwaysAuthorize,
		alwaysProjectAuthorize,
		projectsStore,
		secretsStore,
	).(*secretsService)
	require.True(t, ok)
	require.NotNil(t, svc.authorize)
	require.NotNil(t, svc.projectAuthorize)
	require.Same(t, projectsStore, svc.projectsStore)
	require.Same(t, secretsStore, svc.secretsStore)
}

func TestSecretsServiceList(t *testing.T) {
	const testProjectID = "italian"
	testCases := []struct {
		name       string
		service    SecretsService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &secretsService{
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting project from store",
			service: &secretsService{
				authorize: alwaysAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving project")
			},
		},
		{
			name: "error getting secrets from store",
			service: &secretsService{
				authorize: alwaysAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				secretsStore: &mockSecretsStore{
					ListFn: func(
						context.Context,
						Project,
						meta.ListOptions,
					) (meta.List[Secret], error) {
						return meta.List[Secret]{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error getting secrets for project")
			},
		},
		{
			name: "success",
			service: &secretsService{
				authorize: alwaysAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				secretsStore: &mockSecretsStore{
					ListFn: func(
						context.Context,
						Project,
						meta.ListOptions,
					) (meta.List[Secret], error) {
						return meta.List[Secret]{}, nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := testCase.service.List(
				context.Background(),
				testProjectID,
				meta.ListOptions{},
			)
			testCase.assertions(err)
		})
	}
}

func TestSecretsServiceSet(t *testing.T) {
	const testProjectID = "italian"
	testSecret := Secret{
		Key:   "soylentgreen",
		Value: "ispeople",
	}
	testCases := []struct {
		name       string
		service    SecretsService
		assertions func(error)
	}{
		{
			name: "user does not have read permissions",
			service: &secretsService{
				authorize: neverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting project from store",
			service: &secretsService{
				authorize: alwaysAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving project")
			},
		},
		{
			name: "user is not a project admin",
			service: &secretsService{
				authorize:        alwaysAuthorize,
				projectAuthorize: neverProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error setting secret in store",
			service: &secretsService{
				authorize:        alwaysAuthorize,
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				secretsStore: &mockSecretsStore{
					SetFn: func(context.Context, Project, Secret) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error setting secret")
			},
		},
		{
			name: "success",
			service: &secretsService{
				authorize:        alwaysAuthorize,
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				secretsStore: &mockSecretsStore{
					SetFn: func(context.Context, Project, Secret) error {
						return nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.assertions(
				testCase.service.Set(context.Background(), testProjectID, testSecret),
			)
		})
	}
}

func TestSecretsServiceUnSet(t *testing.T) {
	const testProjectID = "italian"
	const testKey = "soylentgreen"
	testCases := []struct {
		name       string
		service    SecretsService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &secretsService{
				projectAuthorize: neverProjectAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting project from store",
			service: &secretsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving project")
			},
		},
		{
			name: "error unsetting secret in store",
			service: &secretsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				secretsStore: &mockSecretsStore{
					UnsetFn: func(context.Context, Project, string) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error unsetting secret")
			},
		},
		{
			name: "success",
			service: &secretsService{
				projectAuthorize: alwaysProjectAuthorize,
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				secretsStore: &mockSecretsStore{
					UnsetFn: func(context.Context, Project, string) error {
						return nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.assertions(
				testCase.service.Unset(context.Background(), testProjectID, testKey),
			)
		})
	}
}

type mockSecretsStore struct {
	ListFn func(
		context.Context,
		Project,
		meta.ListOptions,
	) (meta.List[Secret], error)
	SetFn   func(context.Context, Project, Secret) error
	UnsetFn func(context.Context, Project, string) error
}

func (m *mockSecretsStore) List(
	ctx context.Context,
	project Project,
	opts meta.ListOptions,
) (meta.List[Secret], error) {
	return m.ListFn(ctx, project, opts)
}

func (m *mockSecretsStore) Set(
	ctx context.Context,
	project Project,
	secret Secret,
) error {
	return m.SetFn(ctx, project, secret)
}

func (m *mockSecretsStore) Unset(
	ctx context.Context,
	project Project,
	key string,
) error {
	return m.UnsetFn(ctx, project, key)
}
