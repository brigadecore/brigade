package controller

import (
	"strings"
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewWorkerPod_Defaults(t *testing.T) {
	build := &v1.Secret{}
	proj := &v1.Secret{}
	client := fake.NewSimpleClientset()
	config := &Config{
		Namespace: v1.NamespaceDefault,
	}

	c := NewController(client, config)
	pod := c.newWorkerPod(build, proj)

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

}
