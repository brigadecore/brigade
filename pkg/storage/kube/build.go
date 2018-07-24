package kube

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/oklog/ulid"
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/brigade/pkg/brigade"
)

const secretTypeBuild = "brigade.sh/build"

// GetBuild returns the build.
func (s *store) GetBuild(id string) (*brigade.Build, error) {
	build := &brigade.Build{ID: id}

	labels := fmt.Sprint("heritage=brigade,component=build,build=", build.ID)
	listOption := meta.ListOptions{LabelSelector: labels}
	secrets, err := s.client.CoreV1().Secrets(s.namespace).List(listOption)
	if err != nil {
		return nil, err
	}
	if len(secrets.Items) < 1 {
		return nil, fmt.Errorf("could not find build %s: no secrets exist with labels %s", id, labels)
	}
	// Select the first secret as the build IDs are unique
	b := NewBuildFromSecret(secrets.Items[0])
	b.Worker, err = s.GetWorker(build.ID)
	return b, err
}

// Get creates a new project and writes it to storage.
func (s *store) CreateBuild(build *brigade.Build) error {
	if build.ID == "" {
		build.ID = genID()
	}

	buildName := fmt.Sprintf("brigade-worker-%s", build.ID)

	secret := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: buildName,
			Labels: map[string]string{
				"build":     build.ID,
				"component": "build",
				"heritage":  "brigade",
				"project":   build.ProjectID,
			},
		},
		Type: secretTypeBuild,
		Data: map[string][]byte{
			"script":  build.Script,
			"payload": build.Payload,
		},
		StringData: map[string]string{
			"build_id":       buildName,
			"build_name":     buildName,
			"commit_id":      build.Revision.Commit,
			"commit_ref":     build.Revision.Ref,
			"event_provider": build.Provider,
			"event_type":     build.Type,
			"project_id":     build.ProjectID,
			"log_level":      build.LogLevel,
		},
	}

	_, err := s.client.CoreV1().Secrets(s.namespace).Create(&secret)
	return err
}

// GetBuilds returns all the builds in storage.
func (s *store) GetBuilds() ([]*brigade.Build, error) {
	lo := meta.ListOptions{LabelSelector: "heritage=brigade,component=build"}

	secretList, err := s.client.CoreV1().Secrets(s.namespace).List(lo)
	if err != nil {
		return nil, err
	}

	podList, err := s.client.CoreV1().Pods(s.namespace).List(lo)
	if err != nil {
		return nil, err
	}

	buildList := make([]*brigade.Build, len(secretList.Items))
	for i := range secretList.Items {
		b := NewBuildFromSecret(secretList.Items[i])
		// The error is ErrWorkerNotFound, and in that case, we just ignore
		// it and assign nil to the worker.
		b.Worker, _ = findWorker(b.ID, podList.Items)
		buildList[i] = b
	}
	return buildList, nil
}

// GetProjectBuilds returns all the builds for the given project.
func (s *store) GetProjectBuilds(proj *brigade.Project) ([]*brigade.Build, error) {

	// Load the pods that ran as part of this build.
	labelSelectorMap := map[string]string{
		"heritage":  "brigade",
		"component": "build",
		"project":   proj.ID,
	}

	projectSecrets, err := s.apiCache.GetSecretsFilteredBy(labelSelectorMap)
	if err != nil {
		return nil, err
	}

	// The theory here is that the secrets and pods are close to 1:1, so we can
	// preload the pods and take a local hit in walking the list rather than take
	// a network hit to load each pod per secret.
	projectPods, err := s.apiCache.GetPodsFilteredBy(labelSelectorMap)
	if err != nil {
		return nil, err
	}

	buildList := make([]*brigade.Build, len(projectSecrets))
	for i := range projectSecrets {
		b := NewBuildFromSecret(projectSecrets[i])
		// The error is ErrWorkerNotFound, and in that case, we just ignore
		// it and assign nil to the worker.
		b.Worker, _ = findWorker(b.ID, projectPods)
		buildList[i] = b
	}

	return buildList, nil
}

func findWorker(id string, pods []v1.Pod) (*brigade.Worker, bool) {
	for _, i := range pods {
		buildID, ok := i.Labels["build"]
		if !ok {
			continue
		}
		if id == buildID {
			return NewWorkerFromPod(i), true
		}
	}
	return nil, false
}

// NewBuildFromSecret creates a Build object from a secret.
func NewBuildFromSecret(secret v1.Secret) *brigade.Build {
	lbs := secret.ObjectMeta.Labels
	sv := SecretValues(secret.Data)
	return &brigade.Build{
		ID:        lbs["build"],
		ProjectID: lbs["project"],
		Type:      sv.String("event_type"),
		Provider:  sv.String("event_provider"),
		Revision: &brigade.Revision{
			Commit: sv.String("commit_id"),
			Ref:    sv.String("commit_ref"),
		},
		Payload: sv.Bytes("payload"),
		Script:  sv.Bytes("script"),
	}
}

var entropy = rand.New(rand.NewSource(time.Now().UnixNano()))

func genID() string {
	id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	return strings.ToLower(id.String())
}
