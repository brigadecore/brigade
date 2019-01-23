package mock

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
)

var (
	// Now is the date used in all stub date fields.
	Now = time.Now()
	// StubProject is a Project stub.
	StubProject = &brigade.Project{
		ID:           "project-id",
		Name:         "project-name",
		SharedSecret: "shared-secre3t",
		Secrets:      map[string]string{"key": "value"},
	}
	// StubWorker1 is a stub Worker. It is used in StubBuild1, too.
	StubWorker1 = &brigade.Worker{
		ID:        "worker-id1",
		BuildID:   "build-id1",
		ProjectID: "project-id",
		StartTime: Now,
		EndTime:   Now,
		ExitCode:  0,
		Status:    brigade.JobSucceeded,
	}
	// StubWorker2 is a stub Worker. It is used in StubBuild2, too.
	StubWorker2 = &brigade.Worker{
		ID:        "worker-id2",
		BuildID:   "build-id2",
		ProjectID: "project-id",
		StartTime: Now.AddDate(0, 0, -1),
		EndTime:   Now,
		ExitCode:  0,
		Status:    brigade.JobSucceeded,
	}
	// StubBuild1 is a stub Build.
	StubBuild1 = &brigade.Build{
		ID:        "build-id1", // do not change this as it's used in LastBuild related tests on Brigade API
		ProjectID: "project-id",
		Revision: &brigade.Revision{
			Commit: "commit1",
		},
		Type:     "type",
		Provider: "provider",
		Payload:  []byte("payload"),
		Script:   []byte("script"),
		Worker:   StubWorker1,
	}
	// StubBuild2 is another stub Build.
	StubBuild2 = &brigade.Build{
		ID:        "build-id2",
		ProjectID: "project-id",
		Revision: &brigade.Revision{
			Commit: "commit2",
		},
		Type:     "type",
		Provider: "provider",
		Payload:  []byte("payload"),
		Script:   []byte("script"),
		Worker:   StubWorker2,
	}
	// StubJob is a stub Job.
	StubJob = &brigade.Job{
		ID:           "job-id",
		Name:         "job-name",
		Image:        "image",
		CreationTime: Now,
		StartTime:    Now,
		EndTime:      Now,
		ExitCode:     0,
		Status:       brigade.JobSucceeded,
	}
	// StubLogData is string data representing a log.
	StubLogData = "Hello World"
)

// New returns a new Store with the default stubs.
func New() *Store {
	return &Store{
		ProjectList: []*brigade.Project{StubProject},
		Workers:     []*brigade.Worker{StubWorker1, StubWorker2},
		Builds:      []*brigade.Build{StubBuild1, StubBuild2},
		Job:         StubJob,
		LogData:     StubLogData,
	}
}

// Store implements the storage.Storage interface, but returns mock data.
type Store struct {
	// Builds is a slice of Builds.
	Builds []*brigade.Build
	// Job is the job you want returned.
	Job *brigade.Job
	// Workers is a slice of workers.
	Workers []*brigade.Worker
	// LogData is the log data you want returned.
	LogData string
	// ProjectList on this mock
	ProjectList []*brigade.Project
}

// GetProjects gets the mock project wrapped as a slice of projects.
func (s *Store) GetProjects() ([]*brigade.Project, error) {
	return s.ProjectList, nil
}

// CreateProject adds a project to the internal mock
func (s *Store) CreateProject(p *brigade.Project) error {
	s.ProjectList = append(s.ProjectList, p)
	return nil
}

// ReplaceProject replaces a project in the internal mock
func (s *Store) ReplaceProject(p *brigade.Project) error {
	found := false
	for _, pr := range s.ProjectList {
		if pr.Name == p.Name {
			pr = p
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("Project with ID %s was not found", p.ID)
	}

	return nil
}

// DeleteProject deletes a project from the internal mock
func (s *Store) DeleteProject(id string) error {
	tmp := []*brigade.Project{}
	for _, p := range s.ProjectList {
		if p.ID == id {
			tmp = append(tmp, p)
		}
	}
	s.ProjectList = tmp
	return nil
}

// GetProject returns the Project
func (s *Store) GetProject(id string) (*brigade.Project, error) {
	for _, proj := range s.ProjectList {
		if proj.ID == id {
			return proj, nil
		}
	}
	return nil, fmt.Errorf("mock project not found for %s", id)
}

// GetProjectBuilds returns the mock Build wrapped in a slice.
func (s *Store) GetProjectBuilds(p *brigade.Project) ([]*brigade.Build, error) {
	return s.GetBuilds()
}

// GetBuilds returns the mock build wrapped in a slice.
func (s *Store) GetBuilds() ([]*brigade.Build, error) {
	return s.Builds, nil
}

// GetBuild gets the first mock Build.
func (s *Store) GetBuild(id string) (*brigade.Build, error) {
	return s.Builds[0], nil
}

// GetBuildJobs gets the mock job wrapped in a slice.
func (s *Store) GetBuildJobs(b *brigade.Build) ([]*brigade.Job, error) {
	return []*brigade.Job{s.Job}, nil
}

// GetWorker gets the first mock worker.
func (s *Store) GetWorker(bid string) (*brigade.Worker, error) {
	return s.Workers[0], nil
}

// GetJob gets the mock job.
func (s *Store) GetJob(id string) (*brigade.Job, error) {
	return s.Job, nil
}

// GetJobLog gets the mock log data.
func (s *Store) GetJobLog(j *brigade.Job) (string, error) {
	return s.LogData, nil
}

// GetJobLogStream gets the mock log data as a readcloser.
func (s *Store) GetJobLogStream(j *brigade.Job) (io.ReadCloser, error) {
	return rc(s.LogData), nil
}

// GetJobLogStreamFollow gets the mock log data as a readcloser.
func (s *Store) GetJobLogStreamFollow(j *brigade.Job) (io.ReadCloser, error) {
	return s.GetJobLogStream(j)
}

// GetWorkerLog gets the mock log data.
func (s *Store) GetWorkerLog(w *brigade.Worker) (string, error) {
	return s.LogData, nil
}

// GetWorkerLogStream gets a readcloser of the mock log data.
func (s *Store) GetWorkerLogStream(w *brigade.Worker) (io.ReadCloser, error) {
	return rc(s.LogData), nil
}

// GetWorkerLogStreamFollow gets a readcloser of the mock log data.
func (s *Store) GetWorkerLogStreamFollow(w *brigade.Worker) (io.ReadCloser, error) {
	return s.GetWorkerLogStream(w)
}

// CreateBuild fakes a new build.
func (s *Store) CreateBuild(b *brigade.Build) error {
	s.Builds[0] = b
	return nil
}

// GetStorageClassNames returns the names of the StorageClass instances in the cluster
func (s *Store) GetStorageClassNames() ([]string, error) {
	return []string{}, nil
}

// DeleteBuild fakes a build deletion.
func (s *Store) DeleteBuild(bid string, options storage.DeleteBuildOptions) error {
	return nil
}

// rc wraps a string in a ReadCloser.
func rc(s string) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBufferString(s))
}
