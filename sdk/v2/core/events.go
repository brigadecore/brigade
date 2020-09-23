package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// Event represents an occurrence in some upstream system. Once accepted into
// the system, Brigade amends each Event with a plan for handling it in the form
// of a Worker. An Event's status is, implicitly, the status of its Worker.
type Event struct {
	// ObjectMeta contains Event metadata.
	meta.ObjectMeta `json:"metadata"`
	// ProjectID specifies the Project this Event is for. Often, this field will
	// be left blank, in which case the Event is matched against subscribed
	// Projects on the basis of the Source, Type, and Labels fields, then used as
	// a template to create a discrete Event for each subscribed Project.
	ProjectID string `json:"projectID,omitempty"`
	// Source specifies the source of the event, e.g. what gateway created it.
	// Gateways should populate this field with a unique string that clearly
	// identifies themself as the source of the event. The ServiceAccount used by
	// each gateway can be authorized (by a admin) to only create events having a
	// specified value in the Source field, thereby eliminating the possibility of
	// gateways maliciously creating events that spoof events from another
	// gateway.
	Source string `json:"source,omitempty"`
	// Type specifies the exact event that has occurred in the upstream system.
	// Values are opaque and source-specific.
	Type string `json:"type,omitempty"`
	// Labels convey additional event details for the purposes of matching Events
	// to subscribed projects. For instance, no subscribers to the "GitHub" Source
	// and the "push" Type are likely to want to hear about push events for ALL
	// repositories. If the "GitHub" gateway labels events with the name of the
	// repository from which the event originated (e.g. "repo=github.com/foo/bar")
	// then subscribers can utilize those same criteria to narrow their
	// subscription from all push events emitted by the GitHub gateway to just
	// those having originated from a specific repository.
	Labels Labels `json:"labels,omitempty"`
	// ShortTitle is an optional, succinct title for the Event, ideal for use in
	// lists or in scenarios where UI real estate is constrained.
	ShortTitle string `json:"shortTitle,omitempty"`
	// LongTitle is an optional, detailed title for the Event.
	LongTitle string `json:"longTitle,omitempty"`
	// Git contains git-specific Event details. These can be used to override
	// similar details defined at the Project level. This is useful for scenarios
	// wherein an Event may need to convey an alternative source, branch, etc.
	Git *GitDetails `json:"git,omitempty"`
	// Payload optionally contains Event details provided by the upstream system
	// that was the original source of the event. Payloads MUST NOT contain
	// sensitive information. Since Projects SUBSCRIBE to Events, the potential
	// exists for any Project to express an interest in any or all Events. This
	// being the case, sensitive details must never be present in payloads. The
	// common workaround work this constraint (which is also a sensible practice
	// to begin with) is that event payloads may contain REFERENCES to sensitive
	// details that are useful only to properly configured Workers.
	Payload string `json:"payload,omitempty"`
	// Worker contains details of the Worker assigned to handle the Event.
	Worker *Worker `json:"worker,omitempty"`
}

// MarshalJSON amends Event instances with type metadata so that clients do not
// need to be concerned with the tedium of doing so.
func (e Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "Event",
			},
			Alias: (Alias)(e),
		},
	)
}

// EventList is an ordered and pageable list of Events.
type EventList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Events.
	Items []Event `json:"items,omitempty"`
}

// MarshalJSON amends EventList instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
func (e EventList) MarshalJSON() ([]byte, error) {
	type Alias EventList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "EventList",
			},
			Alias: (Alias)(e),
		},
	)
}

// EventsSelector represents useful filter criteria when selecting multiple
// Events for API group operations like list, cancel, or delete.
type EventsSelector struct {
	// ProjectID specifies that Events belonging to the indicated Project should
	// be selected.
	ProjectID string
	// WorkerPhases specifies that Events with their Workers in any of the
	// indicated phases should be selected.
	WorkerPhases []WorkerPhase
}

// GitDetails represents git-specific Event details. These may override
// Project-level GitConfig.
type GitDetails struct {
	// CloneURL specifies the location from where a source code repository may
	// be cloned.
	CloneURL string `json:"cloneURL,omitempty"`
	// Commit specifies a commit (by sha) to be checked out.
	Commit string `json:"commit,omitempty"`
	// Ref specifies a tag or branch to be checked out. If left blank, this will
	// default to "master" at runtime.
	Ref string `json:"ref,omitempty"`
}

// CancelManyEventsResult represents a summary of a mass Event cancellation
// operation.
type CancelManyEventsResult struct {
	// Count represents the number of Events canceled.
	Count int64 `json:"count"`
}

// DeleteManyEventsResult represents a summary of a mass Event deletion
// operation.
type DeleteManyEventsResult struct {
	// Count represents the number of Events deleted.
	Count int64 `json:"count"`
}

// EventsClient is the specialized client for managing Events with the Brigade
// API.
type EventsClient interface {
	// Create creates one new Event if the Event provided references a Project by
	// ID. Otherwise, the Event provided is treated as a template and zero or more
	// discrete Events may be created-- one for each subscribed Project. An
	// EventList is returned containing all newly created Events.
	Create(context.Context, Event) (EventList, error)
	// List returns an EventList, with its Items (Events) ordered by age, newest
	// first. Criteria for which Events should be retrieved can be specified using
	// the EventsSelector parameter.
	List(context.Context, *EventsSelector, *meta.ListOptions) (EventList, error)
	// Get retrieves a single Event specified by its identifier.
	Get(context.Context, string) (Event, error)
	// Cancel cancels a single Event specified by its identifier.
	Cancel(context.Context, string) error
	// CancelMany cancels multiple Events specified by the EventListOptions
	// parameter.
	CancelMany(context.Context, EventsSelector) (CancelManyEventsResult, error)
	// Delete deletes a single Event specified by its identifier.
	Delete(context.Context, string) error
	// DeleteMany deletes multiple Events specified by the EventListOptions
	// parameter.
	DeleteMany(context.Context, EventsSelector) (DeleteManyEventsResult, error)

	// Workers returns a specialized client for Worker management.
	Workers() WorkersClient

	// Logs returns a specialized client for Log management.
	Logs() LogsClient
}

type eventsClient struct {
	*rm.BaseClient
	workersClient WorkersClient
	logsClient    LogsClient
}

// NewEventsClient returns a specialized client for managing Events.
func NewEventsClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) EventsClient {
	return &eventsClient{
		BaseClient:    rm.NewBaseClient(apiAddress, apiToken, opts),
		workersClient: NewWorkersClient(apiAddress, apiToken, opts),
		logsClient:    NewLogsClient(apiAddress, apiToken, opts),
	}
}

func (e *eventsClient) Create(
	ctx context.Context,
	event Event,
) (EventList, error) {
	events := EventList{}
	return events, e.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/events",
			AuthHeaders: e.BearerTokenAuthHeaders(),
			ReqBodyObj:  event,
			SuccessCode: http.StatusCreated,
			RespObj:     &events,
		},
	)
}

func (e *eventsClient) List(
	ctx context.Context,
	selector *EventsSelector,
	opts *meta.ListOptions,
) (EventList, error) {
	queryParams := map[string]string{}
	if selector.ProjectID != "" {
		queryParams["projectID"] = selector.ProjectID
	}
	if len(selector.WorkerPhases) > 0 {
		workerPhaseStrs := make([]string, len(selector.WorkerPhases))
		for i, workerPhase := range selector.WorkerPhases {
			workerPhaseStrs[i] = string(workerPhase)
		}
		queryParams["workerPhases"] = strings.Join(workerPhaseStrs, ",")
	}
	events := EventList{}
	return events, e.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/events",
			AuthHeaders: e.BearerTokenAuthHeaders(),
			QueryParams: e.AppendListQueryParams(queryParams, opts),
			SuccessCode: http.StatusOK,
			RespObj:     &events,
		},
	)
}

func (e *eventsClient) Get(
	ctx context.Context,
	id string,
) (Event, error) {
	event := Event{}
	return event, e.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/events/%s", id),
			AuthHeaders: e.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &event,
		},
	)
}

func (e *eventsClient) Cancel(ctx context.Context, id string) error {
	return e.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/events/%s/cancellation", id),
			AuthHeaders: e.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *eventsClient) CancelMany(
	ctx context.Context,
	opts EventsSelector,
) (CancelManyEventsResult, error) {
	queryParams := map[string]string{}
	if opts.ProjectID != "" {
		queryParams["projectID"] = opts.ProjectID
	}
	if len(opts.WorkerPhases) > 0 {
		workerPhaseStrs := make([]string, len(opts.WorkerPhases))
		for i, workerPhase := range opts.WorkerPhases {
			workerPhaseStrs[i] = string(workerPhase)
		}
		queryParams["workerPhases"] = strings.Join(workerPhaseStrs, ",")
	}
	result := CancelManyEventsResult{}
	return result, e.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/events/cancellations",
			AuthHeaders: e.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
			RespObj:     &result,
		},
	)
}

func (e *eventsClient) Delete(ctx context.Context, id string) error {
	return e.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/events/%s", id),
			AuthHeaders: e.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *eventsClient) DeleteMany(
	ctx context.Context,
	selector EventsSelector,
) (DeleteManyEventsResult, error) {
	queryParams := map[string]string{}
	if selector.ProjectID != "" {
		queryParams["projectID"] = selector.ProjectID
	}
	if len(selector.WorkerPhases) > 0 {
		workerPhaseStrs := make([]string, len(selector.WorkerPhases))
		for i, workerPhase := range selector.WorkerPhases {
			workerPhaseStrs[i] = string(workerPhase)
		}
		queryParams["workerPhases"] = strings.Join(workerPhaseStrs, ",")
	}
	result := DeleteManyEventsResult{}
	return result, e.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/events",
			AuthHeaders: e.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
			RespObj:     &result,
		},
	)
}

func (e *eventsClient) Workers() WorkersClient {
	return e.workersClient
}

func (e *eventsClient) Logs() LogsClient {
	return e.logsClient
}
