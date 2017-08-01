package storage

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/deis/acid/pkg/acid"
)

const DefaultVCSSidecar = "acidic.azurecr.io/vcs-sidecar:latest"

// Store represents a storage engine for a Project.
type Store interface {
	// Get retrieves the project from storage.
	Get(id, namespace string) (*acid.Project, error)
}

// New initializes a new storage backend.
func New() Store { return new(store) }

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
