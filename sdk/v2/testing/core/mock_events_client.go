package core

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
)

type MockEventsClient struct {
	CreateFn func(context.Context, core.Event) (core.EventList, error)
	ListFn   func(
		context.Context,
		*core.EventsSelector,
		*meta.ListOptions,
	) (core.EventList, error)
	GetFn               func(context.Context, string) (core.Event, error)
	CloneFn             func(context.Context, string) (core.Event, error)
	UpdateSourceStateFn func(context.Context, string, core.SourceState) error
	CancelFn            func(context.Context, string) error
	CancelManyFn        func(
		context.Context,
		core.EventsSelector,
	) (core.CancelManyEventsResult, error)
	DeleteFn     func(context.Context, string) error
	DeleteManyFn func(
		context.Context,
		core.EventsSelector,
	) (core.DeleteManyEventsResult, error)
	RetryFn       func(context.Context, string) (core.Event, error)
	WorkersClient core.WorkersClient
	LogsClient    core.LogsClient
}

func (m *MockEventsClient) Create(
	ctx context.Context,
	event core.Event,
) (core.EventList, error) {
	return m.CreateFn(ctx, event)
}

func (m *MockEventsClient) List(
	ctx context.Context,
	selector *core.EventsSelector,
	opts *meta.ListOptions,
) (core.EventList, error) {
	return m.ListFn(ctx, selector, opts)
}

func (m *MockEventsClient) Get(
	ctx context.Context,
	id string,
) (core.Event, error) {
	return m.GetFn(ctx, id)
}

func (m *MockEventsClient) Clone(
	ctx context.Context,
	id string,
) (core.Event, error) {
	return m.CloneFn(ctx, id)
}

func (m *MockEventsClient) UpdateSourceState(
	ctx context.Context,
	id string,
	state core.SourceState,
) error {
	return m.UpdateSourceStateFn(ctx, id, state)
}

func (m *MockEventsClient) Cancel(ctx context.Context, id string) error {
	return m.CancelFn(ctx, id)
}

func (m *MockEventsClient) CancelMany(
	ctx context.Context,
	selector core.EventsSelector,
) (core.CancelManyEventsResult, error) {
	return m.CancelManyFn(ctx, selector)
}

func (m *MockEventsClient) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *MockEventsClient) DeleteMany(
	ctx context.Context,
	selector core.EventsSelector,
) (core.DeleteManyEventsResult, error) {
	return m.DeleteManyFn(ctx, selector)
}

func (m *MockEventsClient) Retry(
	ctx context.Context,
	id string,
) (core.Event, error) {
	return m.RetryFn(ctx, id)
}

func (m *MockEventsClient) Workers() core.WorkersClient {
	return m.WorkersClient
}

func (m *MockEventsClient) Logs() core.LogsClient {
	return m.LogsClient
}
