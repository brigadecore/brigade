package core

import (
	"context"
	"time"
)

// JobPhase represents where a Job is within its lifecycle.
type JobPhase string

const (
	// JobPhaseAborted represents the state wherein a Job was forcefully
	// stopped during execution.
	JobPhaseAborted JobPhase = "ABORTED"
	// JobPhaseFailed represents the state wherein a Job has run to
	// completion but experienced errors.
	JobPhaseFailed JobPhase = "FAILED"
	// JobPhasePending represents the state wherein a Job is awaiting
	// execution.
	JobPhasePending JobPhase = "PENDING"
	// JobPhaseRunning represents the state wherein a Job is currently
	// being executed.
	JobPhaseRunning JobPhase = "RUNNING"
	// JobPhaseSucceeded represents the state where a Job has run to
	// completion without error.
	JobPhaseSucceeded JobPhase = "SUCCEEDED"
	// JobPhaseTimedOut represents the state wherein a Job has has not completed
	// within a designated timeframe.
	JobPhaseTimedOut JobPhase = "TIMED_OUT"
	// JobPhaseUnknown represents the state wherein a Job's state is unknown. Note
	// that this is possible if and only if the underlying Job execution substrate
	// (Kubernetes), for some unanticipated, reason does not know the Job's
	// (Pod's) state.
	JobPhaseUnknown JobPhase = "UNKNOWN"
)

// Job represents a component spawned by a Worker to complete a single task
// in the course of handling an Event.
type Job struct {
	// Spec is the technical blueprint for the Job.
	Spec JobSpec `json:"spec" bson:"spec"`
	// Status contains details of the Job's current state.
	Status *JobStatus `json:"status" bson:"status"`
}

// JobSpec is the technical blueprint for a Job.
type JobSpec struct {
	// PrimaryContainer specifies the details of an OCI container that forms the
	// cornerstone of the Job. Job success or failure is tied to completion and
	// exit code of this container.
	PrimaryContainer JobContainerSpec `json:"primaryContainer" bson:"primaryContainer"` // nolint: lll
	// SidecarContainers specifies the details of supplemental, "sidecar"
	// containers. Their completion and exit code do not directly impact Job
	// status. Brigade does not understand dependencies between a Job's multiple
	// containers and cannot enforce any specific startup or shutdown order. When
	// such dependencies exist (for instance, a primary container than cannot
	// proceed with a suite of tests until a database is launched and READY in a
	// sidecar container), then logic within those containers must account for
	// these constraints.
	SidecarContainers map[string]JobContainerSpec `json:"sidecarContainers,omitempty" bson:"sidecarContainers,omitempty"` // nolint: lll
	// TimeoutSeconds specifies the time, in seconds, that must elapse before a
	// running Job should be considered to have timed out.
	TimeoutSeconds int64 `json:"timeoutSeconds,omitempty" bson:"timeoutSeconds,omitempty"` // nolint: lll
	// Host specifies criteria for selecting a suitable host (substrate node) for
	// the Job. This is useful in cases where a Job requires a specific,
	// non-default operating system (i.e. Windows) or specific hardware (e.g. a
	// GPU.)
	Host *JobHost `json:"host,omitempty" bson:"host,omitempty"`
}

// JobContainerSpec amends the ContainerSpec type with additional Job-specific
// fields.
type JobContainerSpec struct {
	// ContainerSpec encapsulates generic specifications for an OCI container.
	ContainerSpec `json:",inline" bson:",inline"`
	// UseWorkspace indicates whether the Job requires the Worker's shared
	// workspace (if one exists) to be mounted into the OCI container.
	UseWorkspace bool `json:"useWorkspace" bson:"useWorkspace"`
	// WorkspaceMountPath specifies the path in the OCI container's file system
	// where, if applicable, the Worker's shared workspace should be mounted.
	WorkspaceMountPath string `json:"workspaceMountPath,omitempty" bson:"workspaceMountPath,omitempty"` // nolint: lll
	// UseSource indicates whether the Job requires source code to be retrieved
	// from a git repository and mounted into the OCI container. Note this
	// requires git configuration to have been specified at the Project and/or
	// Event levels.
	UseSource bool `json:"useSource" bson:"useSource"`
	// SourceMountPath specifies the path in the OCI container's file system
	// where, if applicable, source code retrieved from a git repository should be
	// mounted.
	SourceMountPath string `json:"sourceMountPath,omitempty" bson:"sourceMountPath,omitempty"` // nolint: lll
	// Privileged indicates whether the OCI container should operate in a
	// "privileged" (relaxed permissions) mode. This is commonly used to effect
	// "Docker-in-Docker" ("DinD") scenarios wherein one of a Job's OCI containers
	// must run its own Docker daemon. Note this field REQUESTS privileged status
	// for the container, but that may be disallowed by Project-level
	// configuration.
	Privileged bool `json:"privileged" bson:"privileged"`
	// UseHostDockerSocket indicates whether the OCI container should mount the
	// host's Docker socket into its own file system. This is commonly used to
	// effect "Docker-out-of-Docker" ("DooD") scenarios wherein one of a Job's OCI
	// containers must utilize the host's Docker daemon. GENERALLY, THIS IS HIGHLY
	// DISCOURAGED. Note this field REQUESTS to mount the host's Docker socket
	// into the container, but that may be disallowed by Project-level
	// configuration.
	UseHostDockerSocket bool `json:"useHostDockerSocket" bson:"useHostDockerSocket"` // nolint: lll
}

// JobHost represents criteria for selecting a suitable host (substrate node)
// for a Job.
type JobHost struct {
	// OS specifies which "family" of operating system is required on a substrate
	// node to host a Job. Valid values are "linux" and "windows". When empty,
	// Brigade assumes "linux".
	OS string `json:"os,omitempty" bson:"os,omitempty"`
	// NodeSelector specifies labels that must be present on the substrate node to
	// host a Job. This provides an opaque mechanism for communicating Job needs
	// such as specific hardware like an SSD or GPU.
	NodeSelector map[string]string `json:"nodeSelector,omitempty" bson:"nodeSelector,omitempty"` // nolint: lll
}

// JobStatus represents the status of a Job.
type JobStatus struct {
	// Started indicates the time the Job began execution.
	Started *time.Time `json:"started,omitempty" bson:"started,omitempty"`
	// Ended indicates the time the Job concluded execution. It will be nil
	// for a Job that is not done executing.
	Ended *time.Time `json:"ended,omitempty" bson:"ended,omitempty"`
	// Phase indicates where the Job is in its lifecycle.
	Phase JobPhase `json:"phase,omitempty" bson:"phase,omitempty"`
}

// JobsService is the specialized interface for managing Jobs. It's
// decoupled from underlying technology choices (e.g. data store, message bus,
// etc.) to keep business logic reusable and consistent while the underlying
// tech stack remains free to change.
type JobsService interface {
	// Create, given an Event identifier and JobSpec, creates a new Job and starts
	// it on Brigade's workload execution substrate. If the specified Event does
	// not exist, implementations MUST return a *meta.ErrNotFound error. If the
	// specified Event already has a Job with the specified name, implementations
	// MUST return a *meta.ErrConflict error.
	Create(
		ctx context.Context,
		eventID string,
		jobName string,
		job Job,
	) error
	// Start, given an Event identifier and Job name, starts that Job on
	// Brigade's workload execution substrate. If the specified Event or specified
	// Job thereof does not exist, implementations MUST return a *meta.ErrNotFound
	// error.
	Start(
		ctx context.Context,
		eventID string,
		jobName string,
	) error
	// GetStatus, given an Event identifier and Job name, returns the Job's
	// status. If the specified Event or specified Job thereof does not exist,
	// implementations MUST return a *meta.ErrNotFound error.
	GetStatus(
		ctx context.Context,
		eventID string,
		jobName string,
	) (JobStatus, error)
	// WatchStatus, given an Event identifier and Job name, returns a channel over
	// which the Job's status is streamed. The channel receives a new JobStatus
	// every time there is any change in that status. If the specified Event or
	// specified Job thereof does not exist, implementations MUST return a
	// *meta.ErrNotFound error.
	WatchStatus(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan JobStatus, error)
	// UpdateStatus, given an Event identifier and Job name, updates the status of
	// that Job. If the specified Event or specified Job thereof does not exist,
	// implementations MUST return a *meta.ErrNotFound error.
	UpdateStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status JobStatus,
	) error
}
