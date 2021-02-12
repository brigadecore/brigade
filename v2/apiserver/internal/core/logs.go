package core

import (
	"context"
	"encoding/json"
	"time"

	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

// LogsSelector represents useful criteria for selecting logs to be streamed
// from any container belonging to some Worker OR any container belonging to
// Jobs spawned by that Worker.
type LogsSelector struct {
	// Job specifies, by name, a Job spawned by some Worker. If not specified, log
	// streaming operations presume logs are desired for the Worker itself.
	Job string
	// Container specifies, by name, a container belonging to some Worker or, if
	// Job is specified, that Job. If not specified, log streaming operations
	// presume logs are desired from a container having the same name as the
	// selected Worker or Job.
	Container string
}

// LogStreamOptions represents useful options for streaming logs from some
// container of a Worker or Job.
type LogStreamOptions struct {
	// Follow indicates whether the stream should conclude after the last
	// available line of logs has been sent to the client (false) or remain open
	// until closed by the client (true), continuing to send new lines as they
	// become available.
	Follow bool `json:"follow"`
}

// LogEntry represents one line of output from an OCI container.
type LogEntry struct {
	// Time is the time the line was written.
	Time *time.Time `json:"time,omitempty" bson:"time,omitempty"`
	// Message is a single line of log output from an OCI container.
	Message string `json:"message,omitempty" bson:"log,omitempty"`
}

// MarshalJSON amends LogEntry instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
func (l LogEntry) MarshalJSON() ([]byte, error) {
	type Alias LogEntry
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "LogEntry",
			},
			Alias: (Alias)(l),
		},
	)
}

// LogsService is the specialized interface for accessing logs. It's
// decoupled from underlying technology choices (e.g. data store, message bus,
// etc.) to keep business logic reusable and consistent while the underlying
// tech stack remains free to change.
type LogsService interface {
	// Stream returns a channel over which logs for an Event's Worker, or using
	// the LogsSelector parameter, a Job spawned by that Worker (or specific
	// container thereof), are streamed. If the specified Event, Job, or Container
	// thereof does not exist, implementations MUST return a *meta.ErrNotFound
	// error.
	Stream(
		ctx context.Context,
		eventID string,
		selector LogsSelector,
		opts LogStreamOptions,
	) (<-chan LogEntry, error)
}

type logsService struct {
	authorize     libAuthz.AuthorizeFn
	projectsStore ProjectsStore
	eventsStore   EventsStore
	warmLogsStore LogsStore
	coolLogsStore LogsStore
}

// NewLogsService returns a specialized interface for accessing logs.
func NewLogsService(
	authorizeFn libAuthz.AuthorizeFn,
	projectsStore ProjectsStore,
	eventsStore EventsStore,
	warmLogsStore LogsStore,
	coolLogsStore LogsStore,
) LogsService {
	return &logsService{
		authorize:     authorizeFn,
		projectsStore: projectsStore,
		eventsStore:   eventsStore,
		warmLogsStore: warmLogsStore,
		coolLogsStore: coolLogsStore,
	}
}

func (l *logsService) Stream(
	ctx context.Context,
	eventID string,
	selector LogsSelector,
	opts LogStreamOptions,
) (<-chan LogEntry, error) {
	// Set defaults on the selector
	if selector.Job == "" { // If a job isn't specified, then we want worker logs
		if selector.Container == "" {
			// If a container isn't specified, we want the one named "worker"
			selector.Container = "worker"
		}
		// These are the only legitimate container names for ANY worker
		if selector.Container != "worker" && selector.Container != "vcs" {
			// Any other container name is an error
			return nil, &meta.ErrNotFound{
				Type: "WorkerContainer",
				ID:   selector.Container,
			}
		}
	} else { // A job was specified, so we want job logs
		if selector.Container == "" {
			// If a container isn't specified, we want the primary container's logs.
			// The primary container has the same name as the job itself.
			selector.Container = selector.Job
		}
	}

	event, err := l.eventsStore.Get(ctx, eventID)
	if err != nil {
		return nil,
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}

	// Throughout the service layer, we typically only require system.RoleReader()
	// to authorize read-only operations of any kind. In the case of logs,
	// however, there's just too much possibility of secrets bleeding into the
	// logs, not due to any fault of Brigade's but because of some end-user
	// misstep. So, out of an abundance of caution, we raise the bar a little on
	// this one read-only operation and require the principal to be a project user
	// in order to stream logs.
	if err = l.authorize(ctx, RoleProjectUser(event.ProjectID)); err != nil {
		return nil, err
	}

	if selector.Job != "" {
		// Make sure the job exists
		job, ok := event.Worker.Job(selector.Job)
		if !ok {
			return nil, &meta.ErrNotFound{
				Type: "Job",
				ID:   selector.Job,
			}
		}
		if selector.Container != selector.Job {
			if _, ok := job.Spec.SidecarContainers[selector.Container]; !ok {
				return nil, &meta.ErrNotFound{
					Type: "JobContainer",
					ID:   selector.Container,
				}
			}
		}
	}

	// Make sure the project exists
	project, err := l.projectsStore.Get(ctx, event.ProjectID)
	if err != nil {
		return nil,
			errors.Wrapf(
				err,
				"error retrieving project %q from store",
				event.ProjectID,
			)
	}

	logCh, err := l.warmLogsStore.StreamLogs(ctx, project, event, selector, opts)
	if err != nil {
		// If the issue is simply that the warmLogsStore couldn't find the logs
		// (realistically, this is because the underlying pod no longer exists),
		// then fall back to the coolLogsStore.
		if _, ok := errors.Cause(err).(*meta.ErrNotFound); ok {
			logCh, err =
				l.coolLogsStore.StreamLogs(ctx, project, event, selector, opts)
		}
	}
	return logCh, err
}

// LogsStore is an interface for components that implement Log persistence
// concerns.
type LogsStore interface {
	// Stream returns a channel over which logs for an Event's Worker, or using
	// the LogsSelector parameter, a Job spawned by that Worker (or specific
	// container thereof), are streamed. If the specified Event, Job, or Container
	// thereof does not exist, implementations MUST return a *meta.ErrNotFound
	// error.
	StreamLogs(
		ctx context.Context,
		project Project,
		event Event,
		selector LogsSelector,
		opts LogStreamOptions,
	) (<-chan LogEntry, error)
}
