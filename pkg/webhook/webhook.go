package webhook

import (
	"crypto/subtle"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/Masterminds/vcs"
	"github.com/google/go-github/github"
	"gopkg.in/gin-gonic/gin.v1"

	"github.com/deis/acid/pkg/acid"
)

const (
	acidJSFile         = "acid.js"
	hubSignatureHeader = "X-Hub-Signature"
)

type store interface {
	GetProject(id string) (*acid.Project, error)
	CreateBuild(build *acid.Build) error
}

type githubHook struct {
	store        store
	getFile      fileGetter
	createStatus statusCreator
}

type fileGetter func(commit, path string, proj *acid.Project) ([]byte, error)

type statusCreator func(commit string, proj *acid.Project, status *github.RepoStatus) error

// NewGithubHook creates a GitHub webhook handler.
func NewGithubHook(s store) *githubHook {
	return &githubHook{
		store:        s,
		getFile:      getFile,
		createStatus: setRepoStatus,
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
	case "push", "pull_request":
		s.handleEvent(c, event)
		return
	default:
		log.Printf("Expected event push, got %s", event)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Invalid X-GitHub-Event Header"})
		return
	}
}

func (s *githubHook) handleEvent(c *gin.Context, eventType string) {
	var status = new(github.RepoStatus)

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Failed to read body: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Malformed body"})
		return
	}
	defer c.Request.Body.Close()

	repo, commit, err := parsePayload(eventType, body)
	if err != nil {
		if err == ignoreAction {
			c.JSON(http.StatusOK, gin.H{"status": "Action skipped"})
			return
		}
		log.Printf("Failed to parse payload: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"status": "Received data is not valid JSON"})
		return
	}

	targetURL := &url.URL{
		Scheme: "http",
		Host:   c.Request.Host,
		Path:   path.Join("log", repo, "id", commit),
	}
	tURL := targetURL.String()
	log.Printf("TARGET URL: %s", tURL)
	status.TargetURL = &tURL

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

	go s.buildStatus(eventType, commit, body, proj, status)

	c.JSON(http.StatusOK, gin.H{"status": "Complete"})
}

// buildStatus runs a build, and sets upstream status accordingly.
func (s *githubHook) buildStatus(eventType, commit string, payload []byte, proj *acid.Project, status *github.RepoStatus) {
	// If we need an SSH key, set it here
	if proj.Repo.SSHKey != "" {
		key, err := ioutil.TempFile("", "")
		if err != nil {
			log.Printf("error creating ssh key cache: %s", err)
			return
		}
		keyfile := key.Name()
		defer os.Remove(keyfile)
		if _, err := key.WriteString(proj.Repo.SSHKey); err != nil {
			log.Printf("error writing ssh key cache: %s", err)
			return
		}
		os.Setenv("ACID_REPO_KEY", keyfile)
		defer os.Unsetenv("ACID_REPO_KEY") // purely defensive... not really necessary
	}

	msg := "Building"
	svc := StatusContext
	status.State = &StatePending
	status.Description = &msg
	status.Context = &svc
	if err := s.createStatus(commit, proj, status); err != nil {
		// For this one, we just log an error and continue.
		log.Printf("Error setting status to %s: %s", *status.State, err)
	}
	if err := s.build(eventType, commit, payload, proj); err != nil {
		log.Printf("Build failed: %s", err)
		msg = truncAt(err.Error(), 140)
		status.State = &StateFailure
		status.Description = &msg
	} else {
		msg = "Acid build passed"
		status.State = &StateSuccess
		status.Description = &msg
	}
	if err := s.createStatus(commit, proj, status); err != nil {
		// For this one, we just log an error and continue.
		log.Printf("After build, error setting status to %s: %s", *status.State, err)
	}
}

func truncAt(str string, max int) string {
	if len(str) > max {
		short := str[0 : max-3]
		return short + "..."
	}
	return str
}

func parsePayload(eventType string, payload []byte) (repo string, commit string, err error) {
	e, err := github.ParseWebHook(eventType, payload)
	if err != nil {
		return "", "", err
	}
	switch e := e.(type) {
	case *github.PushEvent:
		return e.Repo.GetFullName(), e.HeadCommit.GetID(), nil
	case *github.PullRequestEvent:
		return e.Repo.GetFullName(), e.PullRequest.Head.GetSHA(), checkPullRequestAction(e)
	}
	return "", "", errors.New("failed parsing event")
}

func checkPullRequestAction(event *github.PullRequestEvent) error {
	switch event.GetAction() {
	case "opened", "synchronize", "reopened":
		return nil
	}
	return ignoreAction
}

var ignoreAction = errors.New("ignored")

// TODO create abstraction for mocking
func getFile(commit, path string, proj *acid.Project) ([]byte, error) {
	toDir := filepath.Join("_cache", proj.Repo.Name)
	if err := os.MkdirAll(toDir, 0755); err != nil {
		log.Printf("error making %s: %s", toDir, err)
		return nil, err
	}

	// URL is the definitive location of the Git repo we are fetching. We always
	// take it from the project, which may choose to set the URL to use any
	// supported Git scheme.
	url := proj.Repo.CloneURL

	// TODO:
	// - [ ] Remove the cached directory at the end of the build?
	if err := cloneRepo(url, commit, toDir); err != nil {
		log.Printf("error cloning %s to %s: %s", url, toDir, err)
		return nil, err
	}

	// Path to acid file:
	acidPath := filepath.Join(toDir, path)
	return ioutil.ReadFile(acidPath)
}

func (s *githubHook) build(eventType, commit string, payload []byte, proj *acid.Project) error {
	acidScript, err := s.getFile(commit, acidJSFile, proj)
	if err != nil {
		return err
	}

	b := &acid.Build{
		ProjectID: proj.ID,
		Type:      eventType,
		Provider:  "github",
		Commit:    commit,
		Payload:   payload,
		Script:    acidScript,
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

type originalError interface {
	Original() error
	Out() string
}

func logOriginalError(err error) {
	oerr, ok := err.(originalError)
	if ok {
		log.Println(oerr.Original().Error())
		log.Println(oerr.Out())
	}
}

func cloneRepo(url, version, toDir string) error {

	// TODO: If the URL is 'file://', do we want to symlink or otherwise copy?
	repo, err := vcs.NewRepo(url, toDir)
	if err != nil {
		return err
	}
	if err := repo.Get(); err != nil {
		logOriginalError(err) // FIXME: Audit this in case this might dump sensitive info.
		if err2 := repo.Update(); err2 != nil {
			logOriginalError(err2)
			log.Printf("WARNING: Could neither clone nor update repo %q. Clone: %s Update: %s", url, err, err2)
		}
	}

	if err := repo.UpdateVersion(version); err != nil {
		log.Printf("Failed to checkout %q: %s", version, err)
		return err
	}

	return nil
}
