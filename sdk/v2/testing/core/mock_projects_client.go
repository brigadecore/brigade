package core

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
)

type MockProjectsClient struct {
	CreateFn          func(context.Context, core.Project) (core.Project, error)
	CreateFromBytesFn func(context.Context, []byte) (core.Project, error)
	ListFn            func(
		context.Context,
		*core.ProjectsSelector,
		*meta.ListOptions,
	) (core.ProjectList, error)
	GetFn             func(context.Context, string) (core.Project, error)
	UpdateFn          func(context.Context, core.Project) (core.Project, error)
	UpdateFromBytesFn func(context.Context, string, []byte) (core.Project, error)
	DeleteFn          func(context.Context, string) error
	AuthzClient       core.AuthzClient
	SecretsClient     core.SecretsClient
}

func (m *MockProjectsClient) Create(
	ctx context.Context,
	project core.Project,
) (core.Project, error) {
	return m.CreateFn(ctx, project)
}

func (m *MockProjectsClient) CreateFromBytes(
	ctx context.Context,
	bytes []byte,
) (core.Project, error) {
	return m.CreateFromBytesFn(ctx, bytes)
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
) (core.Project, error) {
	return m.GetFn(ctx, id)
}

func (m *MockProjectsClient) Update(
	ctx context.Context,
	project core.Project,
) (core.Project, error) {
	return m.UpdateFn(ctx, project)
}

func (m *MockProjectsClient) UpdateFromBytes(
	ctx context.Context,
	id string,
	bytes []byte,
) (core.Project, error) {
	return m.UpdateFromBytesFn(ctx, id, bytes)
}

func (m *MockProjectsClient) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *MockProjectsClient) Authz() core.AuthzClient {
	return m.AuthzClient
}

func (m *MockProjectsClient) Secrets() core.SecretsClient {
	return m.SecretsClient
}
