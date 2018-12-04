package controller

import (
	"strings"
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

	emptyQuantity := resource.Quantity{}

	if quantity := container.Resources.Limits.Cpu(); *quantity != emptyQuantity {
		t.Errorf("Unexpected cpu limits quantity: %s", quantity.String())
	}

	if quantity := container.Resources.Limits.Memory(); *quantity != emptyQuantity {
		t.Errorf("Unexpected memory limits quantity: %s", quantity.String())
	}

	if quantity := container.Resources.Requests.Cpu(); *quantity != emptyQuantity {
		t.Errorf("Unexpected cpu requests quantity: %s", quantity.String())
	}

	if quantity := container.Resources.Requests.Memory(); *quantity != emptyQuantity {
		t.Errorf("Unexpected memory requests quantity: %s", quantity.String())
	}

}
