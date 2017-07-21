package acid

// Project describes an Acid project
//
// This is an internal representation of a project, and contains data that
// should not be made available to the JavaScript runtime.
type Project struct {
	// Name is the computed name of the project (acid-aeff2343a3234ff)
	Name string
	// Repo is the GitHub repository URL
	Repo string
	// SharedSecret is the GitHub shared key
	SharedSecret string
	// SSHKey is the SSH key the client will use to clone the repo
	SSHKey string
	// GitHubToken is used for oauth2 for client interactions. This is different than the secret.
	GitHubToken string
	// ShortName is the short project name (deis/acid)
	ShortName string

	// The URL to clone for the repository.
	// It may be any Git-compatible URL format
	CloneURL string

	// The namespace to clone into.
	Namespace string

	// VCSSidecarImage is the image that is used as a VCS sidecar for this project.
	VCSSidecarImage string

	// Secrets is environment variables for acid.js
	Secrets map[string]string
}
