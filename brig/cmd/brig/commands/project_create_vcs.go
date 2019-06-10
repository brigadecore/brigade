package commands

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/goutils"
	survey "gopkg.in/AlecAivazis/survey.v1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/brigadecore/brigade/pkg/storage"
	"github.com/brigadecore/brigade/pkg/storage/kube"
)

// projectCreatePromptsVCS handles all of the prompts.
//
// Default values are read from the given project. Values are then
// replaced on that object.
func projectCreatePromptsVCS(p *brigade.Project, store storage.Store) error {

	err := setProjectName(p, store, true)
	if err != nil {
		return fmt.Errorf(abort, err)
	}

	// a couple of questions that make sense only if the Project is VCS-backed
	qs := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Full repository name",
				Help:    "A protocol-neutral path to your repo, like github.com/foo/bar",
				Default: p.Repo.Name,
			},
		},
		{
			Name: "cloneURL",
			Prompt: &survey.Input{
				Message: "Clone URL (https://github.com/your/repo.git)",
				Help:    "The URL that Git should use to clone. The protocol (https, git, ssh) will determine how the repo is fetched.",
				Default: p.Repo.CloneURL,
			},
		},
	}
	if err = survey.Ask(qs, &p.Repo); err != nil {
		return fmt.Errorf(abort, err)
	}

	// Don't prompt for key if the URL is HTTP(S).
	if !isHTTP(p.Repo.CloneURL) {
		var fname string
		err := survey.AskOne(&survey.Input{
			Message: "Path to SSH key for SSH clone URLs (leave blank to skip)",
			Help:    "The local path to an SSH key file, which will be uploaded to the project. Use this for SSH clone URLs.",
		}, &fname, loadFileValidator)
		if err != nil {
			return fmt.Errorf(abort, err)
		}
		if key := loadFileStr(fname); key != "" {
			p.Repo.SSHKey = replaceNewlines(key)
		}
	}

	err = addEditSecrets(p, store)
	if err != nil {
		return err
	}

	if p.SharedSecret == "" {
		p.SharedSecret, _ = goutils.RandomAlphaNumeric(24)
		fmt.Printf("Auto-generated a Shared Secret: %q\n", p.SharedSecret)
	}

	configureGitHub := false
	if err := survey.AskOne(&survey.Confirm{
		Message: "Configure GitHub Access?",
		Help:    "Configure GitHub CI/CD integration for this project",
	}, &configureGitHub, nil); err != nil {
		return fmt.Errorf(abort, err)
	} else if configureGitHub {
		if err := survey.Ask([]*survey.Question{
			{
				Name: "token",
				Prompt: &survey.Input{
					Message: "OAuth2 token",
					Help:    "Used for contacting the GitHub API. GitHub issues this.",
					Default: p.Github.Token,
				},
			},
			{
				Name: "baseURL",
				Prompt: &survey.Input{
					Message: "GitHub Enterprise URL",
					Help:    "If using GitHub Enterprise, set the base URL here",
					Default: p.Github.BaseURL,
				},
			},
			{
				Name: "uploadURL",
				Prompt: &survey.Input{
					Message: "GitHub Enterprise upload URL",
					Help:    "If using GitHub Enterprise, set the upload URL here",
					Default: p.Github.UploadURL,
				},
			},
		}, &p.Github); err != nil {
			return fmt.Errorf(abort, err)
		}
	}

	doAdvanced := false
	if err := survey.AskOne(&survey.Confirm{
		Message: "Configure advanced options",
		Help:    "Show the advanced configuration options for projects",
	}, &doAdvanced, nil); err != nil {
		return fmt.Errorf(abort, err)
	} else if doAdvanced {
		return projectAdvancedPromptsVCS(p, store)
	}
	return nil
}

func projectAdvancedPromptsVCS(p *brigade.Project, store storage.Store) error {
	questionsKubernetes, err := advancedQuestionsKubernetes(p, store)
	if err != nil {
		return fmt.Errorf(abort, err)
	}
	// VCSSidecar question makes sense only when we use a VCS
	questionsKubernetes = append(questionsKubernetes, &survey.Question{
		Name: "vCSSidecar",
		Prompt: &survey.Input{
			Message: "Custom VCS sidecar (enter 'NONE' for no sidecar)",
			Help:    "The default sidecar uses Git to fetch your repository (enter 'NONE' for no sidecar)",
			Default: p.Kubernetes.VCSSidecar,
		},
	})
	if err := survey.Ask(questionsKubernetes, &p.Kubernetes); err != nil {
		return fmt.Errorf(abort, err)
	}

	questionsWorker := advancedQuestionsWorker(p, store)
	if err := survey.Ask(questionsWorker, &p.Worker); err != nil {
		return fmt.Errorf(abort, err)
	}

	questionsProject := advancedQuestionsProject(p, store)
	// adding a couple of questions that make sense only when we do have a VCS
	questionsProject = append(questionsProject, []*survey.Question{
		{
			Name: "initGitSubmodules",
			Prompt: &survey.Confirm{
				Message: "Initialize Git submodules",
				Help:    "For repos that have submodules, initialize them on each clone. Not recommended on public repos.",
				Default: p.InitGitSubmodules,
			},
		},
		{
			Name: "brigadejsPath",
			Prompt: &survey.Input{
				Message: "brigade.js file path relative to the repository root",
				Help:    "brigade.js file path relative to the repository root, e.g. 'mypath/brigade.js'",
				Default: p.BrigadejsPath,
			},
			Validate: func(ans interface{}) error {
				sans := fmt.Sprintf("%v", ans)
				if filepath.IsAbs(sans) {
					return errors.New("Path must be relative")
				}
				return nil
			},
		},
	}...)
	if err := survey.Ask(questionsProject, p); err != nil {
		return fmt.Errorf(abort, err)
	}

	err = addBrigadeJS(p, store)
	if err != nil {
		return fmt.Errorf(abort, err)
	}

	err = addGenericGatewaySecret(p, store)
	if err != nil {
		return fmt.Errorf(abort, err)
	}

	return nil
}

// loadFileValidator validates that a file exists and can be read.
func loadFileValidator(val interface{}) error {
	name := os.ExpandEnv(val.(string))
	if name == "" {
		return nil
	}
	_, err := ioutil.ReadFile(name)
	return err
}

// loadFileStr should not be called unless loadFileValidator is called first.
func loadFileStr(name string) string {
	if name == "" {
		return ""
	}
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return ""
	}
	return string(data)
}

func replaceNewlines(data string) string {
	return strings.Replace(data, "\n", "$", -1)
}

// loadProjectConfig loads a project configuration from the local filesystem.
func loadProjectConfig(file string, proj *brigade.Project) (*brigade.Project, error) {
	rdr, err := os.Open(file)
	if err != nil {
		return proj, err
	}
	defer rdr.Close()

	sec, err := parseSecret(rdr)
	if err != nil {
		return proj, err
	}

	if sec.Name == "" {
		return proj, fmt.Errorf("secret in %s is missing required name field", file)
	}
	return kube.NewProjectFromSecret(sec, "")
}

func parseSecret(reader io.Reader) (*v1.Secret, error) {
	dec := yaml.NewYAMLOrJSONDecoder(reader, 4096)
	secret := &v1.Secret{}
	// We are only decoding the first item in the YAML.
	err := dec.Decode(secret)

	// Convert StringData to Data
	if len(secret.StringData) > 0 {
		if secret.Data == nil {
			secret.Data = map[string][]byte{}
		}
		for key, val := range secret.StringData {
			secret.Data[key] = []byte(val)
		}
	}

	return secret, err
}

func isHTTP(str string) bool {
	str = strings.ToLower(str)
	return strings.HasPrefix(str, "http:") || strings.HasPrefix(str, "https:")
}
