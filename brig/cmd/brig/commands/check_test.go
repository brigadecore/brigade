package commands

import (
	"testing"

	apps_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
)

func TestGetDeployStatusSuccess(t *testing.T) {
	statusString := getDeployStatusString(createFakeDeployment(v1.DeploymentStatus{
		Replicas:            9,
		ReadyReplicas:       9,
		UpdatedReplicas:     0,
		AvailableReplicas:   9,
		UnavailableReplicas: 0,
	}), "FakeComponent")

	if statusString != "✔️ FakeComponent replicas: Desired 9, Ready 9, Updated 0, Available 9, Unavailable 0 \n" {
		t.Error("Error in getDeployStatusString for success condition")
	}
}

func TestGetDeployStatusError(t *testing.T) {
	statusString := getDeployStatusString(createFakeDeployment(v1.DeploymentStatus{
		Replicas:            9,
		ReadyReplicas:       8,
		UpdatedReplicas:     0,
		AvailableReplicas:   9,
		UnavailableReplicas: 1,
	}), "FakeComponent")

	if statusString != "❌ FakeComponent replicas: Desired 9, Ready 8, Updated 0, Available 9, Unavailable 1 \n" {
		t.Error("Error in getDeployStatusString for error condition")
	}
}

func createFakeDeployment(deploymentStatus v1.DeploymentStatus) apps_v1.Deployment {
	return apps_v1.Deployment{
		Status: deploymentStatus,
	}
}
