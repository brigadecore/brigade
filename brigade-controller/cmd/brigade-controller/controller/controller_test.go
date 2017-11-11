package controller

import (
	"testing"
	"time"

	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
)

func TestController(t *testing.T) {
	createdPod := false
	client := fake.NewSimpleClientset()
	client.PrependReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createdPod = true
		t.Log("creating pod")
		return false, nil, nil
	})
	config := &Config{
		Namespace:        v1.NamespaceDefault,
		WorkerImage:      "deis/brigade-worker:latest",
		WorkerPullPolicy: string(v1.PullIfNotPresent),
	}
	controller := NewController(client, config)

	secret := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      "moby",
			Namespace: v1.NamespaceDefault,
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "build",
				"project":   "ahab",
				"build":     "queequeg",
			},
		},
	}

	sidecarImage := "fake/sidecar:latest"
	project := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      "ahab",
			Namespace: v1.NamespaceDefault,
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "project",
			},
		},
		// This and the missing 'script' will trigger an initContainer
		Data: map[string][]byte{
			"vcsSidecar": []byte(sidecarImage),
		},
	}

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	client.CoreV1().Secrets(v1.NamespaceDefault).Create(&secret)
	client.CoreV1().Secrets(v1.NamespaceDefault).Create(&project)

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
	if c.Name != "brigade-runner" {
		t.Error("Container.Name is not correct")
	}
	if envlen := len(c.Env); envlen != 7 {
		t.Errorf("expected 7 items in Container.Env, got %d", envlen)
	}
	if c.Image != config.WorkerImage {
		t.Error("Container.Image is not correct")
	}
	if c.VolumeMounts[0].Name != volumeName {
		t.Error("Container.VolumeMounts is not correct")
	}

	if l := len(pod.Spec.InitContainers); l != 1 {
		t.Fatalf("Expected 1 init container, got %d", l)
	}
	ic := pod.Spec.InitContainers[0]
	if envlen := len(ic.Env); envlen != 5 {
		t.Errorf("expected 5 env vars, got %d", envlen)
	}

	if ic.Image != sidecarImage {
		t.Errorf("expected sidecar %q, got %q", sidecarImage, ic.Image)
	}

	if ic.VolumeMounts[0].Name != sidecarVolumeName {
		t.Errorf("expected sidecar volume %q, got %q", sidecarVolumeName, ic.VolumeMounts[0].Name)
	}
}

func TestController_WithScript(t *testing.T) {
	// t.Skip("make better")

	createdPod := false
	client := fake.NewSimpleClientset()
	client.PrependReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createdPod = true
		t.Log("creating pod")
		return false, nil, nil
	})

	config := &Config{
		Namespace:        v1.NamespaceDefault,
		WorkerImage:      "deis/brgiade-worker:latest",
		WorkerPullPolicy: string(v1.PullIfNotPresent),
	}
	controller := NewController(client, config)

	secret := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      "moby",
			Namespace: v1.NamespaceDefault,
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "build",
				"project":   "ahab",
				"build":     "queequeg",
			},
		},
		Data: map[string][]byte{
			"script": []byte("hello"),
		},
	}

	sidecarImage := "fake/sidecar:latest"
	project := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      "ahab",
			Namespace: v1.NamespaceDefault,
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "project",
			},
		},
		// This and the missing 'script' will trigger an initContainer
		Data: map[string][]byte{
			"vcsSidecar": []byte(sidecarImage),
		},
	}

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	client.CoreV1().Secrets(v1.NamespaceDefault).Create(&secret)
	client.CoreV1().Secrets(v1.NamespaceDefault).Create(&project)

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
	if c.Name != "brigade-runner" {
		t.Error("Container.Name is not correct")
	}
	if envlen := len(c.Env); envlen != 7 {
		t.Errorf("expected 7 items in Container.Env, got %d", envlen)
	}
	if c.Image != config.WorkerImage {
		t.Error("Container.Image is not correct")
	}
	if c.VolumeMounts[0].Name != volumeName {
		t.Error("Container.VolumeMounts is not correct")
	}

	if l := len(pod.Spec.InitContainers); l != 0 {
		t.Fatalf("Expected no init container, got %d", l)
	}
}

func TestController_NoSidecar(t *testing.T) {
	createdPod := false
	client := fake.NewSimpleClientset()
	client.PrependReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createdPod = true
		t.Log("creating pod")
		return false, nil, nil
	})

	config := &Config{
		Namespace:        v1.NamespaceDefault,
		WorkerImage:      "deis/brgiade-worker:latest",
		WorkerPullPolicy: string(v1.PullIfNotPresent),
	}
	controller := NewController(client, config)

	secret := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      "moby",
			Namespace: v1.NamespaceDefault,
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "build",
				"project":   "ahab",
				"build":     "queequeg",
			},
		},
	}

	project := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      "ahab",
			Namespace: v1.NamespaceDefault,
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "project",
			},
		},
	}

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	client.CoreV1().Secrets(v1.NamespaceDefault).Create(&secret)
	client.CoreV1().Secrets(v1.NamespaceDefault).Create(&project)

	// Let's wait for the controller to create the pod
	wait.Poll(100*time.Millisecond, wait.ForeverTestTimeout, func() (bool, error) {
		return createdPod, nil
	})

	pod, err := client.CoreV1().Pods(v1.NamespaceDefault).Get(secret.Name, meta.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := pod.Spec.Containers[0]
	if envlen := len(c.Env); envlen != 7 {
		t.Errorf("expected 7 items in Container.Env, got %d", envlen)
	}
	if c.Image != config.WorkerImage {
		t.Error("Container.Image is not correct")
	}
	if l := len(pod.Spec.InitContainers); l != 0 {
		t.Fatalf("Expected no init container, got %d", l)
	}
}
