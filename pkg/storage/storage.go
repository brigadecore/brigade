package storage

import (
	"crypto/sha256"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/oklog/ulid"
	"k8s.io/client-go/kubernetes"

	"github.com/deis/acid/pkg/acid"
)

// Store represents a storage engine for a Project.
type Store interface {
	// GetProject retrieves the project from storage.
	GetProject(id string) (*acid.Project, error)
	// GetBuild retrieves the build from storage.
	GetBuild(id string) (*acid.Build, error)
	// CreateBuild creates a new job for the work queue.
	CreateBuild(build *acid.Build) error
	// GetBuildJobs retrieves all build jobs (pods) from storage.
	GetBuildJobs(build *acid.Build) ([]*acid.Job, error)
	// GetJob retrieves the job from storage.
	GetJob(id string) (*acid.Job, error)
	// GetJobLog retrieves all logs for a job from storage.
	GetJobLog(job *acid.Job) (string, error)
	// GetJobLogStream retrieve a stream of all logs for a job from storage.
	GetJobLogStream(job *acid.Job) (io.ReadCloser, error)
}

// New initializes a new storage backend.
func New(c kubernetes.Interface, namespace string) Store {
	return &store{c, namespace}
}

// ProjectID will encode a project name.
func ProjectID(id string) string {
	if strings.HasPrefix(id, "acid-") {
		return id
	}
	return "acid-" + shortSHA(id)
}

// shortSHA returns a 32-char SHA256 digest as a string.
func shortSHA(input string) string {
	sum := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", sum)[0:54]
}

var entropy = rand.New(rand.NewSource(time.Now().UnixNano()))

func genID() string {
	return strings.ToLower(
		ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String(),
	)
}
