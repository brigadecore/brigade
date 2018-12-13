package controller

import (
	"strings"
	"testing"

	"k8s.io/api/core/v1"
)

func TestNewWorkerPod_Defaults(t *testing.T) {
	build := &v1.Secret{}
	proj := &v1.Secret{
		Data: map[string][]byte{
			"vcsSidecar": []byte("my-vcs-sidecar"),
		},
	}
	config := &Config{
		Namespace: v1.NamespaceDefault,
	}

	pod := NewWorkerPod(build, proj, config)

	spec := pod.Spec
	if spec.NodeSelector["beta.kubernetes.io/os"] != "linux" {
		t.Error("expected linux node selector")
	}
	sidecarVolumeExists := false
	for _, volume := range spec.Volumes {
		if volume.Name == "vcs-sidecar" {
			sidecarVolumeExists = true
		}
	}
	if !sidecarVolumeExists {
		t.Error("expected vcs-sidecar volume to exist")
	}
	if len(spec.InitContainers) == 0 {
		t.Error("expected spec.InitContainers to be non-zero")
	}

	container := spec.Containers[0]
	if container.Name != "brigade-runner" {
		t.Error("expected brigade-runner container name")
	}

	sidecarVolumeMountExists := false
	for _, volumeMount := range container.VolumeMounts {
		if volumeMount.Name == "vcs-sidecar" {
			sidecarVolumeMountExists = true
		}
	}
	if !sidecarVolumeMountExists {
		t.Error("Expected vcs-sidecar volume mount to exist")
	}

	if cmd := strings.Join(container.Command, " "); cmd != "yarn -s start" {
		t.Errorf("Unexpected command: %s", cmd)
	}

	if len(container.Resources.Limits) != 0 {
		t.Errorf("Limits should be undefined")
	}

	if len(container.Resources.Requests) != 0 {
		t.Errorf("Requests should be undefined")
	}
}

func TestNewWorkerPod_NoSidecar(t *testing.T) {
	build := &v1.Secret{}
	proj := &v1.Secret{
		Data: map[string][]byte{
			"vcsSidecar": []byte(""),
		},
	}
	config := &Config{
		Namespace: v1.NamespaceDefault,
	}

	pod := NewWorkerPod(build, proj, config)

	spec := pod.Spec
	sidecarVolumeExists := false
	for _, volume := range spec.Volumes {
		if volume.Name == "vcs-sidecar" {
			sidecarVolumeExists = true
		}
	}
	if sidecarVolumeExists {
		t.Error("expected vcs-sidecar volume not to exist")
	}
	if len(spec.InitContainers) > 0 {
		t.Error("expected spec.InitContainers to be empty")
	}

	container := spec.Containers[0]

	sidecarVolumeMountExists := false
	for _, volumeMount := range container.VolumeMounts {
		if volumeMount.Name == "vcs-sidecar" {
			sidecarVolumeMountExists = true
		}
	}
	if sidecarVolumeMountExists {
		t.Error("expected vcs-sidecar volume mount not to exist")
	}
}
