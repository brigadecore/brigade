package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/Masterminds/goutils"
	"github.com/spf13/cobra"

	"k8s.io/api/core/v1"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage/kube"
)

const projectCreateUsage = `Create a new project.

Create a new project by answering questions or supplying a configuration file.

Project names are typically in the form 'ORG/PROJECT' or 'USER/PROJECT". For
example, Azure/brigade or deis/empty-testbed.

A Brigade project provides a context in which a brigade.js file is executed.
Projects are frequently tied to (Git) source code repositories, and contain
configuration data such as secrets, Kubernetes-specific directives, and
settings for gateways or workers.

If a file is provided (-f), then the project defaults will be set from the file.

The user will be prompted to answer questions to build the configuration. If -f
is specified in conjunction with -x/--no-prompts, then the values in the file will
be used without prompting the user for any changes.

`

var (
	projectCreateConfig      = ""
	projectCreateOut         = ""
	projectCreateFromProject = ""
	projectCreateDryRun      = false
	projectCreateNoPrompts   = false
	projectCreateReplace     = false
)

var defaultProject = &brigade.Project{
	Name: "deis/empty-testbed",
	Repo: brigade.Repo{
		Name:     "github.com/deis/empty-testbed",
		CloneURL: "https://github.com/deis/empty-testbed.git",
		SSHKey:   "",
	},
	Secrets:      map[string]string{},
	SharedSecret: "",
	Github:       brigade.Github{},
	Kubernetes:   brigade.Kubernetes{},
	Worker: brigade.WorkerConfig{
		PullPolicy: "IfNotPresent",
	},
}

func init() {
	project.AddCommand(projectCreate)
	flags := projectCreate.Flags()
	flags.StringVarP(&projectCreateConfig, "config", "f", "", "Path to JSON Kubernetes Secret.")
	flags.StringVarP(&projectCreateFromProject, "from-project", "p", "", "Retrieve the given project from Kubernetes and use it to set the default values.")
	flags.StringVarP(&projectCreateOut, "out", "o", "", "File where configuration should be saved. The configuration is stored as a JSON Kubernetes Secret")
	flags.BoolVarP(&projectCreateDryRun, "dry-run", "D", false, "Do not send the config to Kubernetes")
	flags.BoolVarP(&projectCreateNoPrompts, "no-prompts", "x", false, "Do not prompt the user for input. Use this with -f to skip prompts.")
	flags.BoolVarP(&projectCreateReplace, "replace", "r", false, "Replace an existing project with this revised one. Use this to modify existing projects.")
}

var projectCreate = &cobra.Command{
	Use:   "create",
	Short: "create a new Brigade project",
	Long:  projectCreateUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectCreateFromProject != "" && projectCreateConfig != "" {
			return errors.New("cannot specify both -f/--config and -p/--from-project.")
		}
		return createProject(cmd.OutOrStderr())
	},
}

func createProject(out io.Writer) error {
	c, err := kubeClient()
	if err != nil {
		return err
	}

	store := kube.New(c, globalNamespace)

	proj := defaultProject
	if projectCreateConfig != "" {
		if proj, err = loadProjectConfig(projectCreateConfig, proj); err != nil {
			return err
		}
	}

	if projectCreateFromProject != "" {
		var err error
		proj, err = store.GetProject(projectCreateFromProject)
		if err != nil {
			return fmt.Errorf("could not load %s: %s", projectCreateFromProject, err)
		}
	}

	if !projectCreateNoPrompts {
		if err := projectCreatePrompts(proj); err != nil {
			return err
		}
	}

	secret, err := kube.SecretFromProject(proj)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(secret, "", "  ")
	if err != nil {
		return err
	}

	if globalVerbose {
		out.Write(data)
		fmt.Println()
	}

	if projectCreateOut != "" {
		if err := ioutil.WriteFile(projectCreateOut, data, 0755); err != nil {
			return err
		}
	}

	fmt.Fprintf(out, "Project ID: %s\n", proj.ID)
	if projectCreateDryRun {
		return nil
	}

	if !projectCreateReplace {
		if _, err = store.GetProject(proj.Name); err == nil {
			return fmt.Errorf("Project %s already exists. Refusing to overwrite.", proj.Name)
		}
	}

	// Store the project
	if err := store.CreateProject(proj); err != nil {
		return err
	}
	return nil
}

// projectCreatePrompts handles all of the prompts.
//
// Default values are read from the given project. Values are then
// replaced on that object.
func projectCreatePrompts(p *brigade.Project) error {
	if err := survey.AskOne(&survey.Input{
		Message: "Project name",
		Help:    "By convention, this is user/project or org/project",
		Default: p.Name,
	}, &p.Name, survey.Required); err != nil {
		return err
	}

	qs := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Full repository name",
				Help:    "A protocol-neutral path to your repo",
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

	abort := "project could not be saved: %s"
	if err := survey.Ask(qs, &p.Repo); err != nil {
		return fmt.Errorf(abort, err)
	}

	// Only prompt for key if clone URL requires it
	if strings.HasPrefix(p.Repo.CloneURL, "ssh") || strings.HasPrefix(p.Repo.CloneURL, "git+ssh") {
		var fname string
		err := survey.AskOne(&survey.Input{
			Message: "Path to SSH key for SSH clone URLs (leave blank to skip)",
			Help:    "The local path to an SSH key file, which will be uploaded to the project. Use this for SSH clone URLs.",
		}, &fname, loadFileValidator)
		if err != nil {
			return fmt.Errorf(abort, err)
		}
		if key := loadFileStr(fname); key != "" {
			p.Repo.SSHKey = key
		}
	}

	if len(p.Secrets) > 0 {
		fmt.Println("The following secrets are already defined:")
		for k := range p.Secrets {
			fmt.Printf("\t- %s\n", k)
		}
		fmt.Println("  (To remove a secret, entre the key for the name, and '-' as the value)")
	}

	addSecrets := false
	if err := survey.AskOne(&survey.Confirm{
		Message: "Add secrets?",
		Help:    "Add key/value pairs for things like account names and tokens. These are available to your brigade.js",
	}, &addSecrets, nil); err != nil {
		return fmt.Errorf(abort, err)
	} else if addSecrets {
		kvp := struct {
			Another bool
			Key     string
			Value   string
		}{
			Another: true,
		}
		for i := 1; kvp.Another; i++ {
			err := survey.Ask([]*survey.Question{
				{
					Name:   "key",
					Prompt: &survey.Input{Message: fmt.Sprintf("\tSecret %d", i)},
				},
				{
					Name:   "value",
					Prompt: &survey.Input{Message: "\tValue"},
				},
				{
					Name:   "another",
					Prompt: &survey.Confirm{Message: "===> Add another?"},
				},
			}, &kvp)
			if err != nil {
				return fmt.Errorf(abort, err)
			}
			if kvp.Value == "-" {
				delete(p.Secrets, kvp.Key)
				continue
			}
			p.Secrets[kvp.Key] = kvp.Value
		}

	}

	if p.SharedSecret != "" {
		fmt.Println("Auto-generating a Shared Secret...")
		p.SharedSecret, _ = goutils.RandomAlphaNumeric(24)
	}

	configureGitHub := false
	if err := survey.AskOne(&survey.Confirm{
		Message: "Configure GitHub Enterprise or OAuth?",
		Help:    "Configure GitHub CI/CD integration for this project",
	}, &configureGitHub, nil); err != nil {
		return fmt.Errorf(abort, err)
	} else if configureGitHub {
		if err := survey.Ask([]*survey.Question{
			{
				Name: "token",
				Prompt: &survey.Input{
					Message: "OAuth2 Token",
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
					Message: "GitHub Enterprise Upload URL",
					Help:    "If using GitHub Enterprise, set the upload URL here",
					Default: p.Github.UploadURL,
				},
			},
		}, &p.Github); err != nil {
			return fmt.Errorf(abort, err)
		}
	}

	configureK8s := false
	if err := survey.AskOne(&survey.Confirm{
		Message: "Configure advanced Kubernetes options?",
		Help:    "Set advanced options for custom VCS images, storage, and permissions",
	}, &configureK8s, nil); err != nil {
		return fmt.Errorf(abort, err)
	} else if configureK8s {
		if err := survey.Ask([]*survey.Question{
			{
				Name: "namespace",
				Prompt: &survey.Input{
					Message: "Kubernetes Namespace",
					Help:    "The namespace in which this project applies",
					Default: p.Kubernetes.Namespace,
				},
			},
			{
				Name: "vCSSidecar",
				Prompt: &survey.Input{
					Message: "Custom VCS Sidecar",
					Help:    "The default sidecar uses Git to fetch your repository",
					Default: p.Kubernetes.VCSSidecar,
				},
			},
			{
				Name: "buildStorageSize",
				Prompt: &survey.Input{
					Message: "Build storage size",
					Help:    "By default, 50Mi of shared temp space is allocated per build. Larger values slow down build startup.",
					Default: p.Kubernetes.BuildStorageSize,
				},
			},
			{
				Name: "buildStorageClass",
				Prompt: &survey.Input{
					Message: "Build Storage Class",
					Help:    "Kubernetes provides named storage classes. If you want to use a custom storage class, set the class name here.",
					Default: p.Kubernetes.BuildStorageClass,
				},
			},
			{
				Name: "cacheStorageClass",
				Prompt: &survey.Input{
					Message: "Job Cache Storage Class",
					Help:    "Kubernetes provides named storage classes. If you want to use a custom storage class, set the class name here.",
					Default: p.Kubernetes.CacheStorageClass,
				},
			},
		}, &p.Kubernetes); err != nil {
			return fmt.Errorf(abort, err)
		}
	}

	configureWorker := false
	if err := survey.AskOne(&survey.Confirm{
		Message: "Configure Brigade worker?",
		Help:    "Set advanced options for the Brigade worker",
	}, &configureWorker, nil); err != nil {
		return fmt.Errorf(abort, err)
	} else if configureWorker {
		if err := survey.Ask([]*survey.Question{
			{
				Name: "registry",
				Prompt: &survey.Input{
					Message: "Image registry or DockerHub org",
					Help:    "For non-DockerHub, this is the root URL",
					Default: p.Worker.Registry,
				},
			},
			{
				Name: "name",
				Prompt: &survey.Input{
					Message: "Image Name",
					Help:    "The name of the image, e.g. workerImage",
					Default: p.Worker.Name,
				},
			},
			{
				Name: "tag",
				Prompt: &survey.Input{
					Message: "Image Tag",
					Help:    "The image tag to pull, e.g. 1.2.3 or latest",
					Default: p.Worker.Tag,
				},
			},
			{
				Name: "pullPolicy",
				Prompt: &survey.Select{
					Message: "Image Pull Policy",
					Help:    "The image pull policy determines how often Kubernetes will try to refresh this image",
					Options: []string{
						"IfNotPresent",
						"Always",
						"Never",
					},
					Default: p.Worker.PullPolicy,
				},
			},
		}, &p.Worker); err != nil {
			return fmt.Errorf(abort, err)
		}
	}

	advanced := false
	if err := survey.AskOne(&survey.Confirm{
		Message: "Configure Advanced Brigade Settings?",
		Help:    "Set advanced options for Brigade",
	}, &advanced, nil); err != nil {
		return fmt.Errorf(abort, err)
	} else if advanced {
		if err := survey.Ask([]*survey.Question{
			{
				Name: "initGitSubmodules",
				Prompt: &survey.Confirm{
					Message: "Initialize Git Submodules",
					Help:    "For repos that have submodules, initialize them on each clone. Not recommended on public repos.",
					Default: p.InitGitSubmodules,
				},
			},
			{
				Name: "allowHostMounts",
				Prompt: &survey.Confirm{
					Message: "Allow host mounts",
					Help:    "Allow host-mounted volumes for worker and jobs. Not recommended in multi-tenant clusters.",
					Default: p.AllowHostMounts,
				},
			},
			{
				Name: "allowPrivilegedJobs",
				Prompt: &survey.Confirm{
					Message: "Allow privileged jobs",
					Help:    "Allow jobs to mount the Docker socket or perform other privileged operations. Not recommended for multi-tenant clusters.",
					Default: p.AllowPrivilegedJobs,
				},
			},
			{
				Name: "imagePullSecrets",
				Prompt: &survey.Input{
					Message: "Image Pull Secrets",
					Help:    "Comma-separated list of image pull secret names that will be supplied to workers and jobs",
					Default: p.ImagePullSecrets,
				},
			},
			{
				Name: "workerCommand",
				Prompt: &survey.Input{
					Message: "Worker Command",
					Help:    "Override the worker's default command (yarn start). For debugging/expert use",
					Default: p.WorkerCommand,
				},
			},
			{
				Name: "defaultScriptName",
				Prompt: &survey.Input{
					Message: "Default Script ConfigMap Name",
					Help:    "It is possible to store a default script in a ConfigMap. Supply the name of that ConfigMap to use the script.",
					Default: p.DefaultScriptName,
				},
			},
		}, p); err != nil {
			return fmt.Errorf(abort, err)
		}
		var fname string
		err := survey.AskOne(&survey.Input{
			Message: "Upload a Default brigade.js Script",
			Help:    "The local path to a default brigade.js file that will be run if none exists in the repo. Overrides the ConfigMap script.",
		}, &fname, loadFileValidator)
		if err != nil {
			return fmt.Errorf(abort, err)
		}
		if script := loadFileStr(fname); script != "" {
			p.DefaultScript = loadFileStr(fname)
		}
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

// loadFileTransformer should not be called unless loadFileValidator is called first.
func loadFileTransformer(val interface{}) interface{} {
	name := os.ExpandEnv(val.(string))
	return loadFileStr(name)
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
	return strings.Replace(string(data), "\n", "$", 0)
}

// loadProjectConfig loads a project configuration from the local filesystem.
func loadProjectConfig(file string, proj *brigade.Project) (*brigade.Project, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return proj, err
	}
	sec := &v1.Secret{}
	if err := json.Unmarshal(data, sec); err != nil {
		return proj, err
	}
	if sec.Name == "" {
		return proj, fmt.Errorf("secret in %s is missing required name field", file)
	}
	newproj, err := kube.NewProjectFromSecret(sec, "")
	return newproj, err
}
