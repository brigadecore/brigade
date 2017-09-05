package acid

import "k8s.io/client-go/pkg/api/v1"

type Build struct {
	// ID is the unique ID for a webhook event.
	ID string `json:"id"`
	// ProjectID is the computed name of the project (acid-aeff2343a3234ff)
	ProjectID string `json:"project_id"`
	// Type is the event type (push, pull_request, tag, etc.)
	Type string `json:"type"`
	// Provider is the name of the service that caused the event (github, vsts, cron, ...)
	Provider string `json:"provider"`
	// Commit is the ID of the VCS version, such as the Git commit SHA.
	Commit string `json:"commit"`
	// Payload is the raw data as received by the webhook.
	Payload []byte `json:"payload"`
	// Script is the acidJS to be executed.
	Script []byte `json:"script"`
}

func NewBuildFromSecret(secret v1.Secret) *Build {
	return &Build{
		ID:        secret.ObjectMeta.Labels["build"],
		ProjectID: secret.ObjectMeta.Labels["project"],
		Type:      string(secret.Data["event_type"]),
		Provider:  string(secret.Data["event_provider"]),
		Commit:    secret.ObjectMeta.Labels["commit"],
		Payload:   secret.Data["payload"],
		Script:    secret.Data["script"],
	}
}
