package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/goutils"
	"github.com/spf13/cobra"
	survey "gopkg.in/AlecAivazis/survey.v1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
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

const abort = "project could not be saved: %s"

var (
	projectCreateConfig      string
	projectCreateOut         string
	projectCreateFromProject string
	projectCreateDryRun      bool
	projectCreateNoPrompts   bool
	projectCreateReplace     bool
)

// defaultProject has the default project settings.
// Rather than use this directly, you should get a copy from newProject().
var defaultProject = brigade.Project{
	Name: "deis/empty-testbed",
	Repo: brigade.Repo{
		Name:     "github.com/deis/empty-testbed",
		CloneURL: "https://github.com/deis/empty-testbed.git",
	},
	Secrets: map[string]string{},
	Worker: brigade.WorkerConfig{
		PullPolicy: "IfNotPresent",
	},
	WorkerCommand: "yarn -s start",
	Kubernetes: brigade.Kubernetes{
		VCSSidecar: "deis/git-sidecar:latest",
	},
}

// newProject clones the default project.
func newProject() *brigade.Project {
	proj := defaultProject
	return &proj
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
			return errors.New("cannot specify both -f/--config and -p/--from-project")
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

	proj := newProject()
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
		if err := projectCreatePrompts(proj, store); err != nil {
			return err
		}
	}

	// Disable sidecar container if set to NONE
	if proj.Kubernetes.VCSSidecar == "NONE" {
		proj.Kubernetes.VCSSidecar = ""
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
		fmt.Printf("%s\n", data)
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

	// brig project create --replace
	if projectCreateReplace {
		if _, err = store.GetProject(proj.ID); err != nil {
			return fmt.Errorf("project %s could not be found (error: %s). Cannot replace, exiting", proj.Name, err.Error())
		}
		return store.ReplaceProject(proj)
	}

	// brig project create # no replace
	if _, err = store.GetProject(proj.ID); err == nil {
		return fmt.Errorf("project %s already exists. Refusing to overwrite", proj.Name)
	}
	// Store the project
	return store.CreateProject(proj)
}

// projectCreatePrompts handles all of the prompts.
//
// Default values are read from the given project. Values are then
// replaced on that object.
func projectCreatePrompts(p *brigade.Project, store storage.Store) error {
	// We always set this to the globalNamespace, otherwise things will break.
	p.Kubernetes.Namespace = globalNamespace

	message := "Project Name"
	if projectCreateReplace {
		message = "Existing " + message
	}

	initialName := p.Name
	if err := survey.AskOne(&survey.Input{
		Message: message,
		Help:    "By convention, this is user/project or org/project",
		Default: p.Name,
	}, &p.Name, survey.Required); err != nil {
		return err
	}

	// If the name changes, let's quickly try to set some sensible defaults.
	// If the name has not changed, we assume we should keep using the previously
	// loaded values, since this is an update or a namespace move or something
	// similar.
	if p.Name != initialName {
		p.Repo.Name = fmt.Sprintf("github.com/%s", p.Name)
		p.Repo.CloneURL = fmt.Sprintf("https://%s.git", p.Repo.Name)
	}

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

	if err := survey.Ask(qs, &p.Repo); err != nil {
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

	if len(p.Secrets) > 0 {
		fmt.Println("The following secrets are already defined:")
		for k := range p.Secrets {
			fmt.Printf("\t- %s\n", k)
		}
		fmt.Println("  (To remove a secret, enter the key for the name, and '-' as the value)")
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
		return projectAdvancedPrompts(p, store)
	}
	return nil
}

func projectAdvancedPrompts(p *brigade.Project, store storage.Store) error {
	questions := []*survey.Question{
		{
			Name: "vCSSidecar",
			Prompt: &survey.Input{
				Message: "Custom VCS sidecar (enter 'NONE' for no sidecar)",
				Help:    "The default sidecar uses Git to fetch your repository",
				Default: p.Kubernetes.VCSSidecar,
			},
		},
		{
			Name: "buildStorageSize",
			Prompt: &survey.Input{
				Message: "Build storage size",
				Help:    "By default, 50Mi of shared temp space is allocated per build. Larger values slow down build startup. Units are Ki, Mi, or Gi",
				Default: p.Kubernetes.BuildStorageSize,
			},
		},
		{
			Name: "allowSecretKeyRef",
			Prompt: &survey.Confirm{
				Message: "SecretKeyRef usage",
				Help:    "Allow or disallow usage of secretKeyRef in job environments.",
				Default: p.Kubernetes.AllowSecretKeyRef,
			},
		},
		{
			Name: "serviceAccount",
			Prompt: &survey.Input{
				Message: "Project Service Account",
				Help:    "The kubernetes service account that the project should use. If not given the controller's default will be used",
				Default: "",
			},
		},
	}

	// get the StorageClass.Name for the storage classes in the cluster
	scn, err := store.GetStorageClassNames()
	if err != nil {
		// this error indicates a cluster communication problem, so exit now since project creation will probably fail as well
		return err
	}

	// in the unlikely event that there are no storage classes installed, let the user know
	if len(scn) == 0 {
		fmt.Println("Warning: there are 0 StorageClasses in the cluster. Will not set StorageClasses for this Brigade Project")
	} else {
		storageClassQuestions := []*survey.Question{
			{
				Name: "buildStorageClass",
				Prompt: &survey.Select{
					Message: "Build storage class",
					Help:    "Kubernetes provides named storage classes. If you want to use a custom storage class, set the class name here.",
					Options: scn,
				},
			},
			{
				Name: "cacheStorageClass",
				Prompt: &survey.Select{
					Message: "Job cache storage class",
					Help:    "Kubernetes provides named storage classes. If you want to use a custom storage class, set the class name here.",
					Options: scn,
				},
			},
		}
		questions = append(questions, storageClassQuestions...)
	}

	if err := survey.Ask(questions, &p.Kubernetes); err != nil {
		return fmt.Errorf(abort, err)
	}

	if err := survey.Ask([]*survey.Question{
		{
			Name: "registry",
			Prompt: &survey.Input{
				Message: "Worker image registry or DockerHub org",
				Help:    "For non-DockerHub, this is the root URL. For DockerHub, it is the org.",
				Default: p.Worker.Registry,
			},
		},
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Worker image name",
				Help:    "The name of the worker image, e.g. workerImage",
				Default: p.Worker.Name,
			},
		},
		{
			Name: "tag",
			Prompt: &survey.Input{
				Message: "Custom worker image tag",
				Help:    "The worker image tag to pull, e.g. 1.2.3 or latest",
				Default: p.Worker.Tag,
			},
		},
		{
			Name: "pullPolicy",
			Prompt: &survey.Select{
				Message: "Worker image pull policy",
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
	if err := survey.Ask([]*survey.Question{
		{
			Name: "workerCommand",
			Prompt: &survey.Input{
				Message: "Worker command",
				Help:    "EXPERT: Override the worker's default command (yarn -s start)",
				Default: p.WorkerCommand,
			},
		},
		{
			Name: "initGitSubmodules",
			Prompt: &survey.Confirm{
				Message: "Initialize Git submodules",
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
				Message: "Image pull secrets",
				Help:    "Comma-separated list of image pull secret names that will be supplied to workers and jobs",
				Default: p.ImagePullSecrets,
			},
		},
		{
			Name: "defaultScriptName",
			Prompt: &survey.Input{
				Message: "Default script ConfigMap name",
				Help:    "EXPERT: It is possible to store a default script in a ConfigMap. Supply the name of that ConfigMap to use the script.",
				Default: p.DefaultScriptName,
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
	}, p); err != nil {
		return fmt.Errorf(abort, err)
	}
	var fname string
	err = survey.AskOne(&survey.Input{
		Message: "Upload a default brigade.js script",
		Help:    "The local path to a default brigade.js file that will be run if none exists in the repo. Overrides the ConfigMap script.",
	}, &fname, loadFileValidator)
	if err != nil {
		return fmt.Errorf(abort, err)
	}
	if script := loadFileStr(fname); script != "" {
		p.DefaultScript = loadFileStr(fname)
	}

	err = survey.AskOne(&survey.Input{
		Message: "Secret for the Generic Gateway (alphanumeric characters only). Press Enter if you want it to be auto-generated",
		Help:    "This is the secret that secures the Generic Gateway. Only alphanumeric characters are accepted. Provide an empty string if you want an auto-generated one",
	}, &p.GenericGatewaySecret, genericGatewaySecretValidator)
	if err != nil {
		return fmt.Errorf(abort, err)
	}

	// user pressed Enter key, so let's auto-generate a GenericGateway secret
	if p.GenericGatewaySecret == "" {
		var err error
		p.GenericGatewaySecret, err = goutils.RandomAlphaNumeric(5)
		if err != nil {
			return fmt.Errorf("Error in generating Generic Gateway Secret: %s", err.Error())
		}
		fmt.Printf("Auto-generated Generic Gateway Secret: %s\n", p.GenericGatewaySecret)
	}

	return nil
}

// genericGatewaySecretValidator validates the secret provided by user for the Generic Gateway
// this can be either "" (so it will be auto-generated) or alphanumeric
func genericGatewaySecretValidator(val interface{}) error {
	re := regexp.MustCompile("^[a-zA-Z0-9]*$")
	if val.(string) == "" {
		return nil
	}
	if re.MatchString(val.(string)) {
		return nil
	}
	return fmt.Errorf("Generic Gateway secret should only contain alphanumeric characters")
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
