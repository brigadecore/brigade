package webhook

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
)

// State names for GitHub status
var (
	StatePending = "pending"
	StateFailure = "failure"
	StateError   = "error"
	StateSuccess = "success"
)

// ghClient gets a new GitHub client object.
//
// It authenticates with an OAUTH2 token.
func ghClient(token string) *github.Client {
	t := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	c := context.Background()
	tc := oauth2.NewClient(c, t)
	return github.NewClient(tc)
}

// setRepoStatus sets the status on a particular commit in a repo.
func setRepoStatus(push *PushHook, proj *Project, status *github.RepoStatus) error {
	if proj.GitHubToken == "" {
		return fmt.Errorf("status update skipped because no GitHubToken exists on %s", proj.Name)
	}
	c := context.Background()
	client := ghClient(proj.GitHubToken)
	_, _, err := client.Repositories.CreateStatus(
		c,
		push.Repository.Owner.Name,
		push.Repository.Name,
		push.HeadCommit.Id,
		status)
	return err
}
