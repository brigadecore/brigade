package commands

import (
	"testing"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	stubProject1ID = "brigade-544b459e6ad7267e7791c4f77bfd1722a15e305a22cf9d3c60c5be"
	stubProject2ID = "brigade-544b459e6ad7267e7791c4f77bfd1722a15e305a22cf9d3c60c5bf"
	stubBuild1ID   = "01cxmy71nbq7nasvth8pva1s21"
	stubBuild2ID   = "01cxmy71nbq7nasvth8pva1s22"
	stubBuild3ID   = "01cxmy71nbq7nasvth8pva1s23"
)

var (
	stubDT1Start     = time.Now().Add(-5 * time.Minute)
	stubTimeDT1Start = metav1.NewTime(stubDT1Start)
	stubDT1End       = time.Now().Add(-2 * time.Minute)

	stubDT2Start     = time.Now().Add(-time.Minute)
	stubTimeDT2Start = metav1.NewTime(stubDT2Start)

	stubDT3Start     = time.Now().Add(-3 * time.Minute)
	stubTimeDT3Start = metav1.NewTime(stubDT3Start)
	stubDT3End       = time.Now().Add(-time.Minute)
)

func TestGetEmptyBuildList(t *testing.T) {
	client := fake.NewSimpleClientset()
	bls, err := getBuilds("", client, 0)
	if err != nil {
		t.Error(err)
	}
	if len(bls) != 0 {
		t.Error("Error in getBuilds for no project(s)")
	}

	bls, err = getBuilds("", client, 5)
	if err != nil {
		t.Error(err)
	}
	if len(bls) != 0 {
		t.Error("Error in getBuilds for no project(s) with --count 5")
	}
}

// TestGetBuildList tests the command `brig build list`
func TestGetBuildList(t *testing.T) {
	client := fake.NewSimpleClientset()
	createFakeBuilds(t, client)
	bls, err := getBuilds("", client, 0)
	if err != nil {
		t.Error(err)
	}

	if len(bls) != 3 {
		t.Error("Error in getBuilds for all projects")
	}

	if bls[0].since != "???" || bls[0].ID != stubBuild2ID {
		t.Error("Error in build2 time")
	}
	if bls[1].since != "1m" || bls[1].ID != stubBuild3ID {
		t.Error("Error in build3 time")
	}
	if bls[2].since != "2m" || bls[2].ID != stubBuild1ID {
		t.Error("Error in build1 time")
	}
}

// TestGetBuildListWithProject tests the command `brig build list projectID`
func TestGetBuildListWithProject(t *testing.T) {
	client := fake.NewSimpleClientset()
	createFakeBuilds(t, client)
	bls, err := getBuilds(stubProject1ID, client, 0)
	if err != nil {
		t.Error(err)
	}

	if len(bls) != 2 {
		t.Errorf("Error in getBuilds for project %s", stubProject1ID)
	}

	if bls[0].since != "???" || bls[0].ID != stubBuild2ID {
		t.Error("Error in build2")
	}
	if bls[1].since != "2m" || bls[1].ID != stubBuild1ID {
		t.Error("Error in build1")
	}

}

// TestGetBuildListCountTwo tests the command `brig build list --count 2`
func TestGetBuildListCountTwo(t *testing.T) {
	client := fake.NewSimpleClientset()
	createFakeBuilds(t, client)
	bls, err := getBuilds("", client, 2)
	if err != nil {
		t.Error(err)
	}

	if len(bls) != 2 {
		t.Error("Error in getBuilds for '--count 2'")
	}

	if bls[0].since != "???" || bls[0].ID != stubBuild2ID {
		t.Error("Error in build2 time")
	}
	if bls[1].since != "1m" || bls[1].ID != stubBuild3ID {
		t.Error("Error in build3 time")
	}

}

// createFakeBuilds creates necessary Pods/Secrets for 3 fake builds/jobs
// Build1 started 5 minutes ago and finished 2 minutes ago
// Build2 started 1 minute ago and still running
// Build3 started 3 minutes ago and finished 1 minute ago and belongs to a different project than build1 and build2
func createFakeBuilds(t *testing.T, client kubernetes.Interface) {
	stubProject1Secret := createStubProjectSecret(stubProject1ID)
	_, err := client.CoreV1().Secrets("default").Create(stubProject1Secret)
	if err != nil {
		t.Error(err)
	}

	stubBuild1Secret := createStubBuildSecret(stubProject1ID, stubBuild1ID)
	stubWorker1Pod := createStubPod(stubProject1ID, stubBuild1ID, stubDT1Start, v1.PodStatus{
		Phase:     v1.PodSucceeded,
		StartTime: &stubTimeDT1Start,
		ContainerStatuses: []v1.ContainerStatus{
			{
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{
						ExitCode:   0,
						FinishedAt: metav1.NewTime(stubDT1End),
					},
				},
			},
		},
	})
	_, err = client.CoreV1().Pods("default").Create(stubWorker1Pod)
	if err != nil {
		t.Error(err)
	}
	_, err = client.CoreV1().Secrets("default").Create(stubBuild1Secret)
	if err != nil {
		t.Error(err)
	}

	stubBuild2Secret := createStubBuildSecret(stubProject1ID, stubBuild2ID)
	stubWorker2Pod := createStubPod(stubProject1ID, stubBuild2ID, stubDT2Start, v1.PodStatus{
		Phase:     v1.PodRunning,
		StartTime: &stubTimeDT2Start,
		ContainerStatuses: []v1.ContainerStatus{
			{
				State: v1.ContainerState{
					Running: &v1.ContainerStateRunning{
						StartedAt: stubTimeDT2Start,
					},
				},
			},
		},
	})
	_, err = client.CoreV1().Pods("default").Create(stubWorker2Pod)
	if err != nil {
		t.Error(err)
	}
	_, err = client.CoreV1().Secrets("default").Create(stubBuild2Secret)
	if err != nil {
		t.Error(err)
	}

	stubBuild3Secret := createStubBuildSecret(stubProject2ID, stubBuild3ID)
	stubWorker3Pod := createStubPod(stubProject2ID, stubBuild3ID, stubDT3Start, v1.PodStatus{
		Phase:     v1.PodSucceeded,
		StartTime: &stubTimeDT3Start,
		ContainerStatuses: []v1.ContainerStatus{
			{
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{
						ExitCode:   0,
						FinishedAt: metav1.NewTime(stubDT3End),
					},
				},
			},
		},
	})
	_, err = client.CoreV1().Pods("default").Create(stubWorker3Pod)
	if err != nil {
		t.Error(err)
	}
	_, err = client.CoreV1().Secrets("default").Create(stubBuild3Secret)
	if err != nil {
		t.Error(err)
	}
}

func createStubProjectSecret(projectID string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectID,
			Labels: map[string]string{
				"project":   projectID,
				"component": "project",
				"heritage":  "brigade",
				"app":       "brigade",
			},
			Annotations: map[string]string{
				"projectName": "deis/empty-testbed",
			},
		},
		Type: "brigade.sh/project",
	}
}

func createStubBuildSecret(projectID string, buildID string) *v1.Secret {
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: buildID,
			Labels: map[string]string{
				"project":   projectID,
				"build":     buildID,
				"component": "build",
				"heritage":  "brigade",
				"app":       "brigade",
			},
			Annotations: map[string]string{
				"projectName": "deis/empty-testbed",
			},
		},
		Type: "brigade.sh/build",
	}

	if buildID != "" {
		secret.ObjectMeta.Labels["build"] = buildID
		secret.ObjectMeta.Labels["app"] = "brigade"
	}
	return &secret
}

func createStubPod(projectID string, buildID string, startTime time.Time, podStatus v1.PodStatus) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-id-" + buildID,
			Labels: map[string]string{
				"build":     buildID,
				"project":   projectID,
				"component": "build",
				"heritage":  "brigade",
			},
			CreationTimestamp: metav1.NewTime(startTime),
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Image: "alpine:3.7",
				},
			},
		},
		Status: podStatus,
	}
}
