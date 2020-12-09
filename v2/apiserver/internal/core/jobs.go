package core

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
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
	// JobPhaseSchedulingFailed represents the state wherein a job was not
	// scheduled due to some unexpected and unrecoverable error encountered by the
	// scheduler.
	JobPhaseSchedulingFailed WorkerPhase = "SCHEDULING_FAILED"
	// JobPhaseStarting represents the state wherein a Job is starting on the
	// substrate but isn't running yet.
	JobPhaseStarting JobPhase = "STARTING"
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
	// Create, given an Event identifier and Job, creates a new Job and schedules
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
	// Cleanup removes Job-related resources from the substrate, presumably
	// upon completion, without deleting the Job from the data store.
	Cleanup(ctx context.Context, eventID, jobName string) error
}

type jobsService struct {
	projectsStore ProjectsStore
	eventsStore   EventsStore
	jobsStore     JobsStore
	substrate     Substrate
}

// NewJobsService returns a specialized interface for managing Jobs.
func NewJobsService(
	projectsStore ProjectsStore,
	eventsStore EventsStore,
	jobsStore JobsStore,
	substrate Substrate,
) JobsService {
	return &jobsService{
		projectsStore: projectsStore,
		eventsStore:   eventsStore,
		jobsStore:     jobsStore,
		substrate:     substrate,
	}
}

func (j *jobsService) Create(
	ctx context.Context,
	eventID string,
	jobName string,
	job Job,
) error {
	event, err := j.eventsStore.Get(ctx, eventID)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	if _, ok := event.Worker.Jobs[jobName]; ok {
		return &meta.ErrConflict{
			Type: "Job",
			ID:   jobName,
			Reason: fmt.Sprintf(
				"Event %q already has a job named %q.",
				eventID,
				jobName,
			),
		}
	}

	// Perform some validations...

	// Determine if ANY of the job's containers:
	//   1. Use shared workspace
	//   2. Run in privileged mode
	//   3. Mount the host's Docker socket
	var useWorkspace = job.Spec.PrimaryContainer.UseWorkspace
	var usePrivileged = job.Spec.PrimaryContainer.Privileged
	var useDockerSocket = job.Spec.PrimaryContainer.UseHostDockerSocket
	for _, sidecarContainer := range job.Spec.SidecarContainers {
		if sidecarContainer.UseWorkspace {
			useWorkspace = true
		}
		if sidecarContainer.Privileged {
			usePrivileged = true
		}
		if sidecarContainer.UseHostDockerSocket {
			useDockerSocket = true
		}
	}

	// Fail quickly if any job is trying to run privileged or use the host's
	// Docker socket, but isn't allowed to per worker configuration.
	if usePrivileged &&
		(event.Worker.Spec.JobPolicies == nil ||
			!event.Worker.Spec.JobPolicies.AllowPrivileged) {
		return &meta.ErrAuthorization{
			Reason: "Worker configuration forbids jobs from utilizing privileged " +
				"containers.",
		}
	}
	if useDockerSocket &&
		(event.Worker.Spec.JobPolicies == nil ||
			!event.Worker.Spec.JobPolicies.AllowDockerSocketMount) {
		return &meta.ErrAuthorization{
			Reason: "Worker configuration forbids jobs from mounting the Docker " +
				"socket.",
		}
	}

	// Fail quickly if the job needs to use shared workspace, but the worker
	// doesn't have any shared workspace.
	if useWorkspace && !event.Worker.Spec.UseWorkspace {
		return &meta.ErrConflict{
			Reason: "The job requested access to the shared workspace, but Worker " +
				"configuration has not enabled this feature.",
		}
	}

	// Set the initial status
	job.Status = &JobStatus{
		Phase: JobPhasePending,
	}

	project, err := j.projectsStore.Get(ctx, event.ProjectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			event.ProjectID,
		)
	}

	if err = j.jobsStore.Create(ctx, eventID, jobName, job); err != nil {
		return errors.Wrapf(
			err, "error saving event %q job %q in store",
			eventID,
			eventID,
		)
	}

	if err = j.substrate.ScheduleJob(ctx, project, event, jobName); err != nil {
		return errors.Wrapf(
			err,
			"error scheduling event %q job %q on the substrate",
			event.ID,
			jobName,
		)
	}

	return nil
}

func (j *jobsService) Start(
	ctx context.Context,
	eventID string,
	jobName string,
) error {
	event, err := j.eventsStore.Get(ctx, eventID)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	job, ok := event.Worker.Jobs[jobName]
	if !ok {
		return &meta.ErrNotFound{
			Type: "Job",
			ID:   jobName,
		}
	}

	if job.Status.Phase != JobPhasePending {
		return &meta.ErrConflict{
			Type: "Job",
			ID:   jobName,
			Reason: fmt.Sprintf(
				"Event %q job %q has already been started.",
				eventID,
				jobName,
			),
		}
	}

	project, err := j.projectsStore.Get(ctx, event.ProjectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			event.ProjectID,
		)
	}

	if err = j.jobsStore.UpdateStatus(
		ctx,
		eventID,
		jobName,
		JobStatus{
			Phase: JobPhaseStarting,
		},
	); err != nil {
		return errors.Wrapf(
			err,
			"error updating status of event %q worker job %q in store",
			eventID,
			jobName,
		)
	}

	if err = j.substrate.StartJob(ctx, project, event, jobName); err != nil {
		return errors.Wrapf(
			err,
			"error starting event %q job %q",
			event.ID,
			jobName,
		)
	}

	return nil
}

func (j *jobsService) GetStatus(
	ctx context.Context,
	eventID string,
	jobName string,
) (JobStatus, error) {
	event, err := j.eventsStore.Get(ctx, eventID)
	if err != nil {
		return JobStatus{},
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	job, ok := event.Worker.Jobs[jobName]
	if !ok {
		return JobStatus{}, &meta.ErrNotFound{
			Type: "Job",
			ID:   jobName,
		}
	}
	return *job.Status, nil
}

func (j *jobsService) WatchStatus(
	ctx context.Context,
	eventID string,
	jobName string,
) (<-chan JobStatus, error) {
	// Read the event and job up front to confirm they both exists.
	event, err := j.eventsStore.Get(ctx, eventID)
	if err != nil {
		return nil,
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	if _, ok := event.Worker.Jobs[jobName]; !ok {
		return nil, &meta.ErrNotFound{
			Type: "Job",
			ID:   jobName,
		}
	}
	statusCh := make(chan JobStatus)
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
			event, err := j.eventsStore.Get(ctx, eventID)
			if err != nil {
				log.Printf("error retrieving event %q from store: %s", eventID, err)
				return
			}
			select {
			case statusCh <- *event.Worker.Jobs[jobName].Status:
			case <-ctx.Done():
				return
			}
		}
	}()
	return statusCh, nil
}

func (j *jobsService) UpdateStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status JobStatus,
) error {
	if err := j.jobsStore.UpdateStatus(
		ctx,
		eventID,
		jobName,
		status,
	); err != nil {
		return errors.Wrapf(
			err,
			"error updating status of event %q worker job %q in store",
			eventID,
			jobName,
		)
	}
	return nil
}

func (j *jobsService) Cleanup(
	ctx context.Context,
	eventID string,
	jobName string,
) error {
	event, err := j.eventsStore.Get(ctx, eventID)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	_, ok := event.Worker.Jobs[jobName]
	if !ok {
		return &meta.ErrNotFound{
			Type: "Job",
			ID:   jobName,
		}
	}
	project, err := j.projectsStore.Get(ctx, event.ProjectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			event.ProjectID,
		)
	}
	if err = j.substrate.DeleteJob(ctx, project, event, jobName); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q jobs %q from the substrate",
			eventID,
			jobName,
		)
	}
	return nil
}

// JobsStore is an interface for components that implement Job persistence
// concerns.
type JobsStore interface {
	// Create persists a new Job for the specified Event in the underlying data
	// store.
	Create(
		ctx context.Context,
		eventID string,
		jobName string,
		job Job,
	) error
	// UpdateStatus updates the status of the specified Job in the underlying data
	// store. If the specified job is not found, implementations MUST return a
	// *meta.ErrNotFound error.
	UpdateStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status JobStatus,
	) error
}
