package main

import (
	"log"

	"github.com/brigadecore/brigade-foundations/version"
)

const workspace = "/var/vcs"

func main() {
	log.Printf(
		"Starting Brigade Git Initializer -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	if err := applyGlobalConfig(); err != nil {
		log.Fatal(err)
	}

	event, err := getEvent()
	if err != nil {
		log.Fatal(err)
	}

	auth, err := setupAuth(event)
	if err != nil {
		log.Fatal(err)
	}

	if event.Worker.Git.Commit != "" {
		if err = cloneAndCheckoutCommit(
			event.Worker.Git.CloneURL,
			event.Worker.Git.Commit,
		); err != nil {
			log.Fatal(err)
		}
	} else {
		ref := event.Worker.Git.Ref
		if ref == "" {
			if ref, err = getDefaultBranch(
				event.Worker.Git.CloneURL,
				auth,
			); err != nil {
				log.Fatal(err)
			}
		} else {
			ref = getShortRef(ref)
		}
		if err = cloneAndCheckoutRef(event.Worker.Git.CloneURL, ref); err != nil {
			log.Fatal(err)
		}
	}

	if event.Worker.Git.InitSubmodules {
		if err = initSubmodules(); err != nil {
			log.Fatal(err)
		}
	}
}
