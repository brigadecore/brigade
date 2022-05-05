package main

import (
	"log"

	"github.com/brigadecore/brigade-foundations/version"
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

	repo, err := clone(event)
	if err != nil {
		log.Fatal(err)
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
