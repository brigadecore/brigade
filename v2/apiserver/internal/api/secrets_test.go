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

func TestSecretListMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, &SecretList{}, "SecretList")
}

func TestSecretListLen(t *testing.T) {
	secretList := SecretList{
		Items: []Secret{
			{
				Key:   "foo",
				Value: "bar",
			},
			{
				Key:   "bat",
				Value: "baz",
			},
		},
	}
	require.Equal(t, len(secretList.Items), secretList.Len())
}

func TestSecretListSwap(t *testing.T) {
	testSecret0 := Secret{
		Key:   "foo",
		Value: "bar",
	}
	testSecret1 := Secret{
		Key:   "bat",
		Value: "baz",
	}
	secretList := SecretList{
		Items: []Secret{testSecret0, testSecret1},
	}
	secretList.Swap(0, 1)
	require.Equal(
		t,
		[]Secret{testSecret1, testSecret0},
		secretList.Items,
	)
}

func TestSecretListLess(t *testing.T) {
	secretList := SecretList{
		Items: []Secret{
			{
				Key:   "foo",
				Value: "bar",
			},
			{
				Key:   "bat",
				Value: "baz",
			},
		},
	}
	require.False(t, secretList.Less(0, 0))
	require.False(t, secretList.Less(0, 1))
	require.True(t, secretList.Less(1, 0))
	require.False(t, secretList.Less(1, 1))
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
					) (SecretList, error) {
						return SecretList{}, errors.New("something went wrong")
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
					) (SecretList, error) {
						return SecretList{}, nil
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
	ListFn  func(context.Context, Project, meta.ListOptions) (SecretList, error)
	SetFn   func(context.Context, Project, Secret) error
	UnsetFn func(context.Context, Project, string) error
}

func (m *mockSecretsStore) List(
	ctx context.Context,
	project Project,
	opts meta.ListOptions,
) (SecretList, error) {
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
