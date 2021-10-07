package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/brigadecore/brigade-foundations/retries"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	myk8s "github.com/brigadecore/brigade/v2/internal/kubernetes"
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
	authorize        AuthorizeFn
	projectAuthorize ProjectAuthorizeFn
	projectsStore    ProjectsStore
	eventsStore      EventsStore
	warmLogsStore    LogsStore
	coolLogsStore    LogsStore
}

// NewLogsService returns a specialized interface for accessing logs.
func NewLogsService(
	authorize AuthorizeFn,
	projectAuthorize ProjectAuthorizeFn,
	projectsStore ProjectsStore,
	eventsStore EventsStore,
	warmLogsStore LogsStore,
	coolLogsStore LogsStore,
) LogsService {
	return &logsService{
		authorize:        authorize,
		projectAuthorize: projectAuthorize,
		projectsStore:    projectsStore,
		eventsStore:      eventsStore,
		warmLogsStore:    warmLogsStore,
		coolLogsStore:    coolLogsStore,
	}
}

// nolint: gocyclo
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
			selector.Container = myk8s.LabelKeyWorker
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

	// Throughout the service layer, we typically only require RoleReader() to
	// authorize read-only operations of any kind. In the case of logs, however,
	// there's just too much possibility of secrets bleeding into the logs, not
	// due to any fault of Brigade's but because of some end-user misstep. So, out
	// of an abundance of caution, we raise the bar a little on this one read-only
	// operation and require the principal to be a project user in order to stream
	// logs.
	err = l.projectAuthorize(ctx, event.ProjectID, RoleProjectUser)
	if err != nil {
		// We also permit access by the event's worker
		err = l.authorize(ctx, RoleWorker, event.ID)
	}
	if err != nil {
		// We also permit access by the creator of the event. This enables smarter
		// gateways to send logs "upstream" if appropriate.
		err = l.authorize(ctx, RoleEventCreator, event.Source)
	}
	if err != nil {
		return nil, err
	}

	var containerFound bool
	if selector.Job == "" {
		// If we're here, we want worker logs.
		if selector.Container != myk8s.LabelKeyWorker &&
			!(selector.Container == "vcs" && event.Worker.Spec.Git != nil) {
			return nil, &meta.ErrNotFound{
				Type: "WorkerContainer",
				ID:   selector.Container,
			}
		}
		containerFound = true
	} else {
		// If we're here, we want logs from a specific job. Make sure that job
		// exists.
		job, ok := event.Worker.Job(selector.Job)
		if !ok {
			return nil, &meta.ErrNotFound{
				Type: JobKind,
				ID:   selector.Job,
			}
		}
		// And make sure the container exists.
		if selector.Container == job.Name {
			// If the container name matches the job name, it's a request for logs
			// from the primary container. That always exists.
			containerFound = true
		} else if selector.Container == "vcs" {
			// vcs is a valid container name IF at least one of the job's containers
			// use source from git.
			containerFound = job.Spec.PrimaryContainer.SourceMountPath != ""
			if !containerFound {
				// If we get to here, the primary container didn't use source, so check
				// if any of the sidecars do.
				for _, containerSpec := range job.Spec.SidecarContainers {
					if containerSpec.SourceMountPath != "" {
						containerFound = true
						break
					}
				}
			}
		} else {
			// If we get to here, the container name didn't match the job name (which
			// is also the name of the primary container) and it wasn't "vcs" either.
			// Just loop through the sidecars to see if such a container exists.
			for containerName := range job.Spec.SidecarContainers {
				if containerName == selector.Container {
					containerFound = true
					break
				}
			}
		}
		if !containerFound {
			return nil, &meta.ErrNotFound{
				Type: "JobContainer",
				ID:   selector.Container,
			}
		}

		// Check to see if we need to look up logs via a specific event ID,
		// as job may be cached and carried over on a retry event
		if job.Status != nil && job.Status.LogsEventID != "" {
			event, err = l.eventsStore.Get(ctx, job.Status.LogsEventID)
			if err != nil {
				if _, ok := err.(*meta.ErrNotFound); ok {
					return nil,
						&meta.ErrNotFound{
							Reason: fmt.Sprintf(
								"Unable to retrieve logs for job %q: the "+
									"original logs inherited by this job no longer exist.",
								job.Name,
							),
						}
				}
				return nil,
					errors.Wrapf(
						err,
						"error retrieving logs for job %q",
						job.Name,
					)
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

	// Wait for the target Worker or Job to move past PENDING and STARTING phases
	if err = retries.ManageRetries(
		ctx,
		"waiting for worker or job to move past PENDING and STARTING phases",
		50, // A generous number of retries. Let the client hang up if they want.
		20*time.Second,
		func() (bool, error) {
			if event, err = l.eventsStore.Get(ctx, event.ID); err != nil {
				return false, errors.Wrapf(
					err,
					"error retrieving event %q from store",
					event.ID,
				)
			}
			if selector.Job == "" { // Worker...
				// If the Event's Worker's phase is PENDING or STARTING, then retry.
				// Otherwise, exit the retry loop.
				return event.Worker.Status.Phase == WorkerPhasePending ||
					event.Worker.Status.Phase == WorkerPhaseStarting, nil
			}
			// Else Job...
			// If the Job's phase is PENDING or STARTING, then retry.
			// Otherwise, exit the retry loop.
			job, _ := event.Worker.Job(selector.Job)
			return job.Status.Phase == JobPhasePending ||
				job.Status.Phase == JobPhaseStarting, nil
		},
	); err != nil {
		return nil, err
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

// CoolLogsStore is an interface for components that implement "cool" Log
// persistence concerns.  These log store types are intended to act as
// longterm storehouses for worker and job logs after they have reached a
// terminal state. Thus, log deletion methods are prudent for managing
// the size of the underlying store.
type CoolLogsStore interface {
	LogsStore

	// DeleteEventLogs deletes all logs associated with the provided event.
	DeleteEventLogs(ctx context.Context, id string) error

	// DeleteProjectLogs deletes all logs associated with the provided project.
	DeleteProjectLogs(ctx context.Context, id string) error
}
