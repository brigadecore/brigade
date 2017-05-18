package webhook

import (
	"crypto/subtle"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/vcs"
	"github.com/deis/quokka/pkg/javascript"
	"github.com/deis/quokka/pkg/javascript/libk8s"

	"gopkg.in/gin-gonic/gin.v1"
)

const (
	GitHubEvent  = `X-GitHub-Event`
	HubSignature = `X-Hub-Signature`
)

const (
	runnerJS = "runner.js"
	acidJS   = "acid.js"
)

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

func Push(c *gin.Context) {
	// Only process push for now. Other hooks have different formats.
	signature := c.Request.Header.Get(HubSignature)

	// TODO:
	// - [ ] Validate token
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Failed to read body: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Malformed body"})
		return
	}
	defer c.Request.Body.Close()

	push := &PushHook{}
	if err := json.Unmarshal(body, push); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	// Load config and verify data.
	pname := "acid-" + ShortSHA(push.Repository.FullName)
	proj, err := LoadProjectConfig(pname, "default")
	if err != nil {
		log.Printf("Project %q (%q) not found. No secret loaded. %s", push.Repository.FullName, pname, err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "project not found"})
		return
	}

	if proj.Secret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "No secret is configured for this repo."})
		return
	}

	// Compare the salted digest in the header with our own computing of the
	// body.
	sum := SHA1HMAC([]byte(proj.Secret), body)
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

	// If we need an SSH key, set it here
	if proj.SSHKey != "" {
		key, err := ioutil.TempFile("", "")
		if err != nil {
			log.Printf("error creating ssh key cache: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "Authentication impossible"})
			return
		}
		keyfile := key.Name()
		defer os.Remove(keyfile)
		if _, err := key.WriteString(proj.SSHKey); err != nil {
			log.Printf("error writing ssh key cache: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "Authentication impossible"})
			return
		}
		os.Setenv("ACID_REPO_KEY", keyfile)
		defer os.Setenv("ACID_REPO_KEY", "") // purely defensive... not really necessary
	}

	// Start up a build
	if err := build(push); err != nil {
		log.Printf("error on pushWebhook: %s", err)
		// TODO: Make the returned message pretty. We don't need the error message
		// to go back to GitHub.
		c.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Complete"})
}

func build(push *PushHook) error {
	toDir := filepath.Join("_cache", push.Repository.FullName)
	if err := os.MkdirAll(toDir, 0755); err != nil {
		log.Printf("error making %s: %s", toDir, err)
		return err
	}
	// TODO:
	// - [ ] Remove the cached directory at the end of the build?

	if err := cloneRepo(push.Repository.SSHURL, push.HeadCommit.Id, toDir); err != nil {
		log.Printf("error cloning %s to %s: %s", push.Repository.CloneURL, toDir, err)
		return err
	}

	// Path to acid file:
	acidPath := filepath.Join(toDir, acidJS)
	acidScript, err := ioutil.ReadFile(acidPath)
	if err != nil {
		return err
	}
	log.Print(string(acidScript))

	d, err := ioutil.ReadFile(runnerJS)
	if err != nil {
		return err
	}
	return execScripts(push, d, acidScript)
}

// execScripts prepares the JS runtime and feeds it the objects it needs.
func execScripts(push *PushHook, scripts ...[]byte) error {
	rt := javascript.NewRuntime()
	if err := libk8s.Register(rt.VM); err != nil {
		return err
	}

	// FIXME: This should make its way into quokka.
	rt.VM.Set("sleep", func(seconds int) {
		time.Sleep(time.Duration(seconds) * time.Second)
	})

	out, _ := json.Marshal(push)
	rt.VM.Object("pushRecord = " + string(out))
	for _, script := range scripts {
		if _, err := rt.VM.Run(script); err != nil {
			return err
		}
	}
	return nil
}

func cloneRepo(url, version, toDir string) error {
	repo, err := vcs.NewRepo(url, toDir)
	if err != nil {
		return err
	}
	if err := repo.Get(); err != nil {
		if err2 := repo.Update(); err2 != nil {
			log.Printf("WARNING: Could neither clone nor update repo %q. Clone: %s Update: %s", url, err, err2)
		}
	}

	if err := repo.UpdateVersion(version); err != nil {
		log.Printf("Failed to checkout %q: %s", version, err)
		return err
	}

	return nil
}

type Project struct {
	Name   string
	Repo   string
	Secret string
	SSHKey string
}

func LoadProjectConfig(name, namespace string) (*Project, error) {
	kc, err := libk8s.KubeClient()
	proj := &Project{}
	if err != nil {
		return proj, err
	}

	// The project config is stored in a secret.
	secret, err := kc.CoreV1().Secrets(namespace).Get(name)
	if err != nil {
		return proj, err
	}

	proj.Name = secret.Name
	proj.Repo = string(secret.Data["repository"])
	proj.Secret = string(secret.Data["secret"])
	proj.SSHKey = string(secret.Data["sshKey"])

	return proj, nil
}
