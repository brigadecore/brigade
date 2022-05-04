package main

import (
	"log"

	"github.com/brigadecore/brigade-foundations/version"
	git "github.com/libgit2/git2go/v32"
)

func main() {
	log.Printf(
		"Starting Brigade Git Initializer -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	event, err := getEvent("/var/event/event.json")
	if err != nil {
		log.Fatal(err)
	}
	gitConfig := event.Worker.Git
	if gitConfig == nil {
		log.Fatal("event has no git config")
	}

	credentialsCallback, err := getCredentialsCallback(event.Project.Secrets)
	if err != nil {
		log.Fatal(err)
	}

	const workspacePath = "/var/vcs"
	log.Printf("cloning repository from %q into %q",
		gitConfig.CloneURL,
		workspacePath,
	)
	repo, err := git.Clone(
		event.Worker.Git.CloneURL,
		workspacePath,
		&git.CloneOptions{
			FetchOptions: git.FetchOptions{
				RemoteCallbacks: git.RemoteCallbacks{
					CertificateCheckCallback: func(*git.Certificate, bool, string) error {
						return nil
					},
					CredentialsCallback: credentialsCallback,
				},
			},
		},
	)
	if err != nil {
		log.Fatalf(
			"error cloning repository from %q into %q: %s",
			gitConfig.CloneURL,
			workspacePath,
			err,
		)
	}
	defer repo.Free()

	if err = checkout(repo, gitConfig.Commit, gitConfig.Ref); err != nil {
		log.Fatal(err)
	}

	if gitConfig.InitSubmodules {
		if err = initSubmodules(repo); err != nil {
			log.Fatal(err)
		}
	}
}
