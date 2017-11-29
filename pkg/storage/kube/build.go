package kube

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/oklog/ulid"
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/Azure/brigade/pkg/brigade"
)

// GetBuild returns the build.
func (s *store) GetBuild(id string) (*brigade.Build, error) {
	build := &brigade.Build{ID: id}

	labels := labels.Set{"heritage": "brigade", "component": "build", "build": build.ID}
	listOption := meta.ListOptions{LabelSelector: labels.AsSelector().String()}
	secrets, err := s.client.CoreV1().Secrets(s.namespace).List(listOption)
	if err != nil {
		return nil, err
	}
	if len(secrets.Items) < 1 {
		return nil, fmt.Errorf("could not find build %s: no secrets exist with labels %s", id, labels.AsSelector().String())
	}
	// Select the first secret as the build IDs are unique
	b := NewBuildFromSecret(secrets.Items[0])
	b.Worker, err = s.GetWorker(build.ID)
	return b, err
}

// Get creates a new project and writes it to storage.
func (s *store) CreateBuild(build *brigade.Build) error {
	shortCommit := build.Commit
	if len(shortCommit) > 8 {
		shortCommit = shortCommit[0:8]
	}

	if build.ID == "" {
		build.ID = genID()
	}

	buildName := fmt.Sprintf("brigade-worker-%s-%s", build.ID, shortCommit)

	secret := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: buildName,
			Labels: map[string]string{
				"build":     build.ID,
				"commit":    build.Commit,
				"component": "build",
				"heritage":  "brigade",
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
			"build_name":     buildName,
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
		b.Worker, _ = findWorker(b.ID, podList)
		buildList[i] = b
	}
	return buildList, nil
}

// GetProjectBuilds returns all the builds for the given project.
func (s *store) GetProjectBuilds(proj *brigade.Project) ([]*brigade.Build, error) {
	// Load the pods that ran as part of this build.
	lo := meta.ListOptions{LabelSelector: fmt.Sprintf("heritage=brigade,component=build,project=%s", proj.ID)}

	secretList, err := s.client.CoreV1().Secrets(s.namespace).List(lo)
	if err != nil {
		return nil, err
	}

	// The theory here is that the secrets and pods are close to 1:1, so we can
	// preload the pods and take a local hit in walking the list rather than take
	// a network hit to load each pod per secret.
	podList, err := s.client.CoreV1().Pods(s.namespace).List(lo)
	if err != nil {
		return nil, err
	}

	buildList := make([]*brigade.Build, len(secretList.Items))
	for i := range secretList.Items {
		b := NewBuildFromSecret(secretList.Items[i])
		// The error is ErrWorkerNotFound, and in that case, we just ignore
		// it and assign nil to the worker.
		b.Worker, _ = findWorker(b.ID, podList)
		buildList[i] = b
	}
	return buildList, nil
}

func findWorker(id string, pods *v1.PodList) (*brigade.Worker, bool) {
	for _, i := range pods.Items {
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
	return &brigade.Build{
		ID:        secret.ObjectMeta.Labels["build"],
		ProjectID: secret.ObjectMeta.Labels["project"],
		Type:      string(secret.Data["event_type"]),
		Provider:  string(secret.Data["event_provider"]),
		Commit:    secret.ObjectMeta.Labels["commit"],
		Payload:   secret.Data["payload"],
		Script:    secret.Data["script"],
	}
}

var entropy = rand.New(rand.NewSource(time.Now().UnixNano()))

func genID() string {
	id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	return strings.ToLower(id.String())
}
