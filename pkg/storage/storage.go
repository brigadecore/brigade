package storage

import (
	"io"

	"github.com/Azure/brigade/pkg/brigade"
)

// ProjectStore represents storage for projects.
type ProjectStore interface {
	// GetProjects retrieves all projects from storage.
	GetProjects() ([]*brigade.Project, error)
	// GetProject retrieves the project from storage.
	GetProject(id string) (*brigade.Project, error)
	// GetProjectBuilds retrieves the project's builds from storage.
	GetProjectBuilds(proj *brigade.Project) ([]*brigade.Build, error)
	// CreateProject creates a new project record in storage.
	CreateProject(proj *brigade.Project) error
	// DeleteProject deletes a project from storage.
	DeleteProject(id string) error
}

// Store represents a storage engine for a brigade projects, builds, and jobs.
type Store interface {
	ProjectStore
	// GetBuilds retrieves all active builds from storage.
	GetBuilds() ([]*brigade.Build, error)
	// GetBuild retrieves the build from storage.
	GetBuild(id string) (*brigade.Build, error)
	// CreateBuild creates a new job for the work queue.
	CreateBuild(build *brigade.Build) error
	// GetBuildJobs retrieves all build jobs (pods) from storage.
	GetBuildJobs(build *brigade.Build) ([]*brigade.Job, error)
	// GetWorker returns the worker for a given build.
	GetWorker(buildID string) (*brigade.Worker, error)
	// GetJob retrieves the job from storage.
	GetJob(id string) (*brigade.Job, error)
	// GetJobLog retrieves all logs for a job from storage.
	GetJobLog(job *brigade.Job) (string, error)
	// GetJobLogStream retrieve a stream of all logs for a job from storage.
	GetJobLogStream(job *brigade.Job) (io.ReadCloser, error)
	// GetJobLogStreamFollow retrieve a follow stream of all logs for a job from storage.
	GetJobLogStreamFollow(job *brigade.Job) (io.ReadCloser, error)
	// GetWorkerLog retrieves all logs for a worker from storage.
	GetWorkerLog(job *brigade.Worker) (string, error)
	// GetWorkerLogStream retrieve a stream of all logs for a worker from storage.
	GetWorkerLogStream(job *brigade.Worker) (io.ReadCloser, error)
	// GetWorkerLogStreamFollow retrieve a followed stream of all logs for a worker from storage.
	GetWorkerLogStreamFollow(job *brigade.Worker) (io.ReadCloser, error)
}
