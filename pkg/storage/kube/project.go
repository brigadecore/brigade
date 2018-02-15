package kube

import (
	"encoding/json"
	"strings"

	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/brigade/pkg/brigade"
)

// GetProjects retrieves all projects from storage.
func (s *store) GetProjects() ([]*brigade.Project, error) {
	lo := meta.ListOptions{LabelSelector: "app=brigade,component=project"}
	secretList, err := s.client.CoreV1().Secrets(s.namespace).List(lo)
	if err != nil {
		return nil, err
	}
	projList := make([]*brigade.Project, len(secretList.Items))
	for i := range secretList.Items {
		var err error
		projList[i], err = NewProjectFromSecret(&secretList.Items[i], s.namespace)
		if err != nil {
			return nil, err
		}
	}
	return projList, nil
}

// GetProject retrieves the project from storage.
func (s *store) GetProject(id string) (*brigade.Project, error) {
	return s.loadProjectConfig(brigade.ProjectID(id))
}

// loadProjectConfig loads a project config from inside of Kubernetes.
//
// The namespace is the namespace where the secret is stored.
func (s *store) loadProjectConfig(id string) (*brigade.Project, error) {
	// The project config is stored in a secret.
	secret, err := s.client.CoreV1().Secrets(s.namespace).Get(id, meta.GetOptions{})
	if err != nil {
		return nil, err
	}

	return NewProjectFromSecret(secret, s.namespace)
}

// NewProjectFromSecret creates a new project from a secret.
func NewProjectFromSecret(secret *v1.Secret, namespace string) (*brigade.Project, error) {
	sv := SecretValues(secret.Data)
	proj := new(brigade.Project)
	proj.ID = secret.ObjectMeta.Name
	proj.Name = secret.Annotations["projectName"]

	proj.SharedSecret = sv.String("sharedSecret")
	proj.Github.Token = sv.String("github.token")
	proj.Github.BaseURL = sv.String("github.baseURL")
	proj.Github.UploadURL = sv.String("github.uploadURL")

	proj.Kubernetes.Namespace = def(sv.String("namespace"), namespace)
	proj.Kubernetes.VCSSidecar = sv.String("vcsSidecar")
	proj.Kubernetes.BuildStorageSize = def(sv.String("buildStorageSize"), "50Mi")
	proj.DefaultScript = sv.String("defaultScript")

	proj.Repo = brigade.Repo{
		Name: def(sv.String("repository"), proj.Name),
		// Note that we have to undo the key escaping.
		SSHKey:   strings.Replace(sv.String("sshKey"), "$", "\n", -1),
		CloneURL: sv.String("cloneURL"),
	}

	envVars := map[string]string{}
	if d := sv.Bytes("secrets"); len(d) > 0 {
		if err := json.Unmarshal(d, &envVars); err != nil {
			return nil, err
		}
	}
	proj.Secrets = envVars

	proj.Worker = brigade.WorkerConfig{
		Registry:   sv.String("worker.registry"),
		Name:       sv.String("worker.name"),
		Tag:        sv.String("worker.tag"),
		PullPolicy: sv.String("worker.pullPolicy"),
	}
	return proj, nil
}

func def(a, b string) string {
	if len(a) == 0 {
		return b
	}
	return a
}
