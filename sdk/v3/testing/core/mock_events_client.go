package core

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3/core"
	"github.com/brigadecore/brigade/sdk/v3/meta"
)

type MockEventsClient struct {
	CreateFn func(
		context.Context,
		core.Event,
		*core.EventCreateOptions,
	) (core.EventList, error)
	ListFn func(
		context.Context,
		*core.EventsSelector,
		*meta.ListOptions,
	) (core.EventList, error)
	GetFn func(
		context.Context,
		string,
		*core.EventGetOptions,
	) (core.Event, error)
	CloneFn func(
		context.Context,
		string,
		*core.EventCloneOptions,
	) (core.Event, error)
	UpdateSourceStateFn func(
		context.Context,
		string,
		core.SourceState,
		*core.EventSourceStateUpdateOptions,
	) error
	UpdateSummaryFn func(
		context.Context,
		string,
		core.EventSummary,
		*core.EventSummaryUpdateOptions,
	) error
	CancelFn     func(context.Context, string, *core.EventCancelOptions) error
	CancelManyFn func(
		context.Context,
		core.EventsSelector,
		*core.EventCancelManyOptions,
	) (core.CancelManyEventsResult, error)
	DeleteFn     func(context.Context, string, *core.EventDeleteOptions) error
	DeleteManyFn func(
		context.Context,
		core.EventsSelector,
		*core.EventDeleteManyOptions,
	) (core.DeleteManyEventsResult, error)
	RetryFn func(
		context.Context,
		string,
		*core.EventRetryOptions,
	) (core.Event, error)
	WorkersClient core.WorkersClient
	LogsClient    core.LogsClient
}

func (m *MockEventsClient) Create(
	ctx context.Context,
	event core.Event,
	opts *core.EventCreateOptions,
) (core.EventList, error) {
	return m.CreateFn(ctx, event, opts)
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
	opts *core.EventGetOptions,
) (core.Event, error) {
	return m.GetFn(ctx, id, opts)
}

func (m *MockEventsClient) Clone(
	ctx context.Context,
	id string,
	opts *core.EventCloneOptions,
) (core.Event, error) {
	return m.CloneFn(ctx, id, opts)
}

func (m *MockEventsClient) UpdateSourceState(
	ctx context.Context,
	id string,
	state core.SourceState,
	opts *core.EventSourceStateUpdateOptions,
) error {
	return m.UpdateSourceStateFn(ctx, id, state, opts)
}

func (m *MockEventsClient) UpdateSummary(
	ctx context.Context,
	id string,
	summary core.EventSummary,
	opts *core.EventSummaryUpdateOptions,
) error {
	return m.UpdateSummaryFn(ctx, id, summary, opts)
}

func (m *MockEventsClient) Cancel(
	ctx context.Context,
	id string,
	opts *core.EventCancelOptions,
) error {
	return m.CancelFn(ctx, id, opts)
}

func (m *MockEventsClient) CancelMany(
	ctx context.Context,
	selector core.EventsSelector,
	opts *core.EventCancelManyOptions,
) (core.CancelManyEventsResult, error) {
	return m.CancelManyFn(ctx, selector, opts)
}

func (m *MockEventsClient) Delete(
	ctx context.Context,
	id string,
	opts *core.EventDeleteOptions,
) error {
	return m.DeleteFn(ctx, id, opts)
}

func (m *MockEventsClient) DeleteMany(
	ctx context.Context,
	selector core.EventsSelector,
	opts *core.EventDeleteManyOptions,
) (core.DeleteManyEventsResult, error) {
	return m.DeleteManyFn(ctx, selector, opts)
}

func (m *MockEventsClient) Retry(
	ctx context.Context,
	id string,
	opts *core.EventRetryOptions,
) (core.Event, error) {
	return m.RetryFn(ctx, id, opts)
}

func (m *MockEventsClient) Workers() core.WorkersClient {
	return m.WorkersClient
}

func (m *MockEventsClient) Logs() core.LogsClient {
	return m.LogsClient
}
