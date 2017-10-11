package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// GitHubBaseURL is the base URL to the GitHub APIv3.
var GitHubBaseURL = "https://api.github.com"

// StatusCode a GitHub status code
type StatusCode string

// Status codes, as defined by GitHub API v3 https://developer.github.com/v3/repos/statuses/
const (
	StatusPending = "pending"
	StatusSuccess = "success"
	StatusError   = "error"
	StatusFailure = "failure"
)

// StatusSetter sets the status of a particular build.
type StatusSetter interface {
	// SetTarget indicates the target that this status applies to
	//
	// owner and repo are the GitHub repository owner and repo names
	// (e.g. github.com/OWNER/REPO)
	//
	// sha is the commit shaw that this status applies to.
	SetTarget(owner, repo, sha string) error
	// SetStatus indicates the status.
	//
	// ctx is a context string that the remote system uses to distinguish multiple
	// status messages. A sane default is "brigade"
	//
	// desc is a human-oriented string that explains the status code
	//
	// targetURL is a callback that displays more information. It may be set to
	// ""
	SetStatus(status StatusCode, ctx, desc string, targetURL string) error
}

// GitHubStatus defines the status format that GitHub expects
type GitHubStatus struct {
	State       StatusCode `json:"state"`
	TargetURL   string     `json:"target_url"`
	Description string     `json:"description"`
	Context     string     `json:"context"`
}

// NewGitHubNotifier creates a new GitHubNotifier
//
// It requires a valid GitHub access token.
func NewGitHubNotifier(token string) *GitHubNotifier {
	return &GitHubNotifier{accessToken: token}
}

// GitHubNotifier sets GitHub status messages via the GitHub APIv3.
type GitHubNotifier struct {
	accessToken string
	owner       string
	repo        string
	sha         string
}

func (g *GitHubNotifier) SetTarget(owner, repo, sha string) error {
	g.owner, g.repo, g.sha = owner, repo, sha
	return nil
}

// url creates a URL to the GitHub API server
//
// This will panic if the var GitHubBaseURL is malformed.
func (g *GitHubNotifier) url() string {
	u, err := url.Parse(GitHubBaseURL)
	if err != nil {
		panic(err)
	}
	u.Path = fmt.Sprintf("/repos/%s/%s/statuses/%s", g.owner, g.repo, g.sha)
	return u.String()
}

func (g *GitHubNotifier) body(status GitHubStatus) (io.Reader, error) {
	b := bytes.NewBuffer(nil)
	data, err := json.Marshal(status)
	if err != nil {
		return b, err
	}
	_, err = b.Write(data)
	return b, err
}

func (g *GitHubNotifier) SetStatus(status StatusCode, ctx, desc, target string) error {
	s := GitHubStatus{
		State:       status,
		Context:     ctx,
		Description: desc,
		TargetURL:   target,
	}
	buf, err := g.body(s)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", g.url(), buf)
	if err != nil {
		return err
	}

	// TODO: Set access token
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("Server responded %s", res.Status)
	}
	return nil
}
