package webhook

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/Azure/brigade/pkg/brigade"
)

// State names for GitHub status
var (
	StatePending = "pending"
	StateFailure = "failure"
	StateError   = "error"
	StateSuccess = "success"
)

// StatusContext names the context for a particular status message.
const StatusContext = "brigade"

var ctx = context.Background()

// ghClient gets a new GitHub client object.
//
// It authenticates with an OAUTH2 token.
//
// If an enterpriseHost base URL is provided, this will attempt to connect to
// that instead of the hosted GitHub API server.
func ghClient(gh brigade.Github) (*github.Client, error) {
	t := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: gh.Token})
	tc := oauth2.NewClient(ctx, t)
	if gh.BaseURL != "" {
		return github.NewEnterpriseClient(gh.BaseURL, gh.UploadURL, tc)
	}
	return github.NewClient(tc), nil
}

// setRepoStatus sets the status on a particular commit in a repo.
func setRepoStatus(commit string, proj *brigade.Project, status *github.RepoStatus) error {
	if proj.Github.Token == "" {
		return fmt.Errorf("status update skipped because no GitHubToken exists on %s", proj.Name)
	}
	owner, repo, err := parseRepoName(proj.Repo.Name)
	if err != nil {
		return err
	}
	client, err := ghClient(proj.Github)
	if err != nil {
		return err
	}
	_, _, err = client.Repositories.CreateStatus(ctx, owner, repo, commit, status)
	return err
}

// GetRepoStatus gets the Brigade repository status.
// The ref can be a SHA or a branch or tag.
func GetRepoStatus(proj *brigade.Project, ref string) (*github.RepoStatus, error) {
	client, err := ghClient(proj.Github)
	if err != nil {
		return nil, err
	}
	owner, repo, err := parseRepoName(proj.Repo.Name)
	if err != nil {
		return nil, err
	}
	statii, _, err := client.Repositories.ListStatuses(ctx, owner, repo, ref, &github.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, status := range statii {
		if *status.Context == StatusContext {
			return status, nil
		}
	}
	return nil, fmt.Errorf("no brigade status found")
}

// GetLastCommit gets the last commit on the give reference (branch name or tag).
func GetLastCommit(proj *brigade.Project, ref string) (string, error) {
	client, err := ghClient(proj.Github)
	if err != nil {
		return "", err
	}
	owner, repo, err := parseRepoName(proj.Repo.Name)
	if err != nil {
		return "", err
	}
	sha, _, err := client.Repositories.GetCommitSHA1(ctx, owner, repo, ref, "")
	return sha, err
}

// GetFileContents returns the contents for a particular file in the project.
func GetFileContents(proj *brigade.Project, ref, path string) ([]byte, error) {
	client, err := ghClient(proj.Github)
	if err != nil {
		return []byte{}, err
	}
	owner, repo, err := parseRepoName(proj.Repo.Name)
	if err != nil {
		return nil, err
	}
	opts := &github.RepositoryContentGetOptions{Ref: ref}
	r, err := client.Repositories.DownloadContents(ctx, owner, repo, path, opts)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

type webhookClient interface {
	CreateHook(context.Context, string, string, *github.Hook) (*github.Hook, *github.Response, error)
	ListHooks(context.Context, string, string, *github.ListOptions) ([]*github.Hook, *github.Response, error)
	DeleteHook(context.Context, string, string, int64) (*github.Response, error)
}

var defaultEvents = []string{"pull_request", "push"}

func CreateHook(proj *brigade.Project, host string) error {
	client, err := ghClient(proj.Github)
	if err != nil {
		return err
	}
	return createHook(client.Repositories, proj, host)
}

func createHook(client webhookClient, proj *brigade.Project, host string) error {
	owner, repo, err := parseRepoName(proj.Repo.Name)
	if err != nil {
		return err
	}

	hooks, _, err := client.ListHooks(ctx, owner, repo, &github.ListOptions{})
	if err != nil {
		return err
	}

	callback := fmt.Sprintf("%s/events/github", host)
	events := defaultEvents

	if hook := matchingHook(hooks, proj.ID, callback); hook != nil {
		// get any user configured events
		events = hook.Events

		if _, err := client.DeleteHook(ctx, owner, repo, *hook.ID); err != nil {
			return err
		}
	}

	hook := &github.Hook{
		Name:   github.String("web"),
		Active: github.Bool(true),
		Events: events,
		Config: map[string]interface{}{
			"content_type": "json",
			"insecure_ssl": 0,
			"secret":       proj.Github.Token,
			"url":          callback,
		},
	}

	_, _, err = client.CreateHook(ctx, owner, repo, hook)
	return err
}

func matchingHook(hooks []*github.Hook, projectID, callback string) *github.Hook {
	for _, hook := range hooks {
		if hook.ID == nil {
			continue
		}
		u, ok := hook.Config["url"].(string)
		if !ok || callback != u {
			continue
		}
		p, ok := hook.Config["brigade.sh/project"].(string)
		if !ok || projectID != p {
			continue
		}
		return hook
	}
	return nil
}

func parseRepoName(name string) (owner, repo string, err error) {
	parts := strings.SplitN(name, "/", 3)
	if len(parts) != 3 {
		return "", "", fmt.Errorf("project name %q is malformed", name)
	}
	return parts[1], parts[2], nil
}
