package webhook

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"gopkg.in/gin-gonic/gin.v1"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
)

type githubHook struct {
	store          storage.Store
	createStatus   statusCreator
	allowedAuthors []string
}

type statusCreator func(commit string, proj *brigade.Project, status *github.RepoStatus) error

// NewGithubHook creates a GitHub webhook handler.
func NewGithubHook(s storage.Store, authors []string) gin.HandlerFunc {
	gh := &githubHook{
		store:          s,
		createStatus:   setRepoStatus,
		allowedAuthors: authors,
	}
	return gh.Handle
}

// Handle routes a webhook to its appropriate handler.
//
// It does this by sniffing the event from the header, and routing accordingly.
func (s *githubHook) Handle(c *gin.Context) {
	event := github.WebHookType(c.Request)
	switch event {
	case "push", "pull_request", "create", "release", "status", "commit_comment", "pull_request_review", "deployment", "deployment_status":
		s.handleEvent(c, event)
	case "ping":
		log.Print("Received ping from GitHub")
		c.JSON(200, gin.H{"message": "OK"})
	default:
		// Issue #127: Don't return an error for unimplemented events.
		log.Printf("Unsupported event %q", event)
		c.JSON(200, gin.H{"message": "Ignored"})
	}
}

func (s *githubHook) handleEvent(c *gin.Context, eventType string) {
	projectID := c.Param("project")

	proj, err := s.store.GetProject(projectID)
	if err != nil {
		log.Printf("Project %q not found. No secret loaded. %s", projectID, err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "project not found"})
		return
	}

	if proj.SharedSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "No secret is configured for this repo."})
		return
	}

	payload, err := github.ValidatePayload(c.Request, []byte(proj.SharedSecret))
	if err != nil {
		log.Printf("Failed payload signature check: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"status": "malformed signature"})
		return
	}

	e, err := github.ParseWebHook(eventType, payload)
	if err != nil {
		log.Printf("Failed to parse body: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Malformed body"})
		return
	}

	var rev brigade.Revision

	switch e := e.(type) {
	case *github.PushEvent:
		// If this is a branch deletion, skip the build.
		if e.GetDeleted() {
			c.JSON(http.StatusOK, gin.H{"status": "build skipped on branch deletion"})
			return
		}

		rev.Commit = e.HeadCommit.GetID()
		rev.Ref = e.GetRef()
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

		rev.Commit = e.PullRequest.Head.GetSHA()
		rev.Ref = fmt.Sprintf("refs/pull/%d/head", e.PullRequest.GetNumber())
	case *github.CommitCommentEvent:
		rev.Commit = e.Comment.GetCommitID()
	case *github.CreateEvent:
		// TODO: There are three ref_type values: tag, branch, and repo. Do we
		// want to be opinionated about how we handle these?
		rev.Ref = e.GetRef()
	case *github.ReleaseEvent:
		rev.Ref = e.Release.GetTagName()
	case *github.StatusEvent:
		rev.Commit = e.Commit.GetSHA()
	case *github.PullRequestReviewEvent:
		rev.Commit = e.PullRequest.Head.GetSHA()
		rev.Ref = fmt.Sprintf("refs/pull/%d/head", e.PullRequest.GetNumber())
	case *github.DeploymentEvent:
		rev.Commit = e.Deployment.GetSHA()
		rev.Ref = e.Deployment.GetRef()
	case *github.DeploymentStatusEvent:
		rev.Commit = e.Deployment.GetSHA()
		rev.Ref = e.Deployment.GetRef()
	default:
		log.Printf("Failed to parse payload")
		c.JSON(http.StatusBadRequest, gin.H{"status": "Received data is not valid JSON"})
		return
	}

	s.buildStatus(eventType, rev, payload, proj)

	c.JSON(http.StatusOK, gin.H{"status": "Complete"})
}

// buildStatus runs a build, and sets upstream status accordingly.
func (s *githubHook) buildStatus(eventType string, rev brigade.Revision, payload []byte, proj *brigade.Project) {
	if err := s.build(eventType, rev, payload, proj); err != nil {
		log.Printf("Creating Build failed: %s", err)
		svc := StatusContext
		msg := truncAt(err.Error(), 140)
		status := new(github.RepoStatus)
		status.State = &StatePending
		status.Description = &msg
		status.Context = &svc
		status.State = &StateFailure
		status.Description = &msg
		if err := s.createStatus(rev.Commit, proj, status); err != nil {
			// For this one, we just log an error and continue.
			log.Printf("Error setting status to %s: %s", *status.State, err)
		}
	}
}

// isAllowedPullRequest returns true if this particular pull request is allowed
// to produce an event.
func (s *githubHook) isAllowedPullRequest(e *github.PullRequestEvent) bool {

	isFork := e.PullRequest.Head.Repo.GetFork()

	// This applies the author association to forked PRs.
	// PRs sent against origin will be accepted without a check.
	// See https://developer.github.com/v4/reference/enum/commentauthorassociation/
	if assoc := e.PullRequest.GetAuthorAssociation(); isFork && !s.isAllowedAuthor(assoc) {
		log.Printf("skipping pull request for disallowed author %s", assoc)
		return false
	}
	switch e.GetAction() {
	case "opened", "synchronize", "reopened", "labeled", "unlabeled", "closed":
		return true
	}
	log.Println("unsupported pull_request action:", e.GetAction())
	return false
}

func (s *githubHook) isAllowedAuthor(author string) bool {
	for _, a := range s.allowedAuthors {
		if a == author {
			return true
		}
	}
	return false
}

func truncAt(str string, max int) string {
	if len(str) > max {
		short := str[0 : max-3]
		return short + "..."
	}
	return str
}

func (s *githubHook) build(eventType string, rev brigade.Revision, payload []byte, proj *brigade.Project) error {
	b := &brigade.Build{
		ProjectID: proj.ID,
		Type:      eventType,
		Provider:  "github",
		Revision:  &rev,
		Payload:   payload,
	}

	return s.store.CreateBuild(b)
}
