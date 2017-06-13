package webhook

import (
	"strings"

	"github.com/deis/quokka/pkg/javascript/libk8s"
)

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
}

// LoadProjectConfig loads a project config from inside of Kubernetes.
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
	proj.Repo = string(secret.Data["repository"])
	proj.Secret = string(secret.Data["secret"])
	proj.GitHubToken = string(secret.Data["githubToken"])
	// Note that we have to undo the key escaping.
	proj.SSHKey = strings.Replace(string(secret.Data["sshKey"]), "$", "\n", -1)
	proj.ShortName = secret.Annotations["projectName"]

	return proj, nil
}
