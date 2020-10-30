package kubernetes

const (
	LabelComponent = "brigade.sh/component"
	LabelProject   = "brigade.sh/project"
	LabelEvent     = "brigade.sh/event"

	SecretTypeProjectSecrets = "brigade.sh/project-secrets"
	SecretTypeEvent          = "brigade.sh/event"
)
