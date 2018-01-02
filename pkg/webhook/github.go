package webhook

import (
	"crypto/subtle"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"gopkg.in/gin-gonic/gin.v1"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
)

const (
	brigadeJSFile      = "brigade.js"
	hubSignatureHeader = "X-Hub-Signature"
)

type githubHook struct {
	store                   storage.Store
	getFile                 fileGetter
	createStatus            statusCreator
	buildForkedPullRequests bool
}

type fileGetter func(commit, path string, proj *brigade.Project) ([]byte, error)

type statusCreator func(commit string, proj *brigade.Project, status *github.RepoStatus) error

// NewGithubHook creates a GitHub webhook handler.
func NewGithubHook(s storage.Store, buildForkedPullRequests bool) *githubHook {
	return &githubHook{
		store: s,
		buildForkedPullRequests: buildForkedPullRequests,
		getFile:                 getFileFromGithub,
		createStatus:            setRepoStatus,
	}
}

// Handle routes a webhook to its appropriate handler.
//
// It does this by sniffing the event from the header, and routing accordingly.
func (s *githubHook) Handle(c *gin.Context) {
	event := c.Request.Header.Get("X-GitHub-Event")
	switch event {
	case "ping":
		log.Print("Received ping from GitHub")
		c.JSON(200, gin.H{"message": "OK"})
		return
	case "push", "pull_request", "create", "release", "status", "commit_comment", "pull_request_review":
		s.handleEvent(c, event)
		return
	default:
		// Issue #127: Don't return an error for unimplemented events.
		log.Printf("Unsupported event %q", event)
		c.JSON(200, gin.H{"message": "Ignored"})
		return
	}
}

func (s *githubHook) handleEvent(c *gin.Context, eventType string) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Failed to read body: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Malformed body"})
		return
	}
	defer c.Request.Body.Close()

	e, err := github.ParseWebHook(eventType, body)
	if err != nil {
		log.Printf("Failed to parse body: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Malformed body"})
		return
	}

	var repo, commit string

	switch e := e.(type) {
	case *github.PushEvent:
		// If this is a branch deletion, skip the build.
		if e.GetDeleted() {
			c.JSON(http.StatusOK, gin.H{"status": "build skipped on branch deletion"})
			return
		}

		repo = e.Repo.GetFullName()
		commit = e.HeadCommit.GetID()
	case *github.PullRequestEvent:
		if !s.isAllowedPullRequest(e) {
			c.JSON(http.StatusOK, gin.H{"status": "build skipped"})
			return
		}

		// EXPERIMENTAL: Since labeling and unlabeling PRs doesn't really have a
		// code impact, we don't really want to fire off the same event (or require
		// the user to know the event details). So we add a pseudo-event for labeling
		// actions.
		if a := e.GetAction(); a == "labeled" || a == "unlabeled" {
			eventType = "pull_request:" + a
		}

		repo = e.Repo.GetFullName()
		commit = e.PullRequest.Head.GetSHA()
	case *github.CommitCommentEvent:
		repo = e.Repo.GetFullName()
		commit = e.Comment.GetCommitID()
	case *github.CreateEvent:
		// TODO: There are three ref_type values: tag, branch, and repo. Do we
		// want to be opinionated about how we handle these?
		repo = e.Repo.GetFullName()
		commit = e.GetRef() // Note that this is a ref, not a commit.
	case *github.ReleaseEvent:
		repo = e.Repo.GetFullName()
		commit = e.Release.GetTagName() // Note this is a tag name, not a commit
	case *github.StatusEvent:
		repo = e.Repo.GetFullName()
		commit = e.Commit.GetSHA()
	case *github.PullRequestReviewEvent:
		repo = e.Repo.GetFullName()
		commit = e.PullRequest.Head.GetSHA()
	default:
		log.Printf("Failed to parse payload")
		c.JSON(http.StatusBadRequest, gin.H{"status": "Received data is not valid JSON"})
		return
	}

	proj, err := s.store.GetProject(repo)
	if err != nil {
		log.Printf("Project %q not found. No secret loaded. %s", repo, err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "project not found"})
		return
	}

	if proj.SharedSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "No secret is configured for this repo."})
		return
	}

	signature := c.Request.Header.Get(hubSignatureHeader)
	if err := validateSignature(signature, proj.SharedSecret, body); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "malformed signature"})
		return
	}

	if proj.Name != repo {
		// TODO: Test this. I believe it should error out if these don't match.
		log.Printf("!!!WARNING!!! Expected project secret to have name %q, got %q", repo, proj.Name)
	}

	s.buildStatus(eventType, commit, body, proj)

	c.JSON(http.StatusOK, gin.H{"status": "Complete"})
}

// buildStatus runs a build, and sets upstream status accordingly.
func (s *githubHook) buildStatus(eventType, commit string, payload []byte, proj *brigade.Project) {
	msg := "Building"
	svc := StatusContext
	status := new(github.RepoStatus)
	status.State = &StatePending
	status.Description = &msg
	status.Context = &svc
	if err := s.build(eventType, commit, payload, proj); err != nil {
		log.Printf("Creating Build failed: %s", err)
		msg = truncAt(err.Error(), 140)
		status.State = &StateFailure
		status.Description = &msg
	}
	if err := s.createStatus(commit, proj, status); err != nil {
		// For this one, we just log an error and continue.
		log.Printf("Error setting status to %s: %s", *status.State, err)
	}
}

// isAllowedPullRequest returns true if this particular pull request is allowed
// to produce an event.
func (s *githubHook) isAllowedPullRequest(e *github.PullRequestEvent) bool {
	if !s.buildForkedPullRequests && e.PullRequest.Head.Repo.GetFork() {
		// Log this case for debugging.
		log.Println("skipping forked pull request")
		return false
	}
	switch e.GetAction() {
	case "opened", "synchronize", "reopened", "labeled", "unlabeled":
		return true
	}
	log.Println("unsupported pull_request action:", e.GetAction())
	return false
}

func truncAt(str string, max int) string {
	if len(str) > max {
		short := str[0 : max-3]
		return short + "..."
	}
	return str
}

func getFileFromGithub(commit, path string, proj *brigade.Project) ([]byte, error) {
	return GetFileContents(proj, commit, path)
}

func (s *githubHook) build(eventType, commit string, payload []byte, proj *brigade.Project) error {
	brigadeScript, err := s.getFile(commit, brigadeJSFile, proj)
	if err != nil {
		return err
	}

	b := &brigade.Build{
		ProjectID: proj.ID,
		Type:      eventType,
		Provider:  "github",
		Commit:    commit,
		Payload:   payload,
		Script:    brigadeScript,
	}

	return s.store.CreateBuild(b)
}

// validateSignature compares the salted digest in the header with our own computing of the body.
func validateSignature(signature, secretKey string, payload []byte) error {
	sum := SHA1HMAC([]byte(secretKey), payload)
	if subtle.ConstantTimeCompare([]byte(sum), []byte(signature)) != 1 {
		log.Printf("Expected signature %q (sum), got %q (hub-signature)", sum, signature)
		return errors.New("payload signature check failed")
	}
	return nil
}
