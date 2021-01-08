package core

import (
	"context"
	"encoding/json"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

// SubstrateWorkerCount represents a count of Workers currently executing on
// the substrate.
type SubstrateWorkerCount struct {
	// Count is the cardinality of Workers currently executing on the substrate.
	Count int `json:"count"`
}

// MarshalJSON amends SubstrateWorkerCount instances with type metadata.
func (s SubstrateWorkerCount) MarshalJSON() ([]byte, error) {
	type Alias SubstrateWorkerCount
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "SubstrateWorkerCount",
			},
			Alias: (Alias)(s),
		},
	)
}

// SubstrateJobCount represents a count of Workers currently executing on
// the substrate.
type SubstrateJobCount struct {
	// Count is the cardinality of Jobs currently executing on the substrate.
	Count int `json:"count"`
}

// MarshalJSON amends SubstrateJobCount instances with type metadata.
func (s SubstrateJobCount) MarshalJSON() ([]byte, error) {
	type Alias SubstrateJobCount
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "SubstrateJobCount",
			},
			Alias: (Alias)(s),
		},
	)
}

// SubstrateService is the specialized interface for monitoring the state of the
// substrate.
type SubstrateService interface {
	// CountRunningWorkers returns a count of Workers currently executing on the
	// substrate.
	CountRunningWorkers(context.Context) (SubstrateWorkerCount, error)
	// CountRunningJobs returns a count of Jobs currently executing on the
	// substrate.
	CountRunningJobs(context.Context) (SubstrateJobCount, error)
}

type substrateService struct {
	substrate Substrate
}

// NewSubstrateService returns a specialized interface for managing Projects.
func NewSubstrateService(substrate Substrate) SubstrateService {
	return &substrateService{
		substrate: substrate,
	}
}

func (s *substrateService) CountRunningWorkers(
	ctx context.Context,
) (SubstrateWorkerCount, error) {
	count, err := s.substrate.CountRunningWorkers(ctx)
	if err != nil {
		return count, errors.Wrapf(
			err,
			"error counting running workers on substrate",
		)
	}
	return count, nil
}

func (s *substrateService) CountRunningJobs(
	ctx context.Context,
) (SubstrateJobCount, error) {
	count, err := s.substrate.CountRunningJobs(ctx)
	if err != nil {
		return count, errors.Wrapf(err, "error counting running jobs on substrate")
	}
	return count, nil
}

// Substrate is an interface for components that permit services to coordinate
// with Brigade's underlying workload execution substrate, i.e. Kubernetes.
type Substrate interface {
	// CountRunningWorkers returns a count of Workers currently executing on the
	// substrate.
	CountRunningWorkers(context.Context) (SubstrateWorkerCount, error)
	// CountRunningJobs returns a count of Jobs currently executing on the
	// substrate.
	CountRunningJobs(context.Context) (SubstrateJobCount, error)

	// CreateProject prepares the substrate to host Project workloads. The
	// provided Project argument may be amended with substrate-specific details
	// and returned, so this function should be called prior to a Project being
	// initially persisted so that substrate-specific details will be included.
	CreateProject(context.Context, Project) (Project, error)
	// DeleteProject removes all Project-related resources from the substrate.
	DeleteProject(context.Context, Project) error

	// ScheduleWorker prepares the substrate for the Event's worker and schedules
	// the Worker for async / eventual execution.
	ScheduleWorker(context.Context, Project, Event) error
	// StartWorker starts an Event's Worker on the substrate.
	StartWorker(context.Context, Project, Event) error

	// StoreJobEnvironment securely stores Job environment variables where they
	// are accessible to other substrate operations. This obviates the need to
	// store these potential secrets in the database.
	StoreJobEnvironment(
		ctx context.Context,
		project Project,
		eventID string,
		jobName string,
		jobSpec JobSpec,
	) error
	// ScheduleJob prepares the substrate for a Job and schedules the Job for
	// async / eventual execution.
	ScheduleJob(
		ctx context.Context,
		project Project,
		event Event,
		jobName string,
	) error
	// StartJob starts a Job on the substrate.
	StartJob(
		ctx context.Context,
		project Project,
		event Event,
		jobName string,
	) error

	// DeleteJob deletes all substrate resources pertaining to the specified Job.
	DeleteJob(
		ctx context.Context,
		project Project,
		event Event,
		jobName string,
	) error

	// DeleteWorkerAndJobs deletes all substrate resources pertaining to the
	// specified Event's Worker and Jobs.
	DeleteWorkerAndJobs(context.Context, Project, Event) error
}
