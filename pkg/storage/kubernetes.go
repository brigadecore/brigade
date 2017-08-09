package storage

import (
	"encoding/json"
	"strings"

	"github.com/deis/acid/pkg/acid"
	"github.com/deis/acid/pkg/k8s"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// store represents a storage engine for a acid.Project.
type store struct{}

// Get retrieves the project from storage.
func (s store) Get(id, namespace string) (*acid.Project, error) {
	return loadProjectConfig(projectID(id), namespace)
}

// loadProjectConfig loads a project config from inside of Kubernetes.
//
// The namespace is the namespace where the secret is stored.
func loadProjectConfig(id, namespace string) (*acid.Project, error) {
	kc, err := k8s.Client()
	proj := &acid.Project{ID: id}
	if err != nil {
		return proj, err
	}

	// The project config is stored in a secret.
	secret, err := kc.CoreV1().Secrets(namespace).Get(id, v1.GetOptions{})
	if err != nil {
		return proj, err
	}

	proj.Name = secret.Name
	proj.Repo.Name = secret.Annotations["projectName"]

	return proj, configureProject(proj, secret.Data, namespace)
}

func configureProject(proj *acid.Project, data map[string][]byte, namespace string) error {
	proj.SharedSecret = def(data["sharedSecret"], "")
	proj.GitHubToken = string(data["githubToken"])

	proj.Kubernetes.Namespace = def(data["namespace"], namespace)
	proj.Kubernetes.VCSSidecar = def(data["vcsSidecar"], DefaultVCSSidecar)

	proj.Repo = acid.Repo{
		Name: def(data["repository"], proj.Name),
		// Note that we have to undo the key escaping.
		SSHKey:   strings.Replace(string(data["sshKey"]), "$", "\n", -1),
		CloneURL: def(data["cloneURL"], ""),
	}

	envVars := map[string]string{}
	if d := data["secrets"]; len(d) > 0 {
		if err := json.Unmarshal(d, &envVars); err != nil {
			return err
		}
	}

	proj.Secrets = envVars
	return nil
}

func def(a []byte, b string) string {
	if len(a) == 0 {
		return b
	}
	return string(a)
}
