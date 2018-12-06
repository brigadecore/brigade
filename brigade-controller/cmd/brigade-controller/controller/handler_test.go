package controller

import (
	"strings"
	"testing"

	"k8s.io/api/core/v1"
)

func TestNewWorkerPod_Defaults(t *testing.T) {
	build := &v1.Secret{}
	proj := &v1.Secret{}
	config := &Config{
		Namespace: v1.NamespaceDefault,
	}

	pod := NewWorkerPod(build, proj, config)

	spec := pod.Spec
	if spec.NodeSelector["beta.kubernetes.io/os"] != "linux" {
		t.Error("expected linux node selector")
	}

	container := spec.Containers[0]
	if container.Name != "brigade-runner" {
		t.Error("expected brigade-runner container name")
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
