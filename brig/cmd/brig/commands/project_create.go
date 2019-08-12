package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"

	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/Masterminds/goutils"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/brigadecore/brigade/pkg/storage"
	"github.com/brigadecore/brigade/pkg/storage/kube"

	"github.com/spf13/cobra"
)

const (
	abort = "project could not be saved: %s"

	projectCreateUsage = `Create a new project.

Create a new project by answering questions or supplying a configuration file.

For Projects that use a Version Control System (VCS) like GitHub or BitBucket,
Project names are typically in the form 'ORG/PROJECT' or 'USER/PROJECT". For
example, brigadecore/brigade or brigadecore/empty-testbed. For non-VCS Projects,
feel free to use your own preferred naming terms.

A Brigade project provides a context in which a brigade.js file is executed.
Projects can be tied to (Git) source code repositories, and contain
configuration data such as secrets, Kubernetes-specific directives, and
settings for gateways or workers.

If a file is provided (-f), then the project defaults will be set from the file.

The user will be prompted to answer questions to build the configuration. If -f
is specified in conjunction with -x/--no-prompts, then the values in the file will
be used without prompting the user for any changes.
`

	leftUndefined = "Leave undefined"
)

var (
	projectCreateConfig      string
	projectCreateOut         string
	projectCreateFromProject string
	projectCreateDryRun      bool
	projectCreateNoPrompts   bool
	projectCreateReplace     bool
)

// defaultProjectVCS has the default project settings.
// Rather than use this directly, you should get a copy from newProject().
var defaultProjectVCS = brigade.Project{
	Name: "brigadecore/empty-testbed",
	Repo: brigade.Repo{
		Name:     "github.com/brigadecore/empty-testbed",
		CloneURL: "https://github.com/brigadecore/empty-testbed.git",
	},
	Secrets: map[string]string{},
	Worker: brigade.WorkerConfig{
		PullPolicy: "IfNotPresent",
	},
	Kubernetes: brigade.Kubernetes{
		VCSSidecar: "brigadecore/git-sidecar:latest",
	},
}

// newProjectVCS clones the default project.
func newProjectVCS() *brigade.Project {
	proj := defaultProjectVCS
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

	// set the default to VCS
	proj := newProjectVCS()
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
		if err := projectCreateVCSOrNoVCS(proj, store); err != nil {
			return err
		}
	}

	// Disable sidecar container if set to NONE
	// this is used to cancel the default being set on function `newProjectVCS()`
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

func projectCreateVCSOrNoVCS(p *brigade.Project, store storage.Store) error {
	var vcsQuestionResult string
	err := survey.Ask([]*survey.Question{
		{
			Name: "VCSOrNoVCS",
			Prompt: &survey.Select{
				Message: "VCS or no-VCS project?",
				Help:    "Does your Project require the pull of a repo stored on a Version Control System (e.g. GitHub or BitBucket)?",
				Options: []string{"VCS", "no-VCS"},
			},
		},
	}, &vcsQuestionResult)

	if err != nil {
		return fmt.Errorf(abort, err)
	}

	if vcsQuestionResult == "VCS" {
		return projectCreatePromptsVCS(p, store)
	}

	return projectCreatePromptsNoVCS(p, store)

}

func addEditSecrets(p *brigade.Project, store storage.Store) error {
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
	return nil
}

func addGenericGatewaySecret(p *brigade.Project, store storage.Store) error {
	err := survey.AskOne(&survey.Input{
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

// advancedQuestionsKubernetes asks Kubernetes related questions
func advancedQuestionsKubernetes(p *brigade.Project, store storage.Store) ([]*survey.Question, error) {
	questions := []*survey.Question{
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
		return nil, err
	}

	// in the unlikely event that there are no storage classes installed, let the user know
	if len(scn) == 0 {
		fmt.Println("Warning: there are 0 StorageClasses in the cluster. Will not set StorageClasses for this Brigade Project")
	} else {
		scn = append(scn, leftUndefined)
		storageClassQuestions := []*survey.Question{
			{
				Name: "buildStorageClass",
				Prompt: &survey.Select{
					Message: "Build storage class",
					Help: "Kubernetes provides named storage classes. If you want " +
						"to use a custom storage\nclass, set the class name here." +
						"\n\n" +
						"Choose \"Leave undefined\" IF Brigade is configured with a " +
						"default build storage\nclass (which may differ from the " +
						"cluster-wide default storage class) and you\nwish to use that " +
						"Brigade-level default.",
					Options: scn,
				},
			},
			{
				Name: "cacheStorageClass",
				Prompt: &survey.Select{
					Message: "Job cache storage class",
					Help: "Kubernetes provides named storage classes. If you want " +
						"to use a custom storage\nclass, set the class name here." +
						"\n\n" +
						"Choose \"Leave undefined\" IF Brigade is configured with a " +
						"default cache storage\nclass (which may differ from the " +
						"cluster-wide default storage class) and you\nwish to use that " +
						"Brigade-level default.",
					Options: scn,
				},
			},
		}
		questions = append(questions, storageClassQuestions...)
	}

	return questions, nil
}

// advancedQuestionsWorker asks Worker related questions
func advancedQuestionsWorker(p *brigade.Project, store storage.Store) []*survey.Question {
	return []*survey.Question{
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
	}
}

// advancedQuestionsProject asks Project related questions
func advancedQuestionsProject(p *brigade.Project, store storage.Store) []*survey.Question {
	return []*survey.Question{
		{
			Name: "workerCommand",
			Prompt: &survey.Input{
				Message: "Worker command",
				Help:    "EXPERT: Override the worker's default command",
				Default: "",
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
	}
}

func addBrigadeJS(p *brigade.Project, store storage.Store) error {
	err := survey.AskOne(&survey.Input{
		Message: "Default script ConfigMap name",
		Help:    "It is possible to store a default script in a ConfigMap. Supply the name of that ConfigMap to use the script.",
		Default: p.DefaultScriptName,
	}, &p.DefaultScriptName, nil)
	if err != nil {
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

	return nil
}

func setProjectName(p *brigade.Project, store storage.Store, configureVCS bool) error {
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

	// We do configure VCS details only if asked (i.e. only if the Project is a VCS Project)
	// If the name changes, let's quickly try to set some sensible defaults.
	// If the name has not changed, we assume we should keep using the previously
	// loaded values, since this is an update or a namespace move or something
	// similar.
	if p.Name != initialName && configureVCS {
		p.Repo.Name = fmt.Sprintf("github.com/%s", p.Name)
		p.Repo.CloneURL = fmt.Sprintf("https://%s.git", p.Repo.Name)
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
