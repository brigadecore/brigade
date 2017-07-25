package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/deis/acid/pkg/storage"
	"github.com/deis/acid/pkg/webhook"
	"github.com/google/go-github/github"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "required arg: Git SHA")
		os.Exit(1)
	}
	commit := os.Args[1]

	eventType := "push"

	data, err := ioutil.ReadFile("./_functional_tests/test-repo-push.json")
	if err != nil {
		panic(err)
	}

	event, err := github.ParseWebHook(eventType, data)
	if err != nil {
		panic(err)
	}

	var repo string

	switch event := event.(type) {
	case *github.PushEvent:
		event.HeadCommit.ID = github.String(commit)
		repo = event.Repo.GetFullName()
	case *github.PullRequestEvent:
		event.PullRequest.Head.SHA = github.String(commit)
		repo = event.Repo.GetFullName()
	}

	out, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		panic(err)
	}

	secret := getSecret(repo)
	hmac := webhook.SHA1HMAC([]byte(secret), out)

	ioutil.WriteFile("./_functional_tests/test-repo-generated.json", out, 0755)
	ioutil.WriteFile("./_functional_tests/test-repo-generated.hash", []byte(hmac), 0755)

	fmt.Fprintln(os.Stdout, string(out))
	fmt.Fprintln(os.Stdout, hmac)
}

func getSecret(pname string) string {
	proj, err := storage.New().Get(pname, "default")
	if err != nil {
		panic(err)
	}
	return proj.SharedSecret
}
