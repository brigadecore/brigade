package kube

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/brigade/pkg/brigade"
)

const secretTypeProject = "brigade.sh/project"

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

// SecretFromProject takes a project and converts it to a Kubernetes Secret.
func SecretFromProject(project *brigade.Project) (v1.Secret, error) {
	if project.Name == "" {
		return v1.Secret{}, errors.New("project name is required")
	}

	if project.ID == "" {
		project.ID = brigade.ProjectID(project.Name)
	}

	// The marshal on SecretsMap redacts secrts, so we cast and marshal as a raw
	// map[string]string
	var secrets map[string]string = project.Secrets
	secretsJSON, err := json.Marshal(secrets)
	if err != nil {
		return v1.Secret{}, err
	}

	bfmt := func(b bool) string { return fmt.Sprintf("%t", b) }

	secret := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: project.ID,
			Labels: map[string]string{
				"app":       "brigade",
				"heritage":  "brigade",
				"component": "project",
			},
			Annotations: map[string]string{
				"projectName": project.Name,
			},
		},
		Type: secretTypeProject,
		StringData: map[string]string{
			"sharedSecret":     project.SharedSecret,
			"github.token":     project.Github.Token,
			"github.baseURL":   project.Github.BaseURL,
			"github.uploadURL": project.Github.UploadURL,

			"vcsSidecar":        project.Kubernetes.VCSSidecar,
			"namespace":         project.Kubernetes.Namespace,
			"buildStorageSize":  project.Kubernetes.BuildStorageSize,
			"defaultScript":     project.DefaultScript,
			"defaultScriptName": project.DefaultScriptName,

			"repository": project.Repo.Name,
			"sshKey":     project.Repo.SSHKey,
			"cloneURL":   project.Repo.CloneURL,

			"secrets": string(secretsJSON),

			"worker.registry":   project.Worker.Registry,
			"worker.name":       project.Worker.Name,
			"worker.tag":        project.Worker.Tag,
			"worker.pullPolicy": project.Worker.PullPolicy,

			// These exist in the chart, but not in the brigade.Project
			"initGitSubmodules":   bfmt(project.InitGitSubmodules),
			"imagePullSecrets":    project.ImagePullSecrets,
			"allowPrivilegedJobs": bfmt(project.AllowPrivilegedJobs),
			"allowHostMounts":     bfmt(project.AllowHostMounts),
			"workerCommand":       project.WorkerCommand,

			"kubernetes.cacheStorageClass": project.Kubernetes.CacheStorageClass,
			"kubernetes.buildStorageClass": project.Kubernetes.BuildStorageClass,
		},
	}
	return secret, nil
}

// CreateProject stores a given project.
//
// Project Name is a required field. If not present, Project ID will be calculated
// from project name. This is preferred.
//
// Note that project secrets are not redacted.
func (s *store) CreateProject(project *brigade.Project) error {
	secret, err := SecretFromProject(project)
	if err != nil {
		return err
	}
	_, err = s.client.CoreV1().Secrets(s.namespace).Create(&secret)
	return err
}

// DeleteProject deletes a project from storage.
func (s *store) DeleteProject(id string) error {
	return s.client.CoreV1().Secrets(s.namespace).Delete(id, &meta.DeleteOptions{})
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

	proj.Kubernetes.VCSSidecar = sv.String("vcsSidecar")
	proj.Kubernetes.Namespace = def(sv.String("namespace"), namespace)
	proj.Kubernetes.BuildStorageSize = def(sv.String("buildStorageSize"), "50Mi")
	proj.Kubernetes.BuildStorageClass = sv.String("kubernetes.buildStorageClass")
	proj.Kubernetes.CacheStorageClass = sv.String("kubernetes.cacheStorageClass")

	proj.DefaultScript = sv.String("defaultScript")
	proj.DefaultScriptName = sv.String("defaultScriptName")

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

	// git submodules and host mounts are false by default. Priv jobs are true by default.
	proj.InitGitSubmodules = strings.ToLower(def(sv.String("initGitSubmodules"), "false")) == "true"
	proj.AllowPrivilegedJobs = strings.ToLower(def(sv.String("allowPrivilegedJobs"), "true")) == "true"
	proj.AllowHostMounts = strings.ToLower(def(sv.String("allowHostMounts"), "false")) == "true"

	proj.WorkerCommand = sv.String("workerCommand")
	return proj, nil
}

func def(a, b string) string {
	if len(a) == 0 {
		return b
	}
	return a
}
