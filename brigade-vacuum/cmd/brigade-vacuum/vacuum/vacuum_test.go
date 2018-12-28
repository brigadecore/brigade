package vacuum

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	testBuildPod1Name = "queequeg"
	testJobPod11Name  = "tashtego"
	testBuildPod2Name = "queequeg2"
	testJobPod21Name  = "tashtego21"
	testJobPod22Name  = "tashtego22"
)

func TestRun_Age(t *testing.T) {
	client := setupFakeClient()

	secrets, err := client.CoreV1().Secrets(v1.NamespaceDefault).List(meta.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(secrets.Items) != 6 {
		t.Fatalf("expected 6 secrets, got %d", len(secrets.Items))
	}
	pods, err := client.CoreV1().Pods(v1.NamespaceDefault).List(meta.ListOptions{})
	if err != nil {
		t.Fatal("no pods returned")
	}
	if len(pods.Items) != 6 {
		t.Fatalf("expected 6 pods, got %d", len(pods.Items))
	}

	err = New(time.Now(), NoMaxBuilds, false, client, v1.NamespaceDefault).Run()
	if err != nil {
		t.Errorf("I blame fakeclient: %s", err)
	}

	verifyPodsDeleted(t, client, testBuildPod1Name, testJobPod11Name, testBuildPod2Name, testJobPod21Name, testJobPod22Name)

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
	err := New(time.Time{}, 1, false, client, v1.NamespaceDefault).Run()
	if err != nil {
		t.Errorf("error running: %s", err)
	}

	verifyPodsDeleted(t, client, testBuildPod1Name, testJobPod11Name, testBuildPod2Name, testJobPod21Name, testJobPod22Name)

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

func TestRun_SkipRunningBuilds(t *testing.T) {
	client := setupFakeClient()

	secrets, err := client.CoreV1().Secrets(v1.NamespaceDefault).List(meta.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	err = New(time.Now(), NoMaxBuilds, true, client, v1.NamespaceDefault).Run()
	if err != nil {
		t.Errorf("I blame fakeclient: %s", err)
	}

	verifyPodsExist(t, client, testBuildPod2Name, testJobPod21Name, testJobPod22Name)
	verifyPodsDeleted(t, client, testBuildPod1Name, testJobPod11Name)

	secrets, _ = client.CoreV1().Secrets(v1.NamespaceDefault).List(meta.ListOptions{})
	if len(secrets.Items) != 4 {
		t.Fatalf("expected 4 secret, got %d", len(secrets.Items))
	}
	pods, _ := client.CoreV1().Pods(v1.NamespaceDefault).List(meta.ListOptions{})
	if len(pods.Items) != 4 {
		t.Fatalf("expected 4 pods, got %d", len(pods.Items))
	}
}

func verifyPodsDeleted(t *testing.T, client kubernetes.Interface, podNames ...string) {
	for _, podName := range podNames {
		_, err := client.CoreV1().Pods(v1.NamespaceDefault).Get(podName, meta.GetOptions{})
		if !errors.IsNotFound(err) {
			t.Errorf("expected Pod %s to be deleted", podName)
		}
	}
}

func verifyPodsExist(t *testing.T, client kubernetes.Interface, podNames ...string) {
	for _, podName := range podNames {
		_, err := client.CoreV1().Pods(v1.NamespaceDefault).Get(podName, meta.GetOptions{})
		if errors.IsNotFound(err) {
			t.Errorf("Pod %s cannot be found (was it deleted?)", podName)
		}
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
	buildSecret2 := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      "queequeg2",
			Namespace: v1.NamespaceDefault,
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "build",
				"project":   "moby-dick",
				"build":     "234567",
			},
			CreationTimestamp: started,
		},
	}
	jobSecret21 := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      "tashtego21",
			Namespace: v1.NamespaceDefault,
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "job",
				"project":   "moby-dick",
				"build":     "234567",
			},
			CreationTimestamp: started,
		},
	}
	jobSecret22 := v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      "tashtego22",
			Namespace: v1.NamespaceDefault,
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "job",
				"project":   "moby-dick",
				"build":     "234567",
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
	cs.Create(&buildSecret2)
	cs.Create(&jobSecret21)
	cs.Create(&jobSecret22)
	cs.Create(&unrelatedSecret)

	buildPod := v1.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name: testBuildPod1Name,
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
		Status: v1.PodStatus{
			Phase: v1.PodSucceeded,
		},
	}
	jobPod := v1.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name: testJobPod11Name,
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
		Status: v1.PodStatus{
			Phase: v1.PodSucceeded,
		},
	}
	buildPod2 := v1.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name: testBuildPod2Name,
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "build",
				"project":   "moby-dick",
				"build":     "234567",
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
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	jobPod21 := v1.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name: testJobPod21Name,
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "job",
				"project":   "moby-dick",
				"build":     "234567",
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
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}
	jobPod22 := v1.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name: testJobPod22Name,
			Labels: map[string]string{
				"heritage":  "brigade",
				"component": "job",
				"project":   "moby-dick",
				"build":     "234567",
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
		Status: v1.PodStatus{
			Phase: v1.PodSucceeded,
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
	cb.Create(&buildPod2)
	cb.Create(&jobPod21)
	cb.Create(&jobPod22)
	cb.Create(&unrelatedPod)

	return client
}
