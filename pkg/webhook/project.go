package webhook

import (
	"encoding/json"
	"strings"

	"github.com/deis/acid/pkg/acid"
	"github.com/deis/quokka/pkg/javascript/libk8s"
)

const DefaultVCSSidecar = "acidic.azurecr.io/vcs-sidecar:latest"

// LoadProjectConfig loads a project config from inside of Kubernetes.
//
// The namespace is the namespace where the secret is stored.
func LoadProjectConfig(name, namespace string) (*acid.Project, error) {
	kc, err := libk8s.KubeClient()
	proj := &acid.Project{}
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

func configureProject(proj *acid.Project, data map[string][]byte, namespace string) error {
	proj.Repo = def(data["repository"], proj.Name)
	proj.SharedSecret = def(data["sharedSecret"], "")
	proj.GitHubToken = string(data["githubToken"])
	// Note that we have to undo the key escaping.
	proj.SSHKey = strings.Replace(string(data["sshKey"]), "$", "\n", -1)

	proj.CloneURL = def(data["cloneURL"], "")
	proj.Namespace = def(data["namespace"], namespace)
	proj.VCSSidecarImage = def(data["vcsSidecar"], DefaultVCSSidecar)

	envVars := map[string]string{}
	if d := data["secrets"]; len(d) > 0 {
		if err := json.Unmarshal(d, &envVars); err != nil {
			return err
		}
	}

	proj.Secrets = envVars
	return nil
}
