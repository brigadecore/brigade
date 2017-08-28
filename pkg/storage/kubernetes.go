package storage

import (
	"encoding/json"
	"fmt"
	"strings"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/deis/acid/pkg/acid"
)

// store represents a storage engine for a acid.Project.
type store struct {
	client    kubernetes.Interface
	namespace string
}

// Get retrieves the project from storage.
func (s *store) GetProject(id string) (*acid.Project, error) {
	return s.loadProjectConfig(ProjectID(id))
}

// Get creates a new project and writes it to storage.
func (s *store) CreateBuild(build *acid.Build) error {
	shortCommit := build.Commit
	if len(shortCommit) > 8 {
		shortCommit = shortCommit[0:8]
	}

	if build.ID == "" {
		build.ID = genID()
	}

	buildName := fmt.Sprintf("acid-worker-%s-%s", build.ID, shortCommit)

	secret := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: buildName,
			Labels: map[string]string{
				"build":     build.ID,
				"commit":    build.Commit,
				"component": "build",
				"heritage":  "acid",
				"project":   build.ProjectID,
			},
		},
		Data: map[string][]byte{
			"script":  build.Script,
			"payload": build.Payload,
		},
		StringData: map[string]string{
			"project_id":     build.ProjectID,
			"event_type":     build.Type,
			"event_provider": build.Provider,
			"commit":         build.Commit,
			"build_id":       buildName,
		},
	}

	_, err := s.client.CoreV1().Secrets(s.namespace).Create(&secret)
	return err
}

func (s *store) GetBuild(id string) (*acid.Build, error) {
	build := &acid.Build{ID: id}

	labels := labels.Set{"heritage": "acid", "component": "build", "build": build.ID}
	listOption := meta.ListOptions{LabelSelector: labels.AsSelector().String()}
	secrets, err := s.client.CoreV1().Secrets(s.namespace).List(listOption)
	if err != nil {
		return build, err
	}
	if len(secrets.Items) < 1 {
		return nil, fmt.Errorf("could not find build %s: no secrets exist with labels %s", id, labels.AsSelector().String())
	}
	// select the first secret as the build IDs are unique
	buildSecret := secrets.Items[0]

	build.Commit = buildSecret.Labels["commit"]
	build.ProjectID = buildSecret.Labels["project"]
	build.Type = buildSecret.StringData["event_type"]
	build.Provider = buildSecret.StringData["event_provider"]
	build.Payload = buildSecret.Data["payload"]
	build.Script = buildSecret.Data["script"]

	return build, nil
}

// loadProjectConfig loads a project config from inside of Kubernetes.
//
// The namespace is the namespace where the secret is stored.
func (s *store) loadProjectConfig(id string) (*acid.Project, error) {
	proj := &acid.Project{ID: id}

	// The project config is stored in a secret.
	secret, err := s.client.CoreV1().Secrets(s.namespace).Get(id, meta.GetOptions{})
	if err != nil {
		return proj, err
	}

	proj.Name = secret.Annotations["projectName"]

	return proj, configureProject(proj, secret.Data, s.namespace)
}

func configureProject(proj *acid.Project, data map[string][]byte, namespace string) error {
	proj.SharedSecret = def(data["sharedSecret"], "")
	proj.Github.Token = string(data["github.token"])

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
