package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Azure/brigade/pkg/brigade"
)

func TestNewRouter(t *testing.T) {
	s := &mockStore{
		project: &brigade.Project{
			Name: "pequod/stubbs",
		},
	}
	r := newRouter(s)

	if r == nil {
		t.Fail()
	}

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("Unexpected status on healthz: %s", res.Status)
	}

	body, err := ioutil.ReadFile("./testdata/dockerhub-push.json")
	if err != nil {
		t.Fatal(err)
	}

	// Basically, we're testing to make sure the route exists, but having it bail
	// before it hits the GitHub API.
	res, err = http.Post(ts.URL+"/events/webhook/deis/empty-testbed/master", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 400 {
		t.Fatalf("Expected bad status, got: %s", res.Status)
	}

}

type mockStore struct {
	project *brigade.Project
}

// GetProjects retrieves all projects from storage.
func (m *mockStore) GetProjects() ([]*brigade.Project, error) {
	return []*brigade.Project{m.project}, nil
}

// GetProject retrieves the project from storage.
func (m *mockStore) GetProject(id string) (*brigade.Project, error) {
	return m.project, nil
}

// GetProjectBuilds retrieves the project's builds from storage.
func (m *mockStore) GetProjectBuilds(proj *brigade.Project) ([]*brigade.Build, error) {
	return []*brigade.Build{}, nil
}

// GetBuilds retrieves all active builds from storage.
func (m *mockStore) GetBuilds() ([]*brigade.Build, error) {
	return []*brigade.Build{}, nil
}

// GetBuild retrieves the build from storage.
func (m *mockStore) GetBuild(id string) (*brigade.Build, error) {
	return &brigade.Build{}, nil
}

// CreateBuild creates a new job for the work queue.
func (m *mockStore) CreateBuild(build *brigade.Build) error {
	return nil
}

// GetBuildJobs retrieves all build jobs (pods) from storage.
func (m *mockStore) GetBuildJobs(build *brigade.Build) ([]*brigade.Job, error) {
	return []*brigade.Job{}, nil
}

// GetWorker returns the worker for a given build.
func (m *mockStore) GetWorker(buildID string) (*brigade.Worker, error) {
	return &brigade.Worker{}, nil
}

// GetJob retrieves the job from storage.
func (m *mockStore) GetJob(id string) (*brigade.Job, error) {
	return &brigade.Job{}, nil
}

// GetJobLog retrieves all logs for a job from storage.
func (m *mockStore) GetJobLog(job *brigade.Job) (string, error) {
	return "log", nil
}

// GetJobLogStream retrieve a stream of all logs for a job from storage.
func (m *mockStore) GetJobLogStream(job *brigade.Job) (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewBufferString("log")), nil
}
