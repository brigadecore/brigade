package kube

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"time"

	"github.com/Azure/brigade/pkg/brigade"
)

func TestNewBuildFromSecret(t *testing.T) {
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"build":   stubBuildID,
				"project": stubProjectID,
			},
		},
		Data: map[string][]byte{
			"event_type":     []byte("foo"),
			"event_provider": []byte("bar"),
			"payload":        []byte("this is a payload"),
			"script":         []byte("ohai"),
			"commit_id":      []byte("abc123"),
			"commit_ref":     []byte("refs/heads/master"),
			"log_level":      []byte("LOG"),
		},
	}
	build := NewBuildFromSecret(secret)
	if !reflect.DeepEqual(build, stubBuild) {
		t.Errorf("build differs from expected build, got '%v', expected '%v'", build, stubBuild)
	}
}

func TestCreateBuild(t *testing.T) {
	k, s := fakeStore()
	if err := s.CreateBuild(stubBuild); err != nil {
		t.Fatal(err)
	}

	secrets, _ := k.CoreV1().Secrets("default").List(metav1.ListOptions{})
	if len(secrets.Items) != 1 {
		t.Fatalf("Build was not stored as secret")
	}
}

func TestGetBuild(t *testing.T) {
	k, s := fakeStore()
	createFakeWorker(k, stubWorkerPod)
	if err := s.CreateBuild(stubBuild); err != nil {
		t.Fatal(err)
	}

	b, err := s.GetBuild(stubBuild.ID)
	if err != nil {
		t.Fatal(err)
	}

	// The one that comes back will have a worker attached, because we created
	// one.
	if b.Worker == nil {
		t.Error("expected a worker")
	}

	if b.Worker.ID != stubWorkerPod.Name {
		t.Errorf("expected worker name %s, got %s", stubWorkerPod.Name, b.Worker.ID)
	}

	if b.Worker.ProjectID != stubWorkerPod.Labels["project"] {
		t.Errorf("expected project ID %s got %s", stubWorkerPod.Labels["project"], b.ProjectID)
	}

	if b.ProjectID != stubProjectID {
		t.Errorf("expected build project ID %s, got %s", stubProjectID, b.ProjectID)
	}
}

func TestGetBuilds(t *testing.T) {
	k, s := fakeStore()
	createFakeWorker(k, stubWorkerPod)
	if err := s.CreateBuild(stubBuild); err != nil {
		t.Fatal(err)
	}

	// Deliberately set this to an out-of-date ULID
	secondBuildID := genID()
	secondBuild := &brigade.Build{
		ID:        secondBuildID,
		ProjectID: stubProjectID,
		Type:      "second",
		Provider:  "mock",
		Revision:  &brigade.Revision{Ref: "heads/refs/master"},
	}
	if err := s.CreateBuild(secondBuild); err != nil {
		t.Fatal(err)
	}
	w2 := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-id-" + secondBuildID,
			Labels: map[string]string{
				"build":     secondBuildID,
				"project":   stubProjectID,
				"component": "build",
				"heritage":  "brigade",
			},
			CreationTimestamp: podStartTime,
		},
		Status: v1.PodStatus{
			Phase:     v1.PodSucceeded,
			StartTime: &podStartTime,
			ContainerStatuses: []v1.ContainerStatus{
				{
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							ExitCode:   0,
							FinishedAt: podEndTime,
						},
					},
				},
			},
		},
	}
	createFakeWorker(k, w2)

	builds, err := s.GetBuilds()
	if err != nil {
		t.Fatal(err)
	}

	if l := len(builds); l != 2 {
		t.Fatalf("expected 2 builds, got %d", l)
	}
}

func TestGetProjectBuilds(t *testing.T) {

	// this could be related
	// https://github.com/kubernetes/client-go/issues/352

	k, s := fakeStore()
	createFakeWorker(k, stubWorkerPod)
	if err := s.CreateBuild(stubBuild); err != nil {
		t.Fatal(err)
	}

	// Deliberately set this to an out-of-date ULID
	secondBuildID := genID()
	secondBuild := &brigade.Build{
		ID:        secondBuildID,
		ProjectID: stubProjectID,
		Type:      "second",
		Provider:  "mock",
		Revision:  &brigade.Revision{Ref: "heads/refs/master"},
	}
	if err := s.CreateBuild(secondBuild); err != nil {
		t.Fatal(err)
	}
	w2 := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-id-" + secondBuildID,
			Labels: map[string]string{
				"build":     secondBuildID,
				"project":   stubProjectID,
				"component": "build",
				"heritage":  "brigade",
			},
			CreationTimestamp: podStartTime,
		},
		Status: v1.PodStatus{
			Phase:     v1.PodSucceeded,
			StartTime: &podStartTime,
			ContainerStatuses: []v1.ContainerStatus{
				{
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							ExitCode:   0,
							FinishedAt: podEndTime,
						},
					},
				},
			},
		},
	}
	createFakeWorker(k, w2)

	proj := &brigade.Project{
		ID: stubProjectID,
	}

	// wait until api cache is synced
	s.BlockUntilAPICacheSynced(time.After(time.Second))

	builds, err := s.GetProjectBuilds(proj)
	if err != nil {
		t.Fatal(err)
	}

	if l := len(builds); l != 2 {
		t.Fatalf("expected 2 builds, got %d", l)
	}
}
