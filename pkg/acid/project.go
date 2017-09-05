package acid

// Project describes an Acid project
//
// This is an internal representation of a project, and contains data that
// should not be made available to the JavaScript runtime.
type Project struct {
	// ID is the computed name of the project (acid-aeff2343a3234ff)
	ID string `json:"id"`
	// Name is the human readable name of project.
	Name string `json:"name"`
	// Repo describes the repository where the source code is stored.
	Repo Repo `json:"repo"`
	// Kubernetes holds information about Kubernetes
	Kubernetes Kubernetes `json:"kubernetes"`
	// SharedSecret is the GitHub shared key
	SharedSecret string `json:"shared_secret"`
	// Github holds information about Github.
	Github Github `json:"github"`
	// Secrets is environment variables for acid.js
	Secrets map[string]string `json:"secrets"`
}

// Github describes the Github configuration for a project.
type Github struct {
	// Token is used for oauth2 for client interactions.
	Token string `json:"token"`
}

// Repo describes a Git repository.
type Repo struct {
	// Name of the repository. For GitHub, this is of the form `org/name` or `user/name`
	Name string `json:"name"`
	// Owner of the repositoy. For Github this is `org` or `user`
	Owner string `json:"owner"`
	// CloneURL is the URL at which the repository can be cloned
	// Traditionally, this is an HTTPS URL.
	CloneURL string `json:"cloneURL"`
	// SSHKey is the auth string for SSH-based cloning
	SSHKey string `json:"sshKey"`
}

// DefaultVCSSidecar is the default image that fetches a repo from a VCS.
const DefaultVCSSidecar = "acidic.azurecr.io/vcs-sidecar:latest"

// Kubernetes describes the Kubernetes configuration for a project.
type Kubernetes struct {
	// Namespace is the namespace of this project.
	Namespace string `json:"namespace"`
	// VCSSidecar is the image name/tag for the sidecar that pulls VCS data
	VCSSidecar string `json:"vcsSidecar"`
}
