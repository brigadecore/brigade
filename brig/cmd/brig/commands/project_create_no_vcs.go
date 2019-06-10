package commands

import (
	"fmt"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/brigadecore/brigade/pkg/storage"

	"gopkg.in/AlecAivazis/survey.v1"
)

// projectCreatePromptsNoVCS handles all of the prompts.
//
// Default values are read from the given project. Values are then
// replaced on that object.
func projectCreatePromptsNoVCS(p *brigade.Project, store storage.Store) error {

	setDefaultValuesNoVCS(p)

	err := setProjectName(p, store, false)
	if err != nil {
		return fmt.Errorf(abort, err)
	}

	err = addEditSecrets(p, store)
	if err != nil {
		return err
	}

	err = addGenericGatewaySecret(p, store)
	if err != nil {
		return err
	}

	// ask/edit a Brigade.js file either in a ConfigMap or local
	err = addBrigadeJS(p, store)
	if err != nil {
		return fmt.Errorf(abort, err)
	}
	// this is a no-VCS Build, so we require a Brigade.js file
	for p.DefaultScript == "" && p.DefaultScriptName == "" {
		fmt.Println("A Brigade.js file is mandatory for no-VCS Brigade Projects (either via a ConfigMap reference or local). Please try again.")
		err = addBrigadeJS(p, store)
		if err != nil {
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
		return projectAdvancedPromptsNoVCS(p, store)
	}
	return nil
}

func projectAdvancedPromptsNoVCS(p *brigade.Project, store storage.Store) error {
	questionsKubernetes, err := advancedQuestionsKubernetes(p, store)
	if err != nil {
		return fmt.Errorf(abort, err)
	}
	if err := survey.Ask(questionsKubernetes, &p.Kubernetes); err != nil {
		return fmt.Errorf(abort, err)
	}

	questionsWorker := advancedQuestionsWorker(p, store)
	if err := survey.Ask(questionsWorker, &p.Worker); err != nil {
		return fmt.Errorf(abort, err)
	}

	questionsProject := advancedQuestionsProject(p, store)
	if err := survey.Ask(questionsProject, p); err != nil {
		return fmt.Errorf(abort, err)
	}

	return nil
}

// setDefaultValuesNoVCS sets some Project defaults for a no-VCS Project
func setDefaultValuesNoVCS(p *brigade.Project) {
	// set some non-VCS Project Default
	// set the name to a non-VCS one
	p.Name = "myproject"
	// setting the sidecar to NONE
	p.Kubernetes.VCSSidecar = "NONE"
	// empty values for the repo
	p.Repo.CloneURL = ""
	p.Repo.Name = ""
}
