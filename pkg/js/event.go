package js

// Event describes a Runner event.
type Event struct {
	// Type is the event type (push, pull_request, tag, etc.)
	Type string `json:"type"`
	// Provider is the name of the service that caused the event (github, vsts, cron, ...)
	Provider string `json:"provider"`
	// Commit is the ID of the VCS version, such as the Git commit SHA.
	Commit string `json:"commit"`
	// Repo describes the repository where the source code is stored.
	Repo Repo `json:"repo"`
	// Payload is the raw data as received by the webhook
	Payload interface{} `json:"payload"`
	// ProjectID is the ID of the current project.
	ProjectID string `json:"projectID"`
	// Kubernetes holds information about Kubernetes
	Kubernetes Kubernetes `json:"kubernetes"`
}

// Repo describes a Git repository.
type Repo struct {
	// Name of the repository. For GitHub, this is of the form `org/name` or `user/name`
	Name string `json:"name"`
	// CloneURL is the URL at which the repository can be cloned
	// Traditionally, this is an HTTPS URL.
	CloneURL string `json:"cloneURL"`
	// SSHKey is the auth string for SSH-based cloning
	SSHKey string `json:"sshKey"`
}

type Kubernetes struct {
	Namespace  string `json:"namespace"`
	VCSSidecar string `json:"vcsSidecar"`
}
