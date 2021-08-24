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

// EventKind represents the canonical Event kind string
const EventKind = "Event"

// Event represents an occurrence in some upstream system. Once accepted into
// the system, Brigade amends each Event with a plan for handling it in the form
// of a Worker. An Event's status is, implicitly, the status of its Worker.
type Event struct {
	// ObjectMeta contains Event metadata.
	meta.ObjectMeta `json:"metadata"`
	// ProjectID specifies the Project this Event is for. Often, this field will
	// be left blank, in which case the Event is matched against subscribed
	// Projects on the basis of the Source, Type, Qualifiers, and Labels fields,
	// then used as a template to create a discrete Event for each subscribed
	// Project.
	ProjectID string `json:"projectID,omitempty"`
	// Source specifies the source of the Event, e.g. what gateway created it.
	// Gateways should populate this field with a unique string that clearly
	// identifies themself as the source of the event. The ServiceAccount used by
	// each gateway can be authorized (by a admin) to only create events having a
	// specified value in the Source field, thereby eliminating the possibility of
	// gateways maliciously creating events that spoof events from another
	// gateway.
	Source string `json:"source,omitempty"`
	// SourceState encapsulates opaque, source-specific (e.g. gateway-specific)
	// state.
	SourceState *SourceState `json:"sourceState,omitempty"`
	// Type specifies the exact event that has occurred in the upstream system.
	// Values are opaque and source-specific.
	Type string `json:"type,omitempty"`
	// Qualifiers provide critical disambiguation of an Event's type. A Project is
	// considered subscribed to an Event IF AND ONLY IF (in addition to matching
	// the Event's Source and Type) it matches ALL of the Event's qualifiers
	// EXACTLY. To demonstrate the usefulness of this, consider any event from a
	// hypothetical GitHub gateway. If, by design, that gateway does not intend
	// for any Project to subscribe to ALL Events (i.e. regardless of which
	// repository they originated from), then that gateway can QUALIFY Events it
	// emits into Brigade's event bus with repo=<repository name>. Projects
	// wishing to subscribe to Events from the GitHub gateway MUST include that
	// Qualifier in their EventSubscription. Note that the Qualifiers field's
	// "MUST match" subscription semantics differ from the Labels field's "MAY
	// match" subscription semantics.
	Qualifiers map[string]string `json:"qualifiers,omitempty"`
	// Labels convey supplementary Event details that Projects may OPTIONALLY use
	// to narrow EventSubscription criteria. A Project is considered subscribed to
	// an Event if (in addition to matching the Event's Source, Type, and
	// Qualifiers) the Event has ALL labels expressed in the Project's
	// EventSubscription. If the Event has ADDITIONAL labels, not mentioned by the
	// EventSubscription, these do not preclude a match. To demonstrate the
	// usefulness of this, consider any event from a hypothetical Slack gateway.
	// If, by design, that gateway intends for Projects to select between
	// subscribing to ALL Events or ONLY events originating from a specific
	// channel, then that gateway can LABEL Events it emits into Brigade's event
	// bus with channel=<channel name>. Projects wishing to subscribe to ALL
	// Events from the Slack gateway MAY omit that Label from their
	// EventSubscription, while Projects wishing to subscribe to only Events
	// originating from a specific channel MAY include that Label in their
	// EventSubscription. Note that the Labels field's "MAY match" subscription
	// semantics differ from the Qualifiers field's "MUST match" subscription
	// semantics.
	Labels map[string]string `json:"labels,omitempty"`
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
				Kind:       EventKind,
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

// SourceState encapsulates opaque, source-specific (e.g. gateway-specific)
// state.
type SourceState struct {
	// State is a map of arbitrary and opaque key/value pairs that the source of
	// an Event (e.g. the gateway that created it) can use to store
	// source-specific state.
	State map[string]string `json:"state,omitempty"`
}

// MarshalJSON amends SourceState instances with type metadata so that clients
// do not need to be concerned with the tedium of doing so.
func (s SourceState) MarshalJSON() ([]byte, error) {
	type Alias SourceState
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "SourceState",
			},
			Alias: (Alias)(s),
		},
	)
}

// EventsSelector represents useful filter criteria when selecting multiple
// Events for API group operations like list, cancel, or delete.
type EventsSelector struct {
	// ProjectID specifies that only Events belonging to the indicated Project
	// should be selected.
	ProjectID string
	// Source specifies that only Events from the indicated source should be
	// selected.
	Source string
	// SourceState specifies that only Events having all of the indicated source
	// state key/value pairs should be selected.
	SourceState map[string]string
	// Type specifies that only Events having the indicated type should be
	// selected.
	Type string
	// WorkerPhases specifies that only Events with their Workers in any of the
	// indicated phases should be selected.
	WorkerPhases []WorkerPhase
	// Qualifiers specifies that only Events qualified with these key/value pairs
	// should be selected.
	Qualifiers map[string]string
	// Labels specifies that only Events labeled with these key/value pairs should
	// be selected.
	Labels map[string]string
}

// GitDetails represents git-specific Event details. These may override
// Project-level GitConfig.
type GitDetails struct {
	// CloneURL specifies the location from where a source code repository may
	// be cloned.
	CloneURL string `json:"cloneURL,omitempty"`
	// Commit specifies a revision (by SHA) to be checked out. If non-empty, this
	// field takes precedence over any value in the Ref field.
	Commit string `json:"commit,omitempty"`
	// Ref is a symbolic reference to a revision to be checked out. If non-empty,
	// the value of the Commit field supercedes any value in this field. Example
	// uses of this field include referencing a branch (refs/heads/<branch name>)
	// or a tag (refs/tags/<tag name>). If left blank, this field is interpreted
	// as a reference to the repository's default branch.
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
	// Clones a pre-existing Event, removing the original's metadata and Worker
	// config in the process.  A new Event is created using the rest of the
	// details preserved from the original.
	Clone(context.Context, string) (Event, error)
	// UpdateSourceState updates source-specific (e.g. gateway-specific) Event
	// state.
	UpdateSourceState(context.Context, string, SourceState) error
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
	// Retry copies an Event, including Worker configuration and Jobs, and
	// creates a new Event from this information. Where possible, job results
	// are inherited and the job not re-scheduled, for example when a job has
	// succeeded and does not make use of a shared workspace.
	Retry(context.Context, string) (Event, error)

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
	queryParams := eventsSelectorToQueryParams(selector)
	events := EventList{}
	return events, e.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/events",
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
			SuccessCode: http.StatusOK,
			RespObj:     &event,
		},
	)
}

func (e *eventsClient) Clone(
	ctx context.Context,
	id string,
) (Event, error) {
	event := Event{}
	return event, e.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPost,
			Path:        fmt.Sprintf("v2/events/%s/clones", id),
			SuccessCode: http.StatusCreated,
			RespObj:     &event,
		},
	)
}

func (e *eventsClient) UpdateSourceState(
	ctx context.Context,
	id string,
	sourceState SourceState,
) error {
	return e.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/events/%s/source-state", id),
			ReqBodyObj:  sourceState,
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *eventsClient) Cancel(ctx context.Context, id string) error {
	return e.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/events/%s/cancellation", id),
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *eventsClient) CancelMany(
	ctx context.Context,
	selector EventsSelector,
) (CancelManyEventsResult, error) {
	queryParams := eventsSelectorToQueryParams(&selector)
	result := CancelManyEventsResult{}
	return result, e.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/events/cancellations",
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
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *eventsClient) DeleteMany(
	ctx context.Context,
	selector EventsSelector,
) (DeleteManyEventsResult, error) {
	queryParams := eventsSelectorToQueryParams(&selector)
	result := DeleteManyEventsResult{}
	return result, e.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/events",
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
			RespObj:     &result,
		},
	)
}

func (e *eventsClient) Retry(
	ctx context.Context,
	id string,
) (Event, error) {
	event := Event{}
	return event, e.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPost,
			Path:        fmt.Sprintf("v2/events/%s/retries", id),
			SuccessCode: http.StatusCreated,
			RespObj:     &event,
		},
	)
}

func (e *eventsClient) Workers() WorkersClient {
	return e.workersClient
}

func (e *eventsClient) Logs() LogsClient {
	return e.logsClient
}

func eventsSelectorToQueryParams(selector *EventsSelector) map[string]string {
	if selector == nil {
		return nil
	}
	queryParams := map[string]string{}
	if selector.ProjectID != "" {
		queryParams["projectID"] = selector.ProjectID
	}
	if selector.Source != "" {
		queryParams["source"] = selector.Source
	}
	if len(selector.Qualifiers) > 0 {
		qualifiersStrs := make([]string, len(selector.Qualifiers))
		i := 0
		for k, v := range selector.Qualifiers {
			qualifiersStrs[i] = fmt.Sprintf("%s=%s", k, v)
			i++
		}
		queryParams["qualifiers"] = strings.Join(qualifiersStrs, ",")
	}
	if len(selector.Labels) > 0 {
		labelsStrs := make([]string, len(selector.Labels))
		i := 0
		for k, v := range selector.Labels {
			labelsStrs[i] = fmt.Sprintf("%s=%s", k, v)
			i++
		}
		queryParams["labels"] = strings.Join(labelsStrs, ",")
	}
	if len(selector.SourceState) > 0 {
		sourceStateStrs := make([]string, len(selector.SourceState))
		i := 0
		for k, v := range selector.SourceState {
			sourceStateStrs[i] = fmt.Sprintf("%s=%s", k, v)
			i++
		}
		queryParams["sourceState"] = strings.Join(sourceStateStrs, ",")
	}
	if selector.Type != "" {
		queryParams["type"] = selector.Type
	}
	if len(selector.WorkerPhases) > 0 {
		workerPhaseStrs := make([]string, len(selector.WorkerPhases))
		for i, workerPhase := range selector.WorkerPhases {
			workerPhaseStrs[i] = string(workerPhase)
		}
		queryParams["workerPhases"] = strings.Join(workerPhaseStrs, ",")
	}
	return queryParams
}
