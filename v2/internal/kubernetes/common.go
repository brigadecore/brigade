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

	LabelKeyWorker         = "worker"
	LabelKeyJob            = "job"
	LabelKeyEvent          = "event"
	LabelKeyWorkspace      = "workspace"
	LabelKeyProjectSecrets = "project-secrets"

	SecretTypeProjectSecrets = "brigade.sh/project-secrets"
	SecretTypeEvent          = "brigade.sh/event"
	SecretTypeJobSecrets     = "brigade.sh/job"
)

func EventSecretName(eventID string) string {
	return eventID
}

func WorkspacePVCName(eventID string) string {
	return eventID
}

func WorkerPodName(eventID string) string {
	return eventID
}

func WorkerPodsSelector() string {
	return labels.Set(
		map[string]string{
			LabelComponent: LabelKeyWorker,
		},
	).AsSelector().String()
}

func JobSecretName(eventID, jobName string) string {
	return fmt.Sprintf("%s-%s", eventID, jobName)
}

func JobPodName(eventID, jobName string) string {
	return fmt.Sprintf("%s-%s", eventID, jobName)
}

func JobPodsSelector() string {
	return labels.Set(
		map[string]string{
			LabelComponent: LabelKeyJob,
		},
	).AsSelector().String()
}
