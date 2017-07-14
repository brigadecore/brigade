package webhook

import (
	"crypto/subtle"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Masterminds/vcs"
	"github.com/deis/acid/pkg/config"
	"github.com/deis/acid/pkg/js"
	"github.com/google/go-github/github"

	"gopkg.in/gin-gonic/gin.v1"
)

const (
	GitHubEvent  = `X-GitHub-Event`
	HubSignature = `X-Hub-Signature`
)

const acidJS = "acid.js"

// EventRouter routes a webhook to its appropriate handler.
//
// It does this by sniffing the event from the header, and routing accordingly.
func EventRouter(c *gin.Context) {
	event := c.Request.Header.Get(GitHubEvent)
	switch event {
	case "":
		// TODO: Once we're wired up with GitHub, need to return here.
		log.Print("No event header.")
		c.JSON(200, gin.H{"message": "OK"})
		return
	case "ping":
		log.Print("Received ping from GitHub")
		c.JSON(200, gin.H{"message": "OK"})
		return
	case "push":
		Push(c)
		return
	default:
		log.Printf("Expected event push, got %s", event)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Only 'push' is supported. Got " + event})
		return
	}
}

// Push responds to a push event.
func Push(c *gin.Context) {
	var status = new(github.RepoStatus)
	// Only process push for now. Other hooks have different formats.
	signature := c.Request.Header.Get(HubSignature)

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Failed to read body: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Malformed body"})
		return
	}
	defer c.Request.Body.Close()

	push := &PushHook{}
	if err := json.Unmarshal(body, push); err != nil {
		log.Printf("Failed to parse payload: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"status": "Received data is not valid JSON"})
		return
	}

	targetURL := &url.URL{
		Scheme: "http",
		Host:   c.Request.Host,
		Path:   path.Join("log", push.Repository.FullName, "id", push.HeadCommit.Id),
	}
	tURL := targetURL.String()
	log.Printf("TARGET URL: %s", tURL)
	status.TargetURL = &tURL

	// Load config and verify data.
	pname := "acid-" + ShortSHA(push.Repository.FullName)
	ns, _ := config.AcidNamespace(c)
	proj, err := LoadProjectConfig(pname, ns)
	if err != nil {
		log.Printf("Project %q (%q) not found in %q. No secret loaded. %s", push.Repository.FullName, pname, ns, err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "project not found"})
		return
	}

	if proj.SharedSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "No secret is configured for this repo."})
		return
	}

	// Compare the salted digest in the header with our own computing of the
	// body.
	sum := SHA1HMAC([]byte(proj.SharedSecret), body)
	if subtle.ConstantTimeCompare([]byte(sum), []byte(signature)) != 1 {
		log.Printf("Expected signature %q (sum), got %q (hub-signature)", sum, signature)
		//log.Printf("%s", body)
		c.JSON(http.StatusForbidden, gin.H{"status": "malformed signature"})
		return
	}

	if proj.Name != push.Repository.FullName {
		// TODO: Test this. I believe it should error out if these don't match.
		log.Printf("!!!WARNING!!! Expected project secret to have name %q, got %q", push.Repository.FullName, proj.Name)
	}

	go buildStatus(push, proj, status)

	c.JSON(http.StatusOK, gin.H{"status": "Complete"})
}

// buildStatus runs a build, and sets upstream status accordingly.
func buildStatus(push *PushHook, proj *Project, status *github.RepoStatus) {
	// If we need an SSH key, set it here
	if proj.SSHKey != "" {
		key, err := ioutil.TempFile("", "")
		if err != nil {
			log.Printf("error creating ssh key cache: %s", err)
			return
		}
		keyfile := key.Name()
		defer os.Remove(keyfile)
		if _, err := key.WriteString(proj.SSHKey); err != nil {
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
	if err := setRepoStatus(push, proj, status); err != nil {
		// For this one, we just log an error and continue.
		log.Printf("Error setting status to %s: %s", *status.State, err)
	}
	if err := build(push, proj); err != nil {
		log.Printf("Build failed: %s", err)
		msg = truncAt(err.Error(), 140)
		status.State = &StateFailure
		status.Description = &msg
	} else {
		msg = "Acid build passed"
		status.State = &StateSuccess
		status.Description = &msg
	}
	if err := setRepoStatus(push, proj, status); err != nil {
		// For this one, we just log an error and continue.
		log.Printf("After build, error setting status to %s: %s", *status.State, err)
	}
}

func truncAt(str string, max int) string {
	if len(str) > max {
		short := str[0 : max-3]
		return string(short) + "..."
	}
	return str
}

func build(push *PushHook, proj *Project) error {
	toDir := filepath.Join("_cache", push.Repository.FullName)
	if err := os.MkdirAll(toDir, 0755); err != nil {
		log.Printf("error making %s: %s", toDir, err)
		return err
	}

	// URL is the definitive location of the Git repo we are fetching. We always
	// take it from the project, which may choose to set the URL to use any
	// supported Git scheme.
	url := proj.CloneURL

	// TODO:
	// - [ ] Remove the cached directory at the end of the build?
	if err := cloneRepo(url, push.HeadCommit.Id, toDir); err != nil {
		log.Printf("error cloning %s to %s: %s", url, toDir, err)
		return err
	}

	// Path to acid file:
	acidPath := filepath.Join(toDir, acidJS)
	acidScript, err := ioutil.ReadFile(acidPath)
	if err != nil {
		return err
	}
	log.Print(string(acidScript))

	projectID := "acid-" + ShortSHA(proj.Repo)
	e := &js.Event{
		Type:     "push",
		Provider: "github",
		Commit:   push.HeadCommit.Id,
		Payload:  push,
	}

	p := &js.Project{
		ID:   projectID,
		Name: proj.Name,
		Repo: js.Repo{
			Name:     proj.Repo,
			CloneURL: url,
			SSHKey:   strings.Replace(proj.SSHKey, "\n", "$", -1),
		},
		Kubernetes: js.Kubernetes{
			Namespace: proj.Namespace,
			// By putting the sidecar image here, we are allowing an acid.js
			// to override it.
			VCSSidecar: proj.VCSSidecarImage,
		},
		Secrets: proj.Secrets,
	}

	return js.HandleEvent(e, p, acidScript)
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
