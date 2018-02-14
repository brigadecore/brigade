package brigade

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
)

// Project describes a Brigade project
//
// This is an internal representation of a project, and contains data that
// should not be made available to the JavaScript runtime.
type Project struct {
	// ID is the computed name of the project (brigade-aeff2343a3234ff)
	ID string `json:"id"`
	// Name is the human readable name of project.
	Name string `json:"name"`
	// Repo describes the repository where the source code is stored.
	Repo Repo `json:"repo"`
	// DefaultScript is a snippet of js used by default when the Repo misses brigade.js in it
	DefaultScript string `json:"defaultScript"`
	// Kubernetes holds information about Kubernetes
	Kubernetes Kubernetes `json:"kubernetes"`
	// SharedSecret is the GitHub shared key
	SharedSecret string `json:"-"`
	// Github holds information about Github.
	Github Github `json:"github"`
	// Secrets is environment variables for brigade.js
	Secrets SecretsMap `json:"secrets"`
	// Worker holds a set of project-specific worker settings which takes precedence over brigade-wide settings
	Worker WorkerConfig `json:"worker"`
}

// SecretsMap is a map[string]string for storing secrets.
//
// When secrets are marshaled, values will be redacted.
type SecretsMap map[string]string

const redacted = "REDACTED"

// MarshalJSON redacts secret values when encoding to JSON.
func (s SecretsMap) MarshalJSON() ([]byte, error) {
	dest := make(map[string]string, len(s))
	for k := range s {
		dest[k] = redacted
	}
	return json.Marshal(dest)
}

// ProjectID will encode a project name.
func ProjectID(id string) string {
	if strings.HasPrefix(id, "brigade-") {
		return id
	}
	return "brigade-" + shortSHA(id)
}

// shortSHA returns a 32-char SHA256 digest as a string.
func shortSHA(input string) string {
	sum := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", sum)[0:54]
}

// Github describes the Github configuration for a project.
type Github struct {
	// Token is used for oauth2 for client interactions.
	Token string `json:"-"`
	// BaseURL is used to construct an Enterprise GitHub client.
	// If not supplied, the assumption is that we are connecting to
	// github.com.
	BaseURL string `json:"baseURL"`
	// UploadURL is the upload URL to be used for GitHub enterprise.
	// Typically, it is the same as the BaseURL.
	UploadURL string `json:"uploadURL"`
}

// Repo describes a Git repository.
type Repo struct {
	// Name of the repository. For GitHub, this is of the form
	// `github.com/org/name` or `github.com/user/name`
	Name string `json:"name"`
	// CloneURL is the URL at which the repository can be cloned
	// Traditionally, this is an HTTPS URL.
	CloneURL string `json:"cloneURL"`
	// SSHKey is the auth string for SSH-based cloning
	SSHKey string `json:"-"`
}

// Kubernetes describes the Kubernetes configuration for a project.
type Kubernetes struct {
	// Namespace is the namespace of this project.
	Namespace string `json:"namespace"`
	// VCSSidecar is the image name/tag for the sidecar that pulls VCS data
	VCSSidecar string `json:"vcsSidecar"`
	// BuildStorageSize is the size of the build shared storage used by the jobs
	BuildStorageSize string `json:"buildStorageSize"`
}
