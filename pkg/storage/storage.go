package storage

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/deis/acid/pkg/acid"
	"github.com/oklog/ulid"
)

// Store represents a storage engine for a Project.
type Store interface {
	// Get retrieves the project from storage.
	GetProject(id, namespace string) (*acid.Project, error)
	// CreateJobSpec creates a new job for the work queue.
	CreateJobSpec(jobSpec *acid.JobSpec, proj *acid.Project) error
}

// New initializes a new storage backend.
func New(c kubernetes.Interface) Store {
	return &store{c}
}

// projectID will encode a project name.
func projectID(id string) string {
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
