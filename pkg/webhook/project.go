package webhook

import (
	"encoding/json"
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

	// Env is environment variables for acid.js
	Env map[string]string
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

	proj.Name = secret.Name
	proj.ShortName = secret.Annotations["projectName"]

	return proj, configureProject(proj, secret.Data, namespace)
}

func def(a []byte, b string) string {
	if len(a) == 0 {
		return b
	}
	return string(a)
}

func configureProject(proj *Project, data map[string][]byte, namespace string) error {
	proj.Repo = def(data["repository"], proj.Name)
	proj.Secret = def(data["secret"], "")
	proj.GitHubToken = string(data["githubToken"])
	// Note that we have to undo the key escaping.
	proj.SSHKey = strings.Replace(string(data["sshKey"]), "$", "\n", -1)

	proj.CloneURL = def(data["cloneURL"], "")
	proj.Namespace = def(data["namespace"], namespace)
	proj.VCSSidecarImage = def(data["vcsSidecar"], DefaultVCSSidecar)

	envVars := map[string]string{}
	if d := data["env"]; len(d) > 0 {
		if err := json.Unmarshal(d, &envVars); err != nil {
			return err
		}
	}

	proj.Env = envVars
	return nil
}
