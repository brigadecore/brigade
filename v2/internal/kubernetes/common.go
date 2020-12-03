package kubernetes

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
)

const (
	LabelComponent = "brigade.sh/component"
	LabelEvent     = "brigade.sh/event"
	LabelJob       = "brigade.sh/job"
	LabelProject   = "brigade.sh/project"

	SecretTypeProjectSecrets = "brigade.sh/project-secrets"
	SecretTypeEvent          = "brigade.sh/event"
	SecretTypeJobSecrets     = "brigade.sh/job"
)

func EventSecretName(eventID string) string {
	return fmt.Sprintf("event-%s", eventID)
}

func WorkspacePVCName(eventID string) string {
	return fmt.Sprintf("workspace-%s", eventID)
}

func WorkerPodName(eventID string) string {
	return fmt.Sprintf("worker-%s", eventID)
}

func WorkerPodsSelector() string {
	return labels.Set(
		map[string]string{
			LabelComponent: "worker",
		},
	).AsSelector().String()
}

func JobSecretName(eventID, jobName string) string {
	return fmt.Sprintf("job-%s-%s", eventID, jobName)
}

func JobPodName(eventID, jobName string) string {
	return fmt.Sprintf("job-%s-%s", eventID, jobName)
}

func JobPodsSelector() string {
	return labels.Set(
		map[string]string{
			LabelComponent: "job",
		},
	).AsSelector().String()
}
