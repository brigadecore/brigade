package core

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3/core"
	"github.com/brigadecore/brigade/sdk/v3/meta"
)

type MockProjectsClient struct {
	CreateFn func(
		context.Context,
		core.Project,
		*core.ProjectCreateOptions,
	) (core.Project, error)
	CreateFromBytesFn func(
		context.Context,
		[]byte,
		*core.ProjectCreateOptions,
	) (core.Project, error)
	ListFn func(
		context.Context,
		*core.ProjectsSelector,
		*meta.ListOptions,
	) (core.ProjectList, error)
	GetFn func(
		context.Context,
		string,
		*core.ProjectGetOptions,
	) (core.Project, error)
	UpdateFn func(
		context.Context,
		core.Project,
		*core.ProjectUpdateOptions,
	) (core.Project, error)
	UpdateFromBytesFn func(
		context.Context,
		string,
		[]byte,
		*core.ProjectUpdateOptions,
	) (core.Project, error)
	DeleteFn      func(context.Context, string, *core.ProjectDeleteOptions) error
	AuthzClient   core.AuthzClient
	SecretsClient core.SecretsClient
}

func (m *MockProjectsClient) Create(
	ctx context.Context,
	project core.Project,
	opts *core.ProjectCreateOptions,
) (core.Project, error) {
	return m.CreateFn(ctx, project, opts)
}

func (m *MockProjectsClient) CreateFromBytes(
	ctx context.Context,
	bytes []byte,
	opts *core.ProjectCreateOptions,
) (core.Project, error) {
	return m.CreateFromBytesFn(ctx, bytes, opts)
}

func (m *MockProjectsClient) List(
	ctx context.Context,
	selector *core.ProjectsSelector,
	opts *meta.ListOptions,
) (core.ProjectList, error) {
	return m.ListFn(ctx, selector, opts)
}

func (m *MockProjectsClient) Get(
	ctx context.Context,
	id string,
	opts *core.ProjectGetOptions,
) (core.Project, error) {
	return m.GetFn(ctx, id, opts)
}

func (m *MockProjectsClient) Update(
	ctx context.Context,
	project core.Project,
	opts *core.ProjectUpdateOptions,
) (core.Project, error) {
	return m.UpdateFn(ctx, project, opts)
}

func (m *MockProjectsClient) UpdateFromBytes(
	ctx context.Context,
	id string,
	bytes []byte,
	opts *core.ProjectUpdateOptions,
) (core.Project, error) {
	return m.UpdateFromBytesFn(ctx, id, bytes, opts)
}

func (m *MockProjectsClient) Delete(
	ctx context.Context,
	id string,
	opts *core.ProjectDeleteOptions,
) error {
	return m.DeleteFn(ctx, id, opts)
}

func (m *MockProjectsClient) Authz() core.AuthzClient {
	return m.AuthzClient
}

func (m *MockProjectsClient) Secrets() core.SecretsClient {
	return m.SecretsClient
}
