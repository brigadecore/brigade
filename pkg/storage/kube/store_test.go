package kube

import (
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
)

const (
	stubProjectID = "brigade-544b459e6ad7267e7791c4f77bfd1722a15e305a22cf9d3c60c5be"
)

var (
	stubBuildID  = genID()
	stubJobID    = "job-" + stubBuildID
	now          = time.Now()
	later        = now.Add(time.Minute)
	podStartTime = metav1.NewTime(now)
	podEndTime   = metav1.NewTime(later)
)

var (
	stubWorkerPod = v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-id-" + stubBuildID,
			Labels: map[string]string{
				"build":     stubBuildID,
				"project":   stubProjectID,
				"component": "build",
				"heritage":  "brigade",
			},
			CreationTimestamp: podStartTime,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Image: "alpine:3.7",
				},
			},
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
	stubJobPod = v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: stubJobID,
			Labels: map[string]string{
				"build":     stubBuildID,
				"project":   stubProjectID,
				"component": "job",
				"heritage":  "brigade",
			},
			CreationTimestamp: podStartTime,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Image: "alpine:3.7",
				},
			},
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

	stubProjectSecret = &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: stubProjectID,
			Labels: map[string]string{
				"project":   stubProjectID,
				"component": "project",
				"heritage":  "brigade",
				"app":       "brigade",
			},
			Annotations: map[string]string{
				"projectName": "deis/empty-testbed",
			},
		},
		Type: secretTypeBuild,
		Data: map[string][]byte{
			"repository":        []byte("myrepo"),
			"defaultScript":     []byte(`console.log("hello default script")`),
			"sharedSecret":      []byte("mysecret"),
			"github.token":      []byte("like a fish needs a bicycle"),
			"github.baseURL":    []byte("https://example.com/base"),
			"github.uploadURL":  []byte("https://example.com/upload"),
			"sshKey":            []byte("hello$world"),
			"namespace":         []byte("zooropa"),
			"secrets":           []byte(`{"bar":"baz","foo":"bar"}`),
			"worker.registry":   []byte("deis"),
			"worker.name":       []byte("brigade-worker"),
			"worker.tag":        []byte("canary"),
			"worker.pullPolicy": []byte("Always"),
			// Intentionally skip cloneURL, test that this is ""
		},
	}

	// stubBuild is a build
	stubBuild = &brigade.Build{
		ID:        stubBuildID,
		ProjectID: stubProjectID,
		Revision: &brigade.Revision{
			Commit: "abc123",
			Ref:    "refs/heads/master",
		},
		Type:     "foo",
		Provider: "bar",
		Payload:  []byte("this is a payload"),
		Script:   []byte("ohai"),
	}
)

// fakeStore returns a fake Kubernetes client and a *store that wraps it.
func fakeStore() (kubernetes.Interface, storage.Store) {
	client := fake.NewSimpleClientset()
	return client, New(client, "default")
}

func createFakeWorker(client kubernetes.Interface, pod v1.Pod) {
	client.CoreV1().Pods("default").Create(&pod)
}

func createFakeJob(client kubernetes.Interface, pod v1.Pod) {
	createFakeWorker(client, pod)
}

func createFakeProject(client kubernetes.Interface, secret *v1.Secret) {
	client.CoreV1().Secrets("default").Create(secret)
}
