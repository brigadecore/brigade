package storage

import (
	"encoding/json"
	"fmt"
	"strings"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/deis/acid/pkg/acid"
)

// store represents a storage engine for a acid.Project.
type store struct {
	client kubernetes.Interface
}

// Get retrieves the project from storage.
func (s *store) GetProject(id, namespace string) (*acid.Project, error) {
	return s.loadProjectConfig(projectID(id), namespace)
}

// Get retrieves the project from storage.
func (s *store) CreateBuild(build *acid.Build, proj *acid.Project) error {
	return s.createSecret(build, proj)
}

func (s *store) createSecret(b *acid.Build, p *acid.Project) error {
	shortCommit := b.Commit
	if len(shortCommit) > 8 {
		shortCommit = shortCommit[0:8]
	}

	if b.ID == "" {
		b.ID = genID()
	}

	buildName := fmt.Sprintf("acid-worker-%s-%s", b.ID, shortCommit)
	cleanProjName := strings.Replace(p.Repo.Name, "/", "-", -1)

	secret := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: buildName,
			Labels: map[string]string{
				"heritage":  "acid",
				"managedBy": "acid",
				"jobname":   buildName,
				"belongsto": cleanProjName,
				"commit":    b.Commit,
				// Need a different name for this.
				"role": "build",
			},
		},
		Data: map[string][]byte{
			"script":  b.Script,
			"payload": b.Payload,
		},
		StringData: map[string]string{
			"project_id":     p.ID,
			"event_type":     b.Type,
			"event_provider": b.Provider,
			"commit":         b.Commit,
		},
	}

	_, err := s.client.CoreV1().Secrets(p.Kubernetes.Namespace).Create(&secret)
	return err
}

// loadProjectConfig loads a project config from inside of Kubernetes.
//
// The namespace is the namespace where the secret is stored.
func (s *store) loadProjectConfig(id, namespace string) (*acid.Project, error) {
	proj := &acid.Project{ID: id}

	// The project config is stored in a secret.
	secret, err := s.client.CoreV1().Secrets(namespace).Get(id, meta.GetOptions{})
	if err != nil {
		return proj, err
	}

	proj.Name = secret.Annotations["projectName"]

	return proj, configureProject(proj, secret.Data, namespace)
}

func configureProject(proj *acid.Project, data map[string][]byte, namespace string) error {
	proj.SharedSecret = def(data["sharedSecret"], "")
	proj.GitHubToken = string(data["githubToken"])

	proj.Kubernetes.Namespace = def(data["namespace"], namespace)
	proj.Kubernetes.VCSSidecar = def(data["vcsSidecar"], acid.DefaultVCSSidecar)

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
