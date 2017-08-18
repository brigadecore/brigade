package controller

import (
	"testing"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	core "k8s.io/client-go/testing"
)

func TestController(t *testing.T) {
	// t.Skip("make better")

	createdPod := false
	client := fake.NewSimpleClientset()
	client.PrependReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createdPod = true
		t.Log("creating pod")
		return false, nil, nil
	})

	controller := NewController(client, v1.NamespaceDefault)

	secret := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      "moby",
			Namespace: v1.NamespaceDefault,
			Labels: map[string]string{
				"heritage":  "acid",
				"managedBy": "acid",
				"role":      "build",
			},
		},
	}

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	client.CoreV1().Secrets(v1.NamespaceDefault).Create(&secret)

	// Let's wait for the controller to create the pod
	wait.Poll(100*time.Millisecond, wait.ForeverTestTimeout, func() (bool, error) {
		return createdPod, nil
	})

	pod, err := client.CoreV1().Pods(v1.NamespaceDefault).Get(secret.Name, meta.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !labels.Equals(pod.GetLabels(), secret.GetLabels()) {
		t.Error("Pod.Lables do not match")
	}

	if pod.Spec.Volumes[0].Name != volumeName {
		t.Error("Spec.Volumes are not correct")
	}

	c := pod.Spec.Containers[0]
	if c.Name != "acid-runner" {
		t.Error("Container.Name is not correct")
	}
	if len(c.Env) != 5 {
		t.Error("expected 5 Container.Env")
	}
	if c.Image != acidWorkerImage {
		t.Error("Container.Image is not correct")
	}
	if c.VolumeMounts[0].Name != volumeName {
		t.Error("Container.VolumeMounts is not correct")
	}
}
