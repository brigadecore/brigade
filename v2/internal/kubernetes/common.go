package kubernetes

import "k8s.io/apimachinery/pkg/labels"

const (
	LabelComponent = "brigade.sh/component"
	LabelEvent     = "brigade.sh/event"
	LabelJob       = "brigade.sh/job"
	LabelProject   = "brigade.sh/project"

	SecretTypeProjectSecrets = "brigade.sh/project-secrets"
	SecretTypeEvent          = "brigade.sh/event"
	SecretTypeJobSecrets     = "brigade.sh/job"
)

func WorkerPodsSelector() string {
	return labels.Set(
		map[string]string{
			LabelComponent: "worker",
		},
	).AsSelector().String()
}

func JobPodsSelector() string {
	return labels.Set(
		map[string]string{
			LabelComponent: "job",
		},
	).AsSelector().String()
}
