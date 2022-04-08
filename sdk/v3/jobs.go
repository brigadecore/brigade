package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	rm "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v3/meta"
	"github.com/brigadecore/brigade/sdk/v3/restmachinery"
)

// JobKind represents the canonical Job kind string
const JobKind = "Job"

// OSFamily represents a type of operating system.
type OSFamily string

const (
	// OSFamilyLinux represents a Linux-based OS.
	OSFamilyLinux OSFamily = "linux"
	// OSFamilyWindows represents a Windows-based OS.
	OSFamilyWindows OSFamily = "windows"
)

// JobPhase represents where a Job is within its lifecycle.
type JobPhase string

const (
	// JobPhaseAborted represents the state wherein a Job was forcefully
	// stopped during execution.
	JobPhaseAborted JobPhase = "ABORTED"
	// JobPhaseCanceled represents the state wherein a pending Job was
	// canceled prior to execution.
	JobPhaseCanceled JobPhase = "CANCELED"
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
	JobPhaseSchedulingFailed JobPhase = "SCHEDULING_FAILED"
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
	// (Kubernetes), for some unanticipated reason, does not know the Job's
	// (Pod's) state.
	JobPhaseUnknown JobPhase = "UNKNOWN"
)

// IsTerminal returns a bool indicating whether the JobPhase is terminal.
func (j JobPhase) IsTerminal() bool {
	switch j {
	case JobPhaseAborted:
		fallthrough
	case JobPhaseCanceled:
		fallthrough
	case JobPhaseFailed:
		fallthrough
	case JobPhaseSchedulingFailed:
		fallthrough
	case JobPhaseSucceeded:
		fallthrough
	case JobPhaseTimedOut:
		return true
	}
	return false
}

// Job represents a component spawned by a Worker to complete a single task
// in the course of handling an Event.
type Job struct {
	// Name is the Job's name. It should be unique among a given Worker's Jobs.
	Name string `json:"name"`
	// Created indicates the time at which a Job was created. This is recorded by
	// the system. Clients must leave the value of this field set to nil when
	// using the API to create a Job.
	Created *time.Time `json:"created,omitempty"`
	// Spec is the technical blueprint for the Job.
	Spec JobSpec `json:"spec"`
	// Status contains details of the Job's current state.
	Status *JobStatus `json:"status,omitempty"`
}

// MarshalJSON amends Job instances with type metadata so that clients do not
// need to be concerned with the tedium of doing so.
func (j Job) MarshalJSON() ([]byte, error) {
	type Alias Job
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       JobKind,
			},
			Alias: (Alias)(j),
		},
	)
}

// JobSpec is the technical blueprint for a Job.
type JobSpec struct {
	// PrimaryContainer specifies the details of an OCI container that forms the
	// cornerstone of the Job. Job success or failure is tied to completion and
	// exit code of this container.
	PrimaryContainer JobContainerSpec `json:"primaryContainer"`
	// SidecarContainers specifies the details of supplemental, "sidecar"
	// containers. Their completion and exit code do not directly impact Job
	// status. Brigade does not understand dependencies between a Job's multiple
	// containers and cannot enforce any specific startup or shutdown order. When
	// such dependencies exist (for instance, a primary container than cannot
	// proceed with a suite of tests until a database is launched and READY in a
	// sidecar container), then logic within those containers must account for
	// these constraints.
	SidecarContainers map[string]JobContainerSpec `json:"sidecarContainers,omitempty"` // nolint: lll
	// TimeoutDuration specifies the time duration that must elapse before a
	// running Job should be considered to have timed out. This duration string
	// is a sequence of decimal numbers, each with optional fraction and a unit
	// suffix, such as "300ms", "3.14s" or "2h45m".
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	TimeoutDuration string `json:"timeoutDuration,omitempty"`
	// Host specifies criteria for selecting a suitable host (substrate node) for
	// the Job. This is useful in cases where a Job requires a specific,
	// non-default operating system (i.e. Windows) or specific hardware (e.g. a
	// GPU.)
	Host *JobHost `json:"host,omitempty"`
	// Fallible specifies whether the job is permitted to fail WITHOUT causing the
	// worker process to fail. The API server does not use this field directly,
	// but it is information that may be valuable to gateways that report job
	// success/failure upstream to original event sources.
	//
	// Note that omitempty keeps this compatible with older API servers (whose
	// schema-based validation will reject the unknown field) as long as it's not
	// set to true.
	Fallible bool `json:"fallible,omitempty"`
}

// JobContainerSpec amends the ContainerSpec type with additional Job-specific
// fields.
type JobContainerSpec struct {
	// ContainerSpec encapsulates generic specifications for an OCI container.
	ContainerSpec `json:",inline"`
	// WorkingDirectory specifies the OCI container's working directory.
	WorkingDirectory string `json:"workingDirectory,omitempty"`
	// WorkspaceMountPath specifies the path in the OCI container's file system
	// where, if applicable, the Worker's shared workspace should be mounted. If
	// left blank, the Job implicitly does not use the Worker's shared workspace.
	WorkspaceMountPath string `json:"workspaceMountPath,omitempty"`
	// SourceMountPath specifies the path in the OCI container's file system
	// where, if applicable, source code retrieved from a VCS repository should be
	// mounted. If left blank, the Job implicitly does not use source code
	// retrieved from a VCS repository.
	SourceMountPath string `json:"sourceMountPath,omitempty"`
	// Privileged indicates whether the OCI container should operate in a
	// "privileged" (relaxed permissions) mode. This is commonly used to effect
	// "Docker-in-Docker" ("DinD") scenarios wherein one of a Job's OCI containers
	// must run its own Docker daemon. Note this field REQUESTS privileged status
	// for the container, but that may be disallowed by Project-level
	// configuration.
	Privileged bool `json:"privileged"`
	// UseHostDockerSocket indicates whether the OCI container should mount the
	// host's Docker socket into its own file system. This is commonly used to
	// effect "Docker-out-of-Docker" ("DooD") scenarios wherein one of a Job's OCI
	// containers must utilize the host's Docker daemon. GENERALLY, THIS IS HIGHLY
	// DISCOURAGED. Note this field REQUESTS to mount the host's Docker socket
	// into the container, but that may be disallowed by Project-level
	// configuration.
	//
	// Note: This is being removed for the 2.0.0 release because of security
	// issues AND declining usefulness. (Many Kubernetes distros now use
	// containerd instead of Docker.) This can be put back in the future if the
	// need is proven AND if it can be done safely.
	//
	// For more details, see https://github.com/brigadecore/brigade/issues/1666
	//
	// UseHostDockerSocket bool `json:"useHostDockerSocket"`
}

// JobHost represents criteria for selecting a suitable host (substrate node)
// for a Job.
type JobHost struct {
	// OS specifies which "family" of operating system is required on a substrate
	// node to host a Job. Valid values are "linux" and "windows". When empty,
	// Brigade assumes "linux".
	OS OSFamily `json:"os,omitempty"`
	// NodeSelector specifies labels that must be present on the substrate node to
	// host a Job. This provides an opaque mechanism for communicating Job needs
	// such as specific hardware like an SSD or GPU.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// JobStatus represents the status of a Job.
type JobStatus struct {
	// Started indicates the time the Job began execution.
	Started *time.Time `json:"started,omitempty"`
	// Ended indicates the time the Job concluded execution. It will be nil
	// for a Job that is not done executing.
	Ended *time.Time `json:"ended,omitempty"`
	// Phase indicates where the Job is in its lifecycle.
	Phase JobPhase `json:"phase,omitempty"`
}

// MarshalJSON amends JobStatus instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
func (j JobStatus) MarshalJSON() ([]byte, error) {
	type Alias JobStatus
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "JobStatus",
			},
			Alias: (Alias)(j),
		},
	)
}

// JobCreateOptions represents useful, optional settings for creating a new
// Job. It currently has no fields, but exists to preserve the possibility of
// future expansion without having to change client function signatures.
type JobCreateOptions struct{}

// JobStartOptions represents useful, optional settings for starting a Job. It
// currently has no fields, but exists to preserve the possibility of future
// expansion without having to change client function signatures.
type JobStartOptions struct{}

// JobStatusGetOptions represents useful, optional criteria for retrieving a
// Job's status. It currently has no fields, but exists to preserve the
// possibility of future expansion without having to change client function
// signatures.
type JobStatusGetOptions struct{}

// JobStatusWatchOptions represents useful, optional criteria for establishing a
// stream of a Job's Status. It currently has no fields, but exists to preserve
// the possibility of future expansion without having to change client function
// signatures.
type JobStatusWatchOptions struct{}

// JobStatusUpdateOptions represents useful, optional settings for updating a
// Job's status. It currently has no fields, but exists to preserve the
// possibility of future expansion without having to change client function
// signatures.
type JobStatusUpdateOptions struct{}

// JobCleanupOptions represents useful, optional settings for cleaning up after
// a Job. It currently has no fields, but exists to preserve the possibility of
// future expansion without having to change client function signatures.
type JobCleanupOptions struct{}

// JobTimeoutOptions represents useful, optional settings for timing out a Job.
// It currently has no fields, but exists to preserve the possibility of future
// expansion without having to change client function signatures.
type JobTimeoutOptions struct{}

// JobsClient is the specialized client for managing Event Jobs with the
// Brigade API.
type JobsClient interface {
	// Create, given an Event identifier and Job, creates a new pending Job
	// and schedules it for execution.
	Create(
		ctx context.Context,
		eventID string,
		job Job,
		opts *JobCreateOptions,
	) error
	// Start initiates execution of a pending Job.
	Start(
		ctx context.Context,
		eventID string,
		jobName string,
		opts *JobStartOptions,
	) error
	// GetStatus, given an Event identifier and Job name, returns the Job's
	// status.
	GetStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		opts *JobStatusGetOptions,
	) (JobStatus, error)
	// WatchStatus, given an Event identifier and Job name, returns a channel
	// over which the Job's status is streamed. The channel receives a new
	// JobStatus every time there is any change in that status.
	WatchStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		opts *JobStatusWatchOptions,
	) (<-chan JobStatus, <-chan error, error)
	// UpdateStatus, given an Event identifier and Job name, updates the status
	// of that Job.
	UpdateStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status JobStatus,
		opts *JobStatusUpdateOptions,
	) error
	Cleanup(
		ctx context.Context,
		eventID,
		jobName string,
		opts *JobCleanupOptions,
	) error
	// Timeout, given an Event identifier and Job name, executes timeout logic
	// for a Job that has exceeded its timeout limit.
	Timeout(
		ctx context.Context,
		eventID,
		jobName string,
		opts *JobTimeoutOptions,
	) error
}

type jobsClient struct {
	*rm.BaseClient
}

// NewJobsClient returns a specialized client for managing Event Jobs.
func NewJobsClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) JobsClient {
	return &jobsClient{
		BaseClient: rm.NewBaseClient(apiAddress, apiToken, opts),
	}
}

func (j *jobsClient) Create(
	ctx context.Context,
	eventID string,
	job Job,
	_ *JobCreateOptions,
) error {
	return j.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPost,
			Path:        fmt.Sprintf("v2/events/%s/worker/jobs", eventID),
			ReqBodyObj:  job,
			SuccessCode: http.StatusCreated,
		},
	)
}

func (j *jobsClient) Start(
	ctx context.Context,
	eventID string,
	jobName string,
	_ *JobStartOptions,
) error {
	return j.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method: http.MethodPut,
			Path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/start",
				eventID,
				jobName,
			),
			SuccessCode: http.StatusOK,
		},
	)
}

func (j *jobsClient) GetStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	_ *JobStatusGetOptions,
) (JobStatus, error) {
	status := JobStatus{}
	return status, j.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method: http.MethodGet,
			Path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/status",
				eventID,
				jobName,
			),
			SuccessCode: http.StatusOK,
			RespObj:     &status,
		},
	)
}

func (j *jobsClient) WatchStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	_ *JobStatusWatchOptions,
) (<-chan JobStatus, <-chan error, error) {
	resp, err := j.SubmitRequest( // nolint: bodyclose
		ctx,
		rm.OutboundRequest{
			Method: http.MethodGet,
			Path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/status",
				eventID,
				jobName,
			),
			QueryParams: map[string]string{
				"watch": trueStr,
			},
			SuccessCode: http.StatusOK,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	statusCh := make(chan JobStatus)
	errCh := make(chan error)

	// This goroutine will close the response body when it completes
	go j.receiveStatusStream(ctx, resp.Body, statusCh, errCh)

	return statusCh, errCh, nil
}

func (j *jobsClient) UpdateStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status JobStatus,
	_ *JobStatusUpdateOptions,
) error {
	return j.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method: http.MethodPut,
			Path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/status",
				eventID,
				jobName,
			),
			ReqBodyObj:  status,
			SuccessCode: http.StatusOK,
		},
	)
}

func (j *jobsClient) Cleanup(
	ctx context.Context,
	eventID,
	jobName string,
	_ *JobCleanupOptions,
) error {
	return j.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method: http.MethodPut,
			Path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/cleanup",
				eventID,
				jobName,
			),
			SuccessCode: http.StatusOK,
		},
	)
}

func (j *jobsClient) Timeout(
	ctx context.Context,
	eventID,
	jobName string,
	_ *JobTimeoutOptions,
) error {
	return j.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method: http.MethodPut,
			Path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/timeout",
				eventID,
				jobName,
			),
			SuccessCode: http.StatusOK,
		},
	)
}

func (j *jobsClient) receiveStatusStream(
	ctx context.Context,
	reader io.ReadCloser,
	statusCh chan<- JobStatus,
	errCh chan<- error,
) {
	defer reader.Close()
	decoder := json.NewDecoder(reader)
	for {
		status := JobStatus{}
		if err := decoder.Decode(&status); err != nil {
			select {
			case errCh <- err:
			case <-ctx.Done():
			}
			return
		}
		select {
		case statusCh <- status:
		case <-ctx.Done():
			return
		}
	}
}
