package kube

import (
	"encoding/json"
	"fmt"
	"strings"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/deis/brigade/pkg/brigade"
)

// GetProjects retrieves all projects from storage.
func (s *store) GetProjects() ([]*brigade.Project, error) {
	lo := meta.ListOptions{LabelSelector: fmt.Sprintf("app=brigade,component=project")}
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

func NewProjectFromSecret(secret *v1.Secret, namespace string) (*brigade.Project, error) {
	proj := new(brigade.Project)
	proj.ID = secret.ObjectMeta.Name
	proj.Name = secret.Annotations["projectName"]
	proj.SharedSecret = def(secret.Data["sharedSecret"], "")
	proj.Github.Token = string(secret.Data["github.token"])

	proj.Kubernetes.Namespace = def(secret.Data["namespace"], namespace)
	proj.Kubernetes.VCSSidecar = def(secret.Data["vcsSidecar"], brigade.DefaultVCSSidecar)

	proj.Repo = brigade.Repo{
		Name: def(secret.Data["repository"], proj.Name),
		// Note that we have to undo the key escaping.
		SSHKey:   strings.Replace(string(secret.Data["sshKey"]), "$", "\n", -1),
		CloneURL: def(secret.Data["cloneURL"], ""),
	}

	envVars := map[string]string{}
	if d := secret.Data["secrets"]; len(d) > 0 {
		if err := json.Unmarshal(d, &envVars); err != nil {
			return nil, err
		}
	}

	proj.Secrets = envVars
	return proj, nil
}

func def(a []byte, b string) string {
	if len(a) == 0 {
		return b
	}
	return string(a)
}
