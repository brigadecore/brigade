package core

import (
	"context"
	"fmt"
	"log"
	"time"

	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/brigadecore/brigade/v2/apiserver/internal/system"
	"github.com/pkg/errors"
)

// LogLevel represents the desired granularity of Worker log output.
type LogLevel string

// LogLevelInfo represents INFO level granularity in Worker log output.
const LogLevelInfo LogLevel = "INFO"

// WorkerPhase represents where a Worker is within its lifecycle.
type WorkerPhase string

const (
	// WorkerPhaseAborted represents the state wherein a Worker was forcefully
	// stopped during execution.
	WorkerPhaseAborted WorkerPhase = "ABORTED"
	// WorkerPhaseCanceled represents the state wherein a pending Worker was
	// canceled prior to execution.
	WorkerPhaseCanceled WorkerPhase = "CANCELED"
	// WorkerPhaseFailed represents the state wherein a Worker has run to
	// completion but experienced errors.
	WorkerPhaseFailed WorkerPhase = "FAILED"
	// WorkerPhasePending represents the state wherein a Worker is awaiting
	// execution.
	WorkerPhasePending WorkerPhase = "PENDING"
	// WorkerPhaseRunning represents the state wherein a Worker is currently
	// being executed.
	WorkerPhaseRunning WorkerPhase = "RUNNING"
	// WorkerPhaseSchedulingFailed represents the state wherein a Worker was not
	// scheduled due to some unexpected and unrecoverable error encountered by the
	// scheduler.
	WorkerPhaseSchedulingFailed WorkerPhase = "SCHEDULING_FAILED"
	// WorkerPhaseStarting represents the state wherein a Worker is starting on
	// the substrate but isn't running yet.
	WorkerPhaseStarting WorkerPhase = "STARTING"
	// WorkerPhaseSucceeded represents the state where a Worker has run to
	// completion without error.
	WorkerPhaseSucceeded WorkerPhase = "SUCCEEDED"
	// WorkerPhaseTimedOut represents the state wherein a Worker has has not
	// completed within a designated timeframe.
	WorkerPhaseTimedOut WorkerPhase = "TIMED_OUT"
	// WorkerPhaseUnknown represents the state wherein a Worker's state is
	// unknown. Note that this is possible if and only if the underlying Worker
	// execution substrate (Kubernetes), for some unanticipated, reason does not
	// know the Worker's (Pod's) state.
	WorkerPhaseUnknown WorkerPhase = "UNKNOWN"
)

// WorkerPhasesAll returns a slice of WorkerPhases containing ALL possible
// phases. Note that instead of utilizing a package-level slice, this a function
// returns ad-hoc copies of the slice in order to preclude the possibility of
// this important collection being modified at runtime.
func WorkerPhasesAll() []WorkerPhase {
	return []WorkerPhase{
		WorkerPhaseAborted,
		WorkerPhaseCanceled,
		WorkerPhaseFailed,
		WorkerPhasePending,
		WorkerPhaseRunning,
		WorkerPhaseSucceeded,
		WorkerPhaseTimedOut,
		WorkerPhaseUnknown,
	}
}

// IsTerminal returns a bool indicating whether the WorkerPhase is terminal.
func (w WorkerPhase) IsTerminal() bool {
	switch w {
	case WorkerPhaseAborted:
		fallthrough
	case WorkerPhaseCanceled:
		fallthrough
	case WorkerPhaseFailed:
		fallthrough
	case WorkerPhaseSchedulingFailed:
		fallthrough
	case WorkerPhaseSucceeded:
		fallthrough
	case WorkerPhaseTimedOut:
		return true
	}
	return false
}

// Worker represents a component that orchestrates handling of a single Event.
type Worker struct {
	// Spec is the technical blueprint for the Worker.
	Spec WorkerSpec `json:"spec" bson:"spec"`
	// Status contains details of the Worker's current state.
	Status WorkerStatus `json:"status" bson:"status"`
	// Token is an API token that grants a Worker permission to create new Jobs
	// only for the Event to which it belongs.
	Token string `json:"-" bson:"-"`
	// HashedToken is a secure hash of the Token field.
	HashedToken string `json:"-" bson:"hashedToken"`
	// Jobs contains details of all Jobs spawned by the Worker during handling of
	// the Event.
	Jobs []Job `json:"jobs,omitempty" bson:"jobs"`
}

// Job retrieves a Job by name. It returns a boolean indicating whether the
// returned Job is the one requested (true) or a zero value (false) because no
// Job with the specified name belongs to this Worker.
func (w *Worker) Job(jobName string) (Job, bool) {
	for _, j := range w.Jobs {
		if j.Name == jobName {
			return j, true
		}
	}
	return Job{}, false
}

// WorkerSpec is the technical blueprint for a Worker.
type WorkerSpec struct {
	// Container specifies the details of an OCI container that forms the
	// cornerstone of the Worker.
	Container *ContainerSpec `json:"container,omitempty" bson:"container,omitempty"` // nolint: lll
	// UseWorkspace indicates whether the Worker and/or any Jobs it may spawn
	// requires access to a shared workspace. When false, no such workspace is
	// provisioned prior to Worker creation. This is a generally useful feature,
	// but by opting out of it (or rather, not opting-in), Job results can be made
	// cacheable and Jobs resumable/retriable-- something which cannot be done
	// otherwise since managing the state of the shared volume would require a
	// layered file system that we currently do not have.
	UseWorkspace bool `json:"useWorkspace" bson:"useWorkspace"`
	// WorkspaceSize specifies the size of a volume that will be provisioned as
	// a shared workspace for the Worker and any Jobs it spawns.
	// The value can be expressed in bytes (as a plain integer) or as a
	// fixed-point integer using one of these suffixes: E, P, T, G, M, K.
	// Power-of-two equivalents may also be used: Ei, Pi, Ti, Gi, Mi, Ki.
	WorkspaceSize string `json:"workspaceSize,omitempty" bson:"workspaceSize,omitempty"` // nolint: lll
	// Git contains git-specific Worker details.
	Git *GitConfig `json:"git,omitempty"`
	// Kubernetes contains Kubernetes-specific Worker details.
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes,omitempty"` // nolint: lll
	// JobPolicies specifies policies for any Jobs spawned by the Worker.
	JobPolicies *JobPolicies `json:"jobPolicies,omitempty" bson:"jobPolicies,omitempty"` // nolint: lll
	// LogLevel specifies the desired granularity of Worker log output.
	LogLevel LogLevel `json:"logLevel,omitempty" bson:"logLevel,omitempty"`
	// ConfigFilesDirectory specifies a directory within the Worker's workspace
	// where any relevant configuration files (e.g. brigade.json, brigade.js,
	// etc.) can be located.
	ConfigFilesDirectory string `json:"configFilesDirectory,omitempty" bson:"configFilesDirectory,omitempty"` // nolint: lll
	// DefaultConfigFiles is a map of configuration file names to configuration
	// file content. This is useful for Workers that do not integrate with any
	// source control system and would like to embed configuration (e.g.
	// brigade.json) or scripts (e.g. brigade.js) directly within the WorkerSpec.
	DefaultConfigFiles map[string]string `json:"defaultConfigFiles,omitempty" bson:"defaultConfigFiles,omitempty"` // nolint: lll
}

// GitConfig represents git-specific Worker details.
type GitConfig struct {
	// CloneURL specifies the location from where a source code repository may
	// be cloned.
	CloneURL string `json:"cloneURL,omitempty" bson:"cloneURL,omitempty"`
	// Commit specifies a revision (by SHA) to be checked out. If non-empty, this
	// field takes precedence over any value in the Ref field.
	Commit string `json:"commit,omitempty" bson:"commit,omitempty"`
	// Ref is a symbolic reference to a revision to be checked out. If non-empty,
	// the value of the Commit field supercedes any value in this field. Example
	// uses of this field include referencing a branch (refs/heads/<branch name>)
	// or a tag (refs/tags/<tag name>). If left blank, the default value
	// refs/heads/master will be assumed at runtime.
	Ref string `json:"ref,omitempty" bson:"ref,omitempty"`
	// InitSubmodules indicates whether to clone the repository's submodules.
	InitSubmodules bool `json:"initSubmodules" bson:"initSubmodules"`
}

// KubernetesConfig represents Kubernetes-specific Worker or Job configuration.
type KubernetesConfig struct {
	// ImagePullSecrets enumerates any image pull secrets that Kubernetes may use
	// when pulling the OCI image on which a Worker's or Job's container is based.
	// This field only needs to be utilized in the case of private, custom Worker
	// or Job images. The image pull secrets in question must be created
	// out-of-band by a sufficiently authorized user of the Kubernetes cluster.
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty" bson:"imagePullSecrets,omitempty"` // nolint: lll
}

// JobPolicies represents policies for any Jobs spawned by a Worker.
type JobPolicies struct {
	// AllowPrivileged specifies whether the Worker is permitted to launch Jobs
	// that utilize privileged containers.
	AllowPrivileged bool `json:"allowPrivileged" bson:"allowPrivileged"`
	// AllowDockerSocketMount specifies whether the Worker is permitted to launch
	// Jobs that mount the underlying host's Docker socket into its own file
	// system.
	AllowDockerSocketMount bool `json:"allowDockerSocketMount" bson:"allowDockerSocketMount"` // nolint: lll
}

// WorkerStatus represents the status of a Worker.
type WorkerStatus struct {
	// Started indicates the time the Worker began execution. It will be nil for
	// a Worker that is not yet executing.
	Started *time.Time `json:"started,omitempty" bson:"started,omitempty"`
	// Ended indicates the time the Worker concluded execution. It will be nil
	// for a Worker that is not done executing (or hasn't started).
	Ended *time.Time `json:"ended,omitempty" bson:"ended,omitempty"`
	// Phase indicates where the Worker is in its lifecycle.
	Phase WorkerPhase `json:"phase,omitempty" bson:"phase,omitempty"`
}

// WorkersService is the specialized interface for managing Workers. It's
// decoupled from underlying technology choices (e.g. data store, message bus,
// etc.) to keep business logic reusable and consistent while the underlying
// tech stack remains free to change.
type WorkersService interface {
	// Start starts the indicated Event's Worker on Brigade's workload
	// execution substrate. If the specified Event does not exist, implementations
	// MUST return a *meta.ErrNotFound.
	Start(ctx context.Context, eventID string) error
	// GetStatus returns an Event's Worker's status. If the specified Event does
	// not exist, implementations MUST return a *meta.ErrNotFound.
	GetStatus(
		ctx context.Context,
		eventID string,
	) (WorkerStatus, error)
	// WatchStatus returns a channel over which an Event's Worker's status is
	// streamed. The channel receives a new WorkerStatus every time there is any
	// change in that status. If the specified Event does not exist,
	// implementations MUST return a *meta.ErrNotFound.
	WatchStatus(
		ctx context.Context,
		eventID string,
	) (<-chan WorkerStatus, error)
	// UpdateStatus updates the status of an Event's Worker. If the specified
	// Event does not exist, implementations MUST return a *meta.ErrNotFound.
	UpdateStatus(
		ctx context.Context,
		eventID string,
		status WorkerStatus,
	) error
	// Cleanup removes Worker-related resources from the substrate, presumably
	// upon completion, without deleting the Worker from the data store.
	Cleanup(ctx context.Context, eventID string) error
}

type workersService struct {
	authorize     libAuthz.AuthorizeFn
	projectsStore ProjectsStore
	eventsStore   EventsStore
	workersStore  WorkersStore
	substrate     Substrate
}

// NewWorkersService returns a specialized interface for managing Workers.
func NewWorkersService(
	authorizeFn libAuthz.AuthorizeFn,
	projectsStore ProjectsStore,
	eventsStore EventsStore,
	workersStore WorkersStore,
	substrate Substrate,
) WorkersService {
	return &workersService{
		authorize:     authorizeFn,
		projectsStore: projectsStore,
		eventsStore:   eventsStore,
		workersStore:  workersStore,
		substrate:     substrate,
	}
}

func (w *workersService) Start(ctx context.Context, eventID string) error {
	if err := w.authorize(ctx, RoleScheduler()); err != nil {
		return err
	}

	event, err := w.eventsStore.Get(ctx, eventID)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}

	if event.Worker.Status.Phase != WorkerPhasePending {
		return &meta.ErrConflict{
			Type: "Event",
			ID:   event.ID,
			Reason: fmt.Sprintf(
				"Event %q worker has already been started.",
				event.ID,
			),
		}
	}

	project, err := w.projectsStore.Get(ctx, event.ProjectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			event.ProjectID,
		)
	}

	if err = w.workersStore.UpdateStatus(
		ctx,
		eventID,
		WorkerStatus{
			Phase: WorkerPhaseStarting,
		},
	); err != nil {
		return errors.Wrapf(
			err,
			"error updating status of event %q worker in store",
			eventID,
		)
	}

	if err = w.substrate.StartWorker(ctx, project, event); err != nil {
		return errors.Wrapf(err, "error starting worker for event %q", event.ID)
	}
	return nil
}

func (w *workersService) GetStatus(
	ctx context.Context,
	eventID string,
) (WorkerStatus, error) {
	if err := w.authorize(ctx, system.RoleReader()); err != nil {
		return WorkerStatus{}, err
	}

	event, err := w.eventsStore.Get(ctx, eventID)
	if err != nil {
		return WorkerStatus{},
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	return event.Worker.Status, nil
}

func (w *workersService) WatchStatus(
	ctx context.Context,
	eventID string,
) (<-chan WorkerStatus, error) {
	if err := w.authorize(ctx, system.RoleReader()); err != nil {
		return nil, err
	}

	// Read the event up front to confirm it exists.
	if _, err := w.eventsStore.Get(ctx, eventID); err != nil {
		return nil,
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	statusCh := make(chan WorkerStatus)
	go func() {
		defer close(statusCh)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
			event, err := w.eventsStore.Get(ctx, eventID)
			if err != nil {
				log.Printf("error retrieving event %q from store: %s", eventID, err)
				return
			}
			select {
			case statusCh <- event.Worker.Status:
			case <-ctx.Done():
				return
			}
		}
	}()
	return statusCh, nil
}

func (w *workersService) UpdateStatus(
	ctx context.Context,
	eventID string,
	status WorkerStatus,
) error {
	if err := w.authorize(ctx, RoleObserver()); err != nil {
		return err
	}

	if err := w.workersStore.UpdateStatus(
		ctx,
		eventID,
		status,
	); err != nil {
		return errors.Wrapf(
			err,
			"error updating status of event %q worker in store",
			eventID,
		)
	}
	return nil
}

func (w *workersService) Cleanup(
	ctx context.Context,
	eventID string,
) error {
	if err := w.authorize(ctx, RoleObserver()); err != nil {
		return err
	}

	event, err := w.eventsStore.Get(ctx, eventID)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	project, err := w.projectsStore.Get(ctx, event.ProjectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			event.ProjectID,
		)
	}
	if err = w.substrate.DeleteWorkerAndJobs(ctx, project, event); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q worker and jobs from the substrate",
			eventID,
		)
	}
	return nil
}

// WorkersStore is an interface for components that implement Worker persistence
// concerns.
type WorkersStore interface {
	UpdateStatus(
		ctx context.Context,
		eventID string,
		status WorkerStatus,
	) error
}
