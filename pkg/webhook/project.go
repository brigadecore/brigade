package webhook

import (
	"strings"

	"github.com/deis/quokka/pkg/javascript/libk8s"
)

const DefaultVCSSidecar = "acidic.azurecr.io/vcs-sidecar:latest"

// Project describes an Acid project
type Project struct {
	// Name is the computed name of the project (acid-aeff2343a3234ff)
	Name string
	// Repo is the GitHub repository URL
	Repo string
	// Secret is the GitHub shared key
	Secret string
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
}

// LoadProjectConfig loads a project config from inside of Kubernetes.
//
// The namespace is the namespace where the secret is stored.
func LoadProjectConfig(name, namespace string) (*Project, error) {
	kc, err := libk8s.KubeClient()
	proj := &Project{}
	if err != nil {
		return proj, err
	}

	// The project config is stored in a secret.
	secret, err := kc.CoreV1().Secrets(namespace).Get(name)
	if err != nil {
		return proj, err
	}

	def := func(a []byte, b string) string {
		if len(a) == 0 {
			return b
		}
		return string(a)
	}

	proj.Name = secret.Name
	proj.Repo = def(secret.Data["repository"], proj.Name)
	proj.Secret = def(secret.Data["secret"], "")
	proj.GitHubToken = string(secret.Data["githubToken"])
	// Note that we have to undo the key escaping.
	proj.SSHKey = strings.Replace(string(secret.Data["sshKey"]), "$", "\n", -1)
	proj.ShortName = secret.Annotations["projectName"]

	proj.CloneURL = def(secret.Data["cloneURL"], "")
	proj.Namespace = def(secret.Data["namespace"], namespace)
	proj.VCSSidecarImage = def(secret.Data["vcsSidecar"], DefaultVCSSidecar)

	return proj, nil
}
