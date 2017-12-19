package vacuum

import (
	"testing"
	"time"

	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestRun_Age(t *testing.T) {
	client := setupFakeClient()

	secrets, err := client.CoreV1().Secrets(v1.NamespaceDefault).List(meta.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(secrets.Items) != 3 {
		t.Fatalf("expected 3 secrets, got %d", len(secrets.Items))
	}
	pods, err := client.CoreV1().Pods(v1.NamespaceDefault).List(meta.ListOptions{})
	if err != nil {
		t.Fatal("no pods returned")
	}
	if len(pods.Items) != 3 {
		t.Fatalf("expected 3 pods, got %d", len(pods.Items))
	}

	num, err := New(time.Now(), 0, client, v1.NamespaceDefault).Run()
	if err != nil {
		t.Errorf("I blame fakeclient: %s", err)
	}

	// We expect one build (two pods, two secrets) to be deleted.
	if num != 1 {
		t.Errorf("expected 1 deletion, got %d", num)
	}

	secrets, _ = client.CoreV1().Secrets(v1.NamespaceDefault).List(meta.ListOptions{})
	if len(secrets.Items) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(secrets.Items))
	}
	pods, _ = client.CoreV1().Pods(v1.NamespaceDefault).List(meta.ListOptions{})
	if len(pods.Items) != 1 {
		t.Fatalf("expected 1 pods, got %d", len(pods.Items))
	}
}

func TestRun_Max(t *testing.T) {
	client := setupFakeClient()
	num, err := New(time.Time{}, 1, client, v1.NamespaceDefault).Run()
	if err != nil {
		t.Errorf("error running: %s", err)
	}

	// We expect one build (two pods, two secrets) to be deleted.
	if num != 1 {
		t.Errorf("expected 1 deletion, got %d", num)
	}

	secrets, _ := client.CoreV1().Secrets(v1.NamespaceDefault).List(meta.ListOptions{})
	if len(secrets.Items) != 1 {
		t.Errorf("expected 1 secret, got %d", len(secrets.Items))
	}
	pods, _ := client.CoreV1().Pods(v1.NamespaceDefault).List(meta.ListOptions{})
	if len(pods.Items) != 1 {
		t.Errorf("expected 1 pods, got %d", len(pods.Items))
	}

	// It should be the case that the newest pod is the one left. However,
	// the fake client goes by insertion order.
	if pods.Items[0].Name != "jim" {
		t.Errorf("expected jim to be the last pod, got %q", pods.Items[0].Name)
	}
	if secrets.Items[0].Name != "scrooge" {
		t.Errorf("expected scrooge to be the last secret, got %q", secrets.Items[0].Name)
	}
}

// setupFakeClient creates a fake Kubernetes client with some "old" data.
func setupFakeClient() kubernetes.Interface {
	client := fake.NewSimpleClientset()

	ts := time.Now().AddDate(0, -1, 0)
	started := meta.NewTime(ts)

	buildSecret := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      "queequeg",
			Namespace: v1.NamespaceDefault,
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "build",
				"project":   "moby-dick",
				"build":     "123456",
			},
			CreationTimestamp: started,
		},
	}

	jobSecret := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      "tashtego",
			Namespace: v1.NamespaceDefault,
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "job",
				"project":   "moby-dick",
				"build":     "123456",
			},
			CreationTimestamp: started,
		},
	}
	unrelatedSecret := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      "scrooge",
			Namespace: v1.NamespaceDefault,
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "build",
				"project":   "christmas-carol",
				"build":     "723457",
			},
			CreationTimestamp: meta.NewTime(time.Now().AddDate(1, 0, 0)),
		},
	}

	cs := client.CoreV1().Secrets(v1.NamespaceDefault)
	cs.Create(&buildSecret)
	cs.Create(&jobSecret)
	cs.Create(&unrelatedSecret)

	buildPod := v1.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name: "queequeg",
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "build",
				"project":   "moby-dick",
				"build":     "123456",
			},
			CreationTimestamp: started,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Image: "foo",
				},
			},
		},
	}
	jobPod := v1.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name: "tashtego",
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "job",
				"project":   "moby-dick",
				"build":     "123456",
			},
			CreationTimestamp: started,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Image: "foo",
				},
			},
		},
	}
	unrelatedPod := v1.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name: "jim",
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "job",
				"project":   "hart-of-darkness",
				"build":     "923456",
			},
			CreationTimestamp: started,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Image: "foo",
				},
			},
		},
	}

	cb := client.CoreV1().Pods(v1.NamespaceDefault)
	cb.Create(&buildPod)
	cb.Create(&jobPod)
	cb.Create(&unrelatedPod)

	return client
}
