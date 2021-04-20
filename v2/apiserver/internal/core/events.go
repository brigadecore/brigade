package core

import (
	"context"
	"encoding/json"
	"log"
	"time"

	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/crypto"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/brigadecore/brigade/v2/apiserver/internal/system"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// EventKind represents the canonical Event kind string
const EventKind = "Event"

const defaultWorkspaceSize = "10Gi"

// Event represents an occurrence in some upstream system. Once accepted into
// the system, Brigade amends each Event with a plan for handling it in the form
// of a Worker. An Event's status is, implicitly, the status of its Worker.
type Event struct {
	// ObjectMeta contains Event metadata.
	meta.ObjectMeta `json:"metadata" bson:",inline"`
	// ProjectID specifies the Project this Event is for. Often, this field will
	// be left blank, in which case the Event is matched against subscribed
	// Projects on the basis of the Source, Type, Qualifiers, and Labels fields,
	// then used as a template to create a discrete Event for each subscribed
	// Project.
	ProjectID string `json:"projectID,omitempty" bson:"projectID,omitempty"`
	// Source specifies the source of the Event, e.g. what gateway created it.
	// Gateways should populate this field with a unique string that clearly
	// identifies themself as the source of the event. The ServiceAccount used by
	// each gateway can be authorized (by a admin) to only create events having a
	// specified value in the Source field, thereby eliminating the possibility of
	// gateways maliciously creating events that spoof events from another
	// gateway.
	Source string `json:"source,omitempty" bson:"source,omitempty"`
	// SourceState encapsulates opaque, source-specific (e.g. gateway-specific)
	// state.
	SourceState *SourceState `json:"sourceState,omitempty" bson:"sourceState,omitempty"` // nolint: lll
	// Type specifies the exact event that has occurred in the upstream system.
	// Values are opaque and source-specific.
	Type string `json:"type,omitempty" bson:"type,omitempty"`
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
	Qualifiers Qualifiers `json:"qualifiers,omitempty" bson:"qualifiers,omitempty"` // nolint: lll
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
	Labels map[string]string `json:"labels,omitempty" bson:"labels,omitempty"`
	// ShortTitle is an optional, succinct title for the Event, ideal for use in
	// lists or in scenarios where UI real estate is constrained.
	ShortTitle string `json:"shortTitle,omitempty" bson:"shortTitle,omitempty"`
	// LongTitle is an optional, detailed title for the Event.
	LongTitle string `json:"longTitle,omitempty" bson:"longTitle,omitempty"`
	// Git contains git-specific Event details. These can be used to override
	// similar details defined at the Project level. This is useful for scenarios
	// wherein an Event may need to convey an alternative source, branch, etc.
	Git *GitDetails `json:"git,omitempty" bson:"git,omitempty"`
	// Payload optionally contains Event details provided by the upstream system
	// that was the original source of the event. Payloads MUST NOT contain
	// sensitive information. Since Projects SUBSCRIBE to Events, the potential
	// exists for any Project to express an interest in any or all Events. This
	// being the case, sensitive details must never be present in payloads. The
	// common workaround work this constraint (which is also a sensible practice
	// to begin with) is that event payloads may contain REFERENCES to sensitive
	// details that are useful only to properly configured Workers.
	Payload string `json:"payload,omitempty" bson:"payload,omitempty"`
	// Worker contains details of the Worker assigned to handle the Event.
	Worker Worker `json:"worker" bson:"worker"`
}

// MarshalJSON amends Event instances with type metadata.
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

// SourceState encapsulates opaque, source-specific (e.g. gateway-specific)
// state.
type SourceState struct {
	// State is a map of arbitrary and opaque key/value pairs that the source of
	// an Event (e.g. the gateway that created it) can use to store
	// source-specific state.
	State map[string]string `json:"state,omitempty" bson:"state,omitempty"`
}

// GitDetails represents git-specific Event details. These may override
// Project-level GitConfig.
type GitDetails struct {
	// CloneURL specifies the location from where a source code repository may
	// be cloned.
	CloneURL string `json:"cloneURL,omitempty" bson:"cloneURL,omitempty"`
	// Commit specifies a commit (by sha) to be checked out.
	Commit string `json:"commit,omitempty" bson:"commit,omitempty"`
	// Ref specifies a tag or branch to be checked out. If left blank, this will
	// default to "master" at runtime.
	Ref string `json:"ref,omitempty" bson:"ref,omitempty"`
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
	Qualifiers Qualifiers
	// Labels specifies that only Events labeled with these key/value pairs should
	// be selected.
	Labels map[string]string
}

// EventList is an ordered and pageable list of Events.
type EventList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Events.
	Items []Event `json:"items,omitempty"`
}

// MarshalJSON amends EventList instances with type metadata.
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

// CancelManyEventsResult represents a summary of a mass Event cancellation
// operation.
type CancelManyEventsResult struct {
	// Count represents the number of Events canceled.
	Count int64 `json:"count"`
}

// MarshalJSON amends CancelManyEventsResult instances with type metadata.
func (c CancelManyEventsResult) MarshalJSON() ([]byte, error) {
	type Alias CancelManyEventsResult
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "CancelManyEventsResult",
			},
			Alias: (Alias)(c),
		},
	)
}

// DeleteManyEventsResult represents a summary of a mass Event deletion
// operation.
type DeleteManyEventsResult struct {
	// Count represents the number of Events deleted.
	Count int64 `json:"count"`
}

// MarshalJSON amends DeleteManyEventsResult instances with type metadata.
func (d DeleteManyEventsResult) MarshalJSON() ([]byte, error) {
	type Alias DeleteManyEventsResult
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "DeleteManyEventsResult",
			},
			Alias: (Alias)(d),
		},
	)
}

// EventsService is the specialized interface for managing Events. It's
// decoupled from underlying technology choices (e.g. data store, message bus,
// etc.) to keep business logic reusable and consistent while the underlying
// tech stack remains free to change.
type EventsService interface {
	// Create creates a new Event.
	Create(context.Context, Event) (
		EventList,
		error,
	)
	// List retrieves an EventList, with its Items (Events) ordered by age, newest
	// first. Criteria for which Events should be retrieved can be specified using
	// the EventListOptions parameter.
	List(
		context.Context,
		EventsSelector,
		meta.ListOptions,
	) (EventList, error)
	// Get retrieves a single Event specified by its identifier. If no such event
	// is found, implementations MUST return a *meta.ErrNotFound error.
	Get(context.Context, string) (Event, error)
	// GetByWorkerToken retrieves a single Event specified by its Worker's token.
	// If no such event is found, implementations MUST return a *meta.ErrNotFound
	// error.
	GetByWorkerToken(context.Context, string) (Event, error)
	// Clones an Event and creates a new Event after removing the original's
	// metadata and Worker configuration
	Clone(context.Context, string) (Event, error)
	// UpdateSourceState updates source-specific (e.g. gateway-specific) Event
	// state.
	UpdateSourceState(context.Context, string, SourceState) error
	// Cancel cancels a single Event specified by its identifier. If no such event
	// is found, implementations MUST return a *meta.ErrNotFound error.
	// Implementations MUST only cancel events whose Workers have not already
	// reached a terminal state. If the specified Event's Worker has already
	// reached a terminal state, implementations MUST return a *meta.ErrConflict.
	Cancel(context.Context, string) error
	// CancelMany cancels multiple Events specified by the EventsSelector
	// parameter. Implementations MUST only cancel events whose Workers have not
	// already reached a terminal state.
	CancelMany(
		context.Context,
		EventsSelector,
	) (CancelManyEventsResult, error)
	// Delete unconditionally deletes a single Event specified by its identifier.
	// If no such event is found, implementations MUST return a *meta.ErrNotFound
	// error.
	Delete(context.Context, string) error
	// DeleteMany unconditionally deletes multiple Events specified by the
	// EventsSelector parameter.
	DeleteMany(
		context.Context,
		EventsSelector,
	) (DeleteManyEventsResult, error)
	// Retry copies an Event, including Worker configuration, and creates a new
	// Event from this information.  Note that it does not carry over Job state
	// in the process.
	Retry(context.Context, string) (EventList, error)
}

type eventsService struct {
	authorize           libAuthz.AuthorizeFn
	projectAuthorize    ProjectAuthorizeFn
	projectsStore       ProjectsStore
	eventsStore         EventsStore
	logsStore           CoolLogsStore
	substrate           Substrate
	createSingleEventFn func(context.Context, Project, Event) (Event, error)
}

// NewEventsService returns a specialized interface for managing Events.
func NewEventsService(
	authorizeFn libAuthz.AuthorizeFn,
	projectAuthorize ProjectAuthorizeFn,
	projectsStore ProjectsStore,
	eventsStore EventsStore,
	logsStore CoolLogsStore,
	substrate Substrate,
) EventsService {
	e := &eventsService{
		authorize:        authorizeFn,
		projectAuthorize: projectAuthorize,
		projectsStore:    projectsStore,
		eventsStore:      eventsStore,
		logsStore:        logsStore,
		substrate:        substrate,
	}
	e.createSingleEventFn = e.createSingleEvent
	return e
}

func (e *eventsService) Create(
	ctx context.Context,
	event Event,
) (EventList, error) {
	events := EventList{}

	if event.ProjectID == "" {
		// This event doesn't reference a discrete project and is instead going to
		// be matched to all subscribing projects, so the only access requirement is
		// that the principal is permitted to create events from the specified
		// source. i.e. In practice, this would be how we make access decisions on
		// events coming from gateways.
		if err := e.authorize(
			ctx,
			system.RoleEventCreator,
			event.Source,
		); err != nil {
			return events, err
		}
	} else {
		// This event references a discrete project, so the access requirement is
		// that the principal is permitted to create events for the specified
		// project. i.e. In practice, this would be how we make access decisions on
		// events coming from a Brigade user.
		if err := e.projectAuthorize(
			ctx,
			event.ProjectID,
			RoleProjectUser,
		); err != nil {
			return events, err
		}
	}

	now := time.Now().UTC()
	event.Created = &now

	if event.ProjectID != "" {
		project, err := e.projectsStore.Get(ctx, event.ProjectID)
		if err != nil {
			return events, errors.Wrapf(
				err,
				"error retrieving project %q from store",
				event.ProjectID,
			)
		}
		evt, err := e.createSingleEventFn(ctx, project, event)
		events.Items = []Event{evt}
		return events, err
	}

	// If we get to here, no project ID is specified, so we search for projects
	// that are subscribed to this event. We iterate over all of those and create
	// a discrete event for each of these.
	projects, err := e.projectsStore.ListSubscribers(ctx, event)
	if err != nil {
		return events, errors.Wrap(
			err,
			"error retrieving subscribed projects from store",
		)
	}
	events.Items = make([]Event, len(projects.Items))
	for i, project := range projects.Items {
		event.ProjectID = project.ID
		evt, err := e.createSingleEventFn(ctx, project, event)
		if err != nil {
			return events, err
		}
		events.Items[i] = evt
	}
	return events, nil
}

func (e *eventsService) createSingleEvent(
	ctx context.Context,
	project Project,
	event Event,
) (Event, error) {

	event.ID = uuid.NewV4().String()

	// Defer to the Worker.Spec on the event itself, which may already exist if
	// the event is a retry of another.  Else, proceed with inheriting from the
	// project.Spec.  Since a Worker would minimally have a HashedToken from
	// prior creation, this is what we check to determine if the event.Worker was
	// previously initialized.
	workerSpec := event.Worker.Spec
	if event.Worker.HashedToken == "" {
		workerSpec = project.Spec.WorkerTemplate
	}

	if workerSpec.WorkspaceSize == "" {
		workerSpec.WorkspaceSize = defaultWorkspaceSize
	}

	// If they exist, git details from the event override any pre-existing git
	// config
	if event.Git != nil {
		if workerSpec.Git == nil {
			workerSpec.Git = &GitConfig{}
		}
		if event.Git.CloneURL != "" {
			workerSpec.Git.CloneURL = event.Git.CloneURL
		}
		if event.Git.Commit != "" {
			workerSpec.Git.Commit = event.Git.Commit
		}
		if event.Git.Ref != "" {
			workerSpec.Git.Ref = event.Git.Ref
		}
	}

	// If no commit (sha) or ref (branch or tag) is specified, default to the
	// master branch
	if workerSpec.Git != nil {
		if workerSpec.Git.Commit == "" && workerSpec.Git.Ref == "" {
			workerSpec.Git.Ref = "refs/heads/master"
		}
	}

	// If no log level is specified, default to INFO
	if workerSpec.LogLevel == "" {
		workerSpec.LogLevel = LogLevelInfo
	}

	if workerSpec.ConfigFilesDirectory == "" {
		workerSpec.ConfigFilesDirectory = ".brigade"
	}

	// This is a token unique to the Event so that the Event's Worker can use when
	// communicating with the API server to do things like spawn a new Job. i.e.
	// Only THIS event's worker can create new Jobs for THIS event.
	token := crypto.NewToken(256)

	event.Worker = Worker{
		Spec: workerSpec,
		Status: WorkerStatus{
			Phase: WorkerPhasePending,
		},
		// Note: The cleartext Token field doesn't get persisted to the data store
		Token:       token,
		HashedToken: crypto.Hash("", token),
	}

	// Persist the Event
	if err := e.eventsStore.Create(ctx, event); err != nil {
		return event, errors.Wrapf(
			err,
			"error storing new event %q",
			event.ID,
		)
	}

	// Prepare the substrate for the Worker and schedule the Worker for async /
	// eventual execution
	if err := e.substrate.ScheduleWorker(ctx, project, event); err != nil {
		return event, errors.Wrapf(
			err,
			"error scheduling event %q worker on the substrate",
			event.ID,
		)
	}

	return event, nil
}

func (e *eventsService) List(
	ctx context.Context,
	selector EventsSelector,
	opts meta.ListOptions,
) (EventList, error) {
	if err := e.authorize(ctx, system.RoleReader, ""); err != nil {
		return EventList{}, err
	}

	// If no worker phase filters were applied, retrieve all phases
	if len(selector.WorkerPhases) == 0 {
		selector.WorkerPhases = WorkerPhasesAll()
	}
	if opts.Limit == 0 {
		opts.Limit = 20
	}

	events, err := e.eventsStore.List(ctx, selector, opts)
	if err != nil {
		return events, errors.Wrap(err, "error retrieving events from store")
	}
	return events, nil
}

func (e *eventsService) Get(
	ctx context.Context,
	id string,
) (Event, error) {
	if err := e.authorize(ctx, system.RoleReader, ""); err != nil {
		return Event{}, err
	}

	event, err := e.eventsStore.Get(ctx, id)
	if err != nil {
		return event, errors.Wrapf(err, "error retrieving event %q from store", id)
	}
	return event, nil
}

func (e *eventsService) GetByWorkerToken(
	ctx context.Context,
	workerToken string,
) (Event, error) {
	// No authz is required here because this is only ever called by the system
	// itself.

	event, err := e.eventsStore.GetByHashedWorkerToken(
		ctx,
		crypto.Hash("", workerToken),
	)
	if err != nil {
		return event, errors.Wrap(err, "error retrieving event from store")
	}
	return event, nil
}

func (e *eventsService) Clone(
	ctx context.Context,
	id string,
) (Event, error) {
	// No authz call here as we'll defer to the checks in e.Create() invoked
	// below

	event, err := e.eventsStore.Get(ctx, id)
	if err != nil {
		return Event{}, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			id,
		)
	}

	// Clone all event details *except* metadata and worker config
	clone := event
	clone.ObjectMeta = meta.ObjectMeta{}
	clone.Worker = Worker{}

	// Add a label for tracing the original cloned event id
	if clone.Labels == nil {
		clone.Labels = map[string]string{}
	}
	clone.Labels["cloneOf"] = id

	events, err := e.Create(ctx, clone)
	if err != nil {
		return Event{}, err
	}

	return events.Items[0], nil
}

func (e *eventsService) UpdateSourceState(
	ctx context.Context,
	id string,
	sourceState SourceState,
) error {
	event, err := e.eventsStore.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", id)
	}

	if err = e.authorize(ctx, system.RoleEventCreator, event.Source); err != nil {
		return err
	}

	err = e.eventsStore.UpdateSourceState(ctx, id, sourceState)
	return errors.Wrapf(
		err,
		"error updating source state of event %q in store",
		id,
	)
}

func (e *eventsService) Cancel(ctx context.Context, id string) error {
	event, err := e.eventsStore.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", id)
	}

	if err =
		e.projectAuthorize(ctx, event.ProjectID, RoleProjectUser); err != nil {
		return err
	}

	project, err := e.projectsStore.Get(ctx, event.ProjectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			event.ProjectID,
		)
	}

	if err = e.eventsStore.Cancel(ctx, id); err != nil {
		return errors.Wrapf(err, "error canceling event %q in store", id)
	}

	if err = e.substrate.DeleteWorkerAndJobs(ctx, project, event); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q worker and jobs from the substrate",
			id,
		)
	}

	return nil
}

func (e *eventsService) CancelMany(
	ctx context.Context,
	selector EventsSelector,
) (CancelManyEventsResult, error) {
	result := CancelManyEventsResult{}

	// Refuse requests not qualified by project
	if selector.ProjectID == "" {
		return result, &meta.ErrBadRequest{
			Reason: "Requests to cancel multiple events must be qualified by " +
				"project.",
		}
	}

	if err :=
		e.projectAuthorize(ctx, selector.ProjectID, RoleProjectUser); err != nil {
		return CancelManyEventsResult{}, err
	}

	// Refuse requests not qualified by worker phases
	if len(selector.WorkerPhases) == 0 {
		return result, &meta.ErrBadRequest{
			Reason: "Requests to cancel multiple events must be qualified by " +
				"worker phase(s).",
		}
	}

	project, err := e.projectsStore.Get(ctx, selector.ProjectID)
	if err != nil {
		return result, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			selector.ProjectID,
		)
	}

	eventCh, affectedCount, err := e.eventsStore.CancelMany(ctx, selector)
	if err != nil {
		return result, errors.Wrap(err, "error canceling events in store")
	}
	result.Count = affectedCount

	// Fan out to a finite number of goroutines to handle cleanup duties
	concurrency := 10
	if result.Count < int64(concurrency) {
		concurrency = int(result.Count)
	}
	for i := 0; i < concurrency; i++ {
		go func() {
			for event := range eventCh {
				if err := e.substrate.DeleteWorkerAndJobs(
					context.Background(), // deliberately not using ctx
					project,
					event,
				); err != nil {
					log.Println(errors.Wrapf(
						err,
						"error deleting event %q worker and jobs from the substrate",
						event.ID,
					))
				}
			}
		}()
	}

	return result, nil
}

func (e *eventsService) Delete(ctx context.Context, id string) error {
	event, err := e.eventsStore.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", id)
	}

	if err =
		e.projectAuthorize(ctx, event.ProjectID, RoleProjectUser); err != nil {
		return err
	}

	project, err := e.projectsStore.Get(ctx, event.ProjectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			event.ProjectID,
		)
	}

	if err = e.eventsStore.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "error deleting event %q from store", id)
	}

	if err = e.substrate.DeleteWorkerAndJobs(ctx, project, event); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q worker and jobs from the substrate",
			id,
		)
	}

	return e.logsStore.DeleteEventLogs(ctx, id)
}

func (e *eventsService) DeleteMany(
	ctx context.Context,
	selector EventsSelector,
) (DeleteManyEventsResult, error) {
	result := DeleteManyEventsResult{}

	// Refuse requests not qualified by project
	if selector.ProjectID == "" {
		return result, &meta.ErrBadRequest{
			Reason: "Requests to delete multiple events must be qualified by " +
				"project.",
		}
	}

	if err := e.projectAuthorize(
		ctx,
		selector.ProjectID,
		RoleProjectUser,
	); err != nil {
		return DeleteManyEventsResult{}, err
	}

	// Refuse requests not qualified by worker phases
	if len(selector.WorkerPhases) == 0 {
		return result, &meta.ErrBadRequest{
			Reason: "Requests to delete multiple events must be qualified by " +
				"worker phase(s).",
		}
	}

	project, err := e.projectsStore.Get(ctx, selector.ProjectID)
	if err != nil {
		return result, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			selector.ProjectID,
		)
	}

	eventCh, affectedCount, err := e.eventsStore.DeleteMany(ctx, selector)
	if err != nil {
		return result, errors.Wrap(err, "error deleting events from store")
	}
	result.Count = affectedCount

	// Fan out to a finite number of goroutines to handle cleanup duties
	concurrency := 10
	if result.Count < int64(concurrency) {
		concurrency = int(result.Count)
	}
	for i := 0; i < concurrency; i++ {
		go func() {
			for event := range eventCh {
				if err := e.substrate.DeleteWorkerAndJobs(
					context.Background(), // deliberately not using ctx
					project,
					event,
				); err != nil {
					log.Println(errors.Wrapf(
						err,
						"error deleting event %q worker and jobs from the substrate",
						event.ID,
					))
				}

				if err := e.logsStore.DeleteEventLogs(
					context.Background(), // deliberately not using ctx
					event.ID,
				); err != nil {
					log.Println(errors.Wrapf(
						err,
						"error deleting logs for event %q",
						event.ID,
					))
				}
			}
		}()
	}

	return result, nil
}

func (e *eventsService) Retry(
	ctx context.Context,
	id string,
) (EventList, error) {
	// No authz call here as we'll defer to the checks in e.Create() invoked
	// below

	event, err := e.eventsStore.Get(ctx, id)
	if err != nil {
		return EventList{}, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			id,
		)
	}

	// Copy all event details, including worker configuration, only omitting
	// metadata and worker jobs array
	retry := event
	retry.ObjectMeta = meta.ObjectMeta{}
	// TODO: logic for:
	// "skips jobs that succeeded already and retries those that didn't, assuming no shared volumes"
	// "copy the battle plan also; clear the state on failed jobs."
	// "We could then amend the logic for Job creation to short circuit when a "new" job is created,
	// but it already exists and is already in a success state."
	retry.Worker.Jobs = nil

	// Add a label for tracing the original event id
	if retry.Labels == nil {
		retry.Labels = map[string]string{}
	}
	retry.Labels["retryOf"] = id

	return e.Create(ctx, retry)
}

// EventsStore is an interface for components that implement Event persistence
// concerns.
type EventsStore interface {
	// Create persists a new Event in the underlying data store. If n Event having
	// the same ID already exists, implementations MUST return a *meta.ErrConflict
	// error.
	Create(context.Context, Event) error
	// List retrieves an EventList from the underlying data store, with its Items
	// (Events) ordered by age, newest first. Criteria for which Events should be
	// retrieved can be specified using the EventListOptions parameter.
	List(
		context.Context,
		EventsSelector,
		meta.ListOptions,
	) (EventList, error)
	// Get retrieves a single Event from the underlying data store. If the
	// specified Event does not exist, implementations MUST return a
	// *meta.ErrNotFound error.
	Get(context.Context, string) (Event, error)
	// GetByHashedWorkerToken retrieves a single Event from the underlying data
	// store by the provided hashed Worker token. If no such Event exists,
	// implementations MUST return a *meta.ErrNotFound error.
	GetByHashedWorkerToken(context.Context, string) (Event, error)
	// UpdateSourceState updates source-specific (e.g. gateway-specific) Event
	// state. Implementations MAY assume the Event's existence has been
	// pre-confirmed by the caller.
	UpdateSourceState(context.Context, string, SourceState) error
	// Cancel updates the specified Event in the underlying data store to reflect
	// that it has been canceled. Implementations MAY assume the Event's existence
	// has been pre-confirmed by the caller. Implementations MUST only cancel
	// events whose Workers have not already reached a terminal state. If the
	// specified Event's Worker has already reached a terminal state,
	// implementations MUST return a *meta.ErrConflict.
	Cancel(context.Context, string) error
	// CancelMany updates multiple Events specified by the EventsSelector
	// parameter in the underlying data store to reflect that they have been
	// canceled. Implementations MUST only cancel events whose Workers have not
	// already reached a terminal state and MUST return the total number of
	// canceled events.
	CancelMany(context.Context, EventsSelector) (<-chan Event, int64, error)
	// Delete unconditionally deletes the specified Event from the underlying data
	// store. If the specified Event does not exist, implementations MUST
	// return a *meta.ErrNotFound error.
	Delete(context.Context, string) error
	// DeleteMany unconditionally deletes multiple Events specified by the
	// EventsSelector parameter from the underlying data store.  Implementations
	// MUST return the total number of deleted events.
	DeleteMany(context.Context, EventsSelector) (<-chan Event, int64, error)
}
