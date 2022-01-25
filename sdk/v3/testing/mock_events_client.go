package testing

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/brigadecore/brigade/sdk/v3/meta"
)

type MockEventsClient struct {
	CreateFn func(
		context.Context,
		sdk.Event,
		*sdk.EventCreateOptions,
	) (sdk.EventList, error)
	ListFn func(
		context.Context,
		*sdk.EventsSelector,
		*meta.ListOptions,
	) (sdk.EventList, error)
	GetFn func(
		context.Context,
		string,
		*sdk.EventGetOptions,
	) (sdk.Event, error)
	CloneFn func(
		context.Context,
		string,
		*sdk.EventCloneOptions,
	) (sdk.Event, error)
	UpdateSourceStateFn func(
		context.Context,
		string,
		sdk.SourceState,
		*sdk.EventSourceStateUpdateOptions,
	) error
	UpdateSummaryFn func(
		context.Context,
		string,
		sdk.EventSummary,
		*sdk.EventSummaryUpdateOptions,
	) error
	CancelFn     func(context.Context, string, *sdk.EventCancelOptions) error
	CancelManyFn func(
		context.Context,
		sdk.EventsSelector,
		*sdk.EventCancelManyOptions,
	) (sdk.CancelManyEventsResult, error)
	DeleteFn     func(context.Context, string, *sdk.EventDeleteOptions) error
	DeleteManyFn func(
		context.Context,
		sdk.EventsSelector,
		*sdk.EventDeleteManyOptions,
	) (sdk.DeleteManyEventsResult, error)
	RetryFn func(
		context.Context,
		string,
		*sdk.EventRetryOptions,
	) (sdk.Event, error)
	WorkersClient sdk.WorkersClient
	LogsClient    sdk.LogsClient
}

func (m *MockEventsClient) Create(
	ctx context.Context,
	event sdk.Event,
	opts *sdk.EventCreateOptions,
) (sdk.EventList, error) {
	return m.CreateFn(ctx, event, opts)
}

func (m *MockEventsClient) List(
	ctx context.Context,
	selector *sdk.EventsSelector,
	opts *meta.ListOptions,
) (sdk.EventList, error) {
	return m.ListFn(ctx, selector, opts)
}

func (m *MockEventsClient) Get(
	ctx context.Context,
	id string,
	opts *sdk.EventGetOptions,
) (sdk.Event, error) {
	return m.GetFn(ctx, id, opts)
}

func (m *MockEventsClient) Clone(
	ctx context.Context,
	id string,
	opts *sdk.EventCloneOptions,
) (sdk.Event, error) {
	return m.CloneFn(ctx, id, opts)
}

func (m *MockEventsClient) UpdateSourceState(
	ctx context.Context,
	id string,
	state sdk.SourceState,
	opts *sdk.EventSourceStateUpdateOptions,
) error {
	return m.UpdateSourceStateFn(ctx, id, state, opts)
}

func (m *MockEventsClient) UpdateSummary(
	ctx context.Context,
	id string,
	summary sdk.EventSummary,
	opts *sdk.EventSummaryUpdateOptions,
) error {
	return m.UpdateSummaryFn(ctx, id, summary, opts)
}

func (m *MockEventsClient) Cancel(
	ctx context.Context,
	id string,
	opts *sdk.EventCancelOptions,
) error {
	return m.CancelFn(ctx, id, opts)
}

func (m *MockEventsClient) CancelMany(
	ctx context.Context,
	selector sdk.EventsSelector,
	opts *sdk.EventCancelManyOptions,
) (sdk.CancelManyEventsResult, error) {
	return m.CancelManyFn(ctx, selector, opts)
}

func (m *MockEventsClient) Delete(
	ctx context.Context,
	id string,
	opts *sdk.EventDeleteOptions,
) error {
	return m.DeleteFn(ctx, id, opts)
}

func (m *MockEventsClient) DeleteMany(
	ctx context.Context,
	selector sdk.EventsSelector,
	opts *sdk.EventDeleteManyOptions,
) (sdk.DeleteManyEventsResult, error) {
	return m.DeleteManyFn(ctx, selector, opts)
}

func (m *MockEventsClient) Retry(
	ctx context.Context,
	id string,
	opts *sdk.EventRetryOptions,
) (sdk.Event, error) {
	return m.RetryFn(ctx, id, opts)
}

func (m *MockEventsClient) Workers() sdk.WorkersClient {
	return m.WorkersClient
}

func (m *MockEventsClient) Logs() sdk.LogsClient {
	return m.LogsClient
}
