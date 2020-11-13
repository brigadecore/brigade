package kubernetes

const (
	LabelComponent = "brigade.sh/component"
	LabelEvent     = "brigade.sh/event"
	LabelJob       = "brigade.sh/job"
	LabelProject   = "brigade.sh/project"

	SecretTypeProjectSecrets = "brigade.sh/project-secrets"
	SecretTypeEvent          = "brigade.sh/event"
	SecretTypeJobSecrets     = "brigade.sh/job"
)
