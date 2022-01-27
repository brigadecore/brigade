package testing

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/brigadecore/brigade/sdk/v3/meta"
)

type MockProjectsClient struct {
	CreateFn func(
		context.Context,
		sdk.Project,
		*sdk.ProjectCreateOptions,
	) (sdk.Project, error)
	CreateFromBytesFn func(
		context.Context,
		[]byte,
		*sdk.ProjectCreateOptions,
	) (sdk.Project, error)
	ListFn func(
		context.Context,
		*sdk.ProjectsSelector,
		*meta.ListOptions,
	) (sdk.ProjectList, error)
	GetFn func(
		context.Context,
		string,
		*sdk.ProjectGetOptions,
	) (sdk.Project, error)
	UpdateFn func(
		context.Context,
		sdk.Project,
		*sdk.ProjectUpdateOptions,
	) (sdk.Project, error)
	UpdateFromBytesFn func(
		context.Context,
		string,
		[]byte,
		*sdk.ProjectUpdateOptions,
	) (sdk.Project, error)
	DeleteFn      func(context.Context, string, *sdk.ProjectDeleteOptions) error
	AuthzClient   sdk.ProjectAuthzClient
	SecretsClient sdk.SecretsClient
}

func (m *MockProjectsClient) Create(
	ctx context.Context,
	project sdk.Project,
	opts *sdk.ProjectCreateOptions,
) (sdk.Project, error) {
	return m.CreateFn(ctx, project, opts)
}

func (m *MockProjectsClient) CreateFromBytes(
	ctx context.Context,
	bytes []byte,
	opts *sdk.ProjectCreateOptions,
) (sdk.Project, error) {
	return m.CreateFromBytesFn(ctx, bytes, opts)
}

func (m *MockProjectsClient) List(
	ctx context.Context,
	selector *sdk.ProjectsSelector,
	opts *meta.ListOptions,
) (sdk.ProjectList, error) {
	return m.ListFn(ctx, selector, opts)
}

func (m *MockProjectsClient) Get(
	ctx context.Context,
	id string,
	opts *sdk.ProjectGetOptions,
) (sdk.Project, error) {
	return m.GetFn(ctx, id, opts)
}

func (m *MockProjectsClient) Update(
	ctx context.Context,
	project sdk.Project,
	opts *sdk.ProjectUpdateOptions,
) (sdk.Project, error) {
	return m.UpdateFn(ctx, project, opts)
}

func (m *MockProjectsClient) UpdateFromBytes(
	ctx context.Context,
	id string,
	bytes []byte,
	opts *sdk.ProjectUpdateOptions,
) (sdk.Project, error) {
	return m.UpdateFromBytesFn(ctx, id, bytes, opts)
}

func (m *MockProjectsClient) Delete(
	ctx context.Context,
	id string,
	opts *sdk.ProjectDeleteOptions,
) error {
	return m.DeleteFn(ctx, id, opts)
}

func (m *MockProjectsClient) Authz() sdk.ProjectAuthzClient {
	return m.AuthzClient
}

func (m *MockProjectsClient) Secrets() sdk.SecretsClient {
	return m.SecretsClient
}
