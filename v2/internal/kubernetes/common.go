package kubernetes

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
)

const (
	AnnotationTimeoutDuration = "brigade.sh/timeoutDuration"

	LabelBrigadeID = "brigade.sh/id"
	LabelComponent = "brigade.sh/component"
	LabelEvent     = "brigade.sh/event"
	LabelJob       = "brigade.sh/job"
	LabelProject   = "brigade.sh/project"

	LabelKeyWorker         = "worker"
	LabelKeyJob            = "job"
	LabelKeyEvent          = "event"
	LabelKeyWorkspace      = "workspace"
	LabelKeyProjectSecrets = "project-secrets"

	SecretTypeProjectSecrets = "brigade.sh/project-secrets" // nolint: gosec
	SecretTypeEvent          = "brigade.sh/event"           // nolint: gosec
	SecretTypeJobSecrets     = "brigade.sh/job"             // nolint: gosec
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

func WorkerPodsSelector(brigadeID string) string {
	return labels.Set(
		map[string]string{
			LabelBrigadeID: brigadeID,
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

func JobPodsSelector(brigadeID string) string {
	return labels.Set(
		map[string]string{
			LabelBrigadeID: brigadeID,
			LabelComponent: LabelKeyJob,
		},
	).AsSelector().String()
}
