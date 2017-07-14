package js

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Repo describes the repository where the source code is stored.
	Repo Repo `json:"repo"`
	// Payload is the raw data as received by the webhook
	Payload interface{} `json:"payload"`
	// ProjectID is the ID of the current project.
	ProjectID string `json:"projectID"`
	// Kubernetes holds information about Kubernetes
	Kubernetes Kubernetes `json:"kubernetes"`
	// Secrets contains passed-in configuration data that is stored in a Secret
	Secrets map[string]string `json:"secrets"`
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

// Kubernetes describes the Kubernetes configuration for a project.
type Kubernetes struct {
	// Namespace is the namespace of this project.
	Namespace string `json:"namespace"`
	// VCSSidecar is the image name/tag for the sidecar that pulls VCS data
	VCSSidecar string `json:"vcsSidecar"`
}
