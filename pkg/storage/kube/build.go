package kube

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/oklog/ulid"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/deis/acid/pkg/acid"
)

// GetBuild returns the build.
func (s *store) GetBuild(id string) (*acid.Build, error) {
	build := &acid.Build{ID: id}

	labels := labels.Set{"heritage": "acid", "component": "build", "build": build.ID}
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

func (s *store) GetProjectBuilds(proj *acid.Project) ([]*acid.Build, error) {
	// Load the pods that ran as part of this build.
	lo := meta.ListOptions{LabelSelector: fmt.Sprintf("heritage=acid,component=build,project=%s", proj.ID)}

	secretList, err := s.client.CoreV1().Secrets(s.namespace).List(lo)
	if err != nil {
		return nil, err
	}
	buildList := make([]*acid.Build, len(secretList.Items))
	for i := range secretList.Items {
		buildList[i] = NewBuildFromSecret(secretList.Items[i])
	}
	return buildList, nil
}

func NewBuildFromSecret(secret v1.Secret) *acid.Build {
	return &acid.Build{
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
