package core

import (
	"context"
	"fmt"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

// LogLevel represents the desired granularity of Worker log output.
type LogLevel string

// LogLevelInfo represents INFO level granularity in Worker log output.
const LogLevelInfo LogLevel = "INFO"

// WorkerPhase represents where a Worker is within its lifecycle.
type WorkerPhase string

const (
	// WorkerPhaseAborted represents the state wherein a worker was forcefully
	// stopped during execution.
	WorkerPhaseAborted WorkerPhase = "ABORTED"
	// WorkerPhaseCanceled represents the state wherein a pending worker was
	// canceled prior to execution.
	WorkerPhaseCanceled WorkerPhase = "CANCELED"
	// WorkerPhaseFailed represents the state wherein a worker has run to
	// completion but experienced errors.
	WorkerPhaseFailed WorkerPhase = "FAILED"
	// WorkerPhasePending represents the state wherein a worker is awaiting
	// execution.
	WorkerPhasePending WorkerPhase = "PENDING"
	// WorkerPhaseRunning represents the state wherein a worker is currently
	// being executed.
	WorkerPhaseRunning WorkerPhase = "RUNNING"
	// WorkerPhaseSchedulingFailed represents the state wherein a worker was not
	// scheduled due to some unexpected and unrecoverable error encountered by the
	// scheduler.
	WorkerPhaseSchedulingFailed WorkerPhase = "SCHEDULING_FAILED"
	// WorkerPhaseSucceeded represents the state where a worker has run to
	// completion without error.
	WorkerPhaseSucceeded WorkerPhase = "SUCCEEDED"
	// WorkerPhaseTimedOut represents the state wherein a worker has has not
	// completed within a designated timeframe.
	WorkerPhaseTimedOut WorkerPhase = "TIMED_OUT"
	// WorkerPhaseUnknown represents the state wherein a worker's state is
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
	Jobs map[string]Job `json:"jobs,omitempty" bson:"jobs,omitempty"`
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
	// Commit specifies a commit (by SHA) to be checked out.
	Commit string `json:"commit,omitempty" bson:"commit,omitempty"`
	// Ref specifies a tag or branch to be checked out. If left blank, this will
	// default to "master" at runtime.
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
}

type workersService struct {
	projectsStore ProjectsStore
	eventsStore   EventsStore
	substrate     Substrate
}

// NewWorkersService returns a specialized interface for managing Workers.
func NewWorkersService(
	projectsStore ProjectsStore,
	eventsStore EventsStore,
	substrate Substrate,
) WorkersService {
	return &workersService{
		projectsStore: projectsStore,
		eventsStore:   eventsStore,
		substrate:     substrate,
	}
}

func (w *workersService) Start(ctx context.Context, eventID string) error {
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

	// TODO: We should consider changing the Worker's phase here so that if the
	// observer is down, the Worker doesn't continue to appear in a pending state.
	// The scheduler uses at least once delivery semantics. If a Worker continued
	// to appear in a pending state despite having been started, the possibility
	// exists that the scheduler could try to start the same Worker more than
	// once.

	if err = w.substrate.StartWorker(ctx, project, event); err != nil {
		return errors.Wrapf(err, "error starting worker for event %q", event.ID)
	}
	return nil
}
