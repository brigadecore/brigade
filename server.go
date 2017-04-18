package main

import (
	"crypto/sha256"
	"encoding/hex"
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
	runnerJS = "runner.js"
	acidJS   = "acid.js"
)

func main() {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "OK"}) })
	router.POST("/webhook/push", pushWebhook)

	router.Run(":7744")
}

const (
	GitHubEvent  = `X-GitHub-Event`
	HubSignature = `X-Hub-Signature`
)

func pushWebhook(c *gin.Context) {
	// Only process push for now. Other hooks have different formats.
	event := c.Request.Header.Get(GitHubEvent)
	if event == "ping" {
		log.Print("Received ping from GitHub")
		c.JSON(200, gin.H{"message": "OK"})
	} else if event == "" {
		// TODO: Once we're wired up with GitHub, need to return here.
		log.Print("No event header.")
	} else if event != "push" {
		log.Printf("Expected event push, got %s", event)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Only 'push' is supported. Got " + event})
		return
	}

	// TODO:
	// - [ ] Validate token

	push := &PushHook{}
	if err := c.BindJSON(push); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	// Load config and verify data.
	pname := "acid-" + shortSHA(push.Repository.FullName)
	proj, err := loadProjectConfig(pname, "default")
	if err != nil {
		log.Printf("Project %q (%q) not found. No secret loaded. %s", push.Repository.FullName, pname, err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "project not found"})
		return
	}

	if proj.secret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "No secret is configured for this repo."})
		return
	}

	if proj.name != push.Repository.FullName {
		// TODO: Test this. I believe it should error out if these don't match.
		log.Printf("!!!WARNING!!! Expected project secret to have name %q, got %q", push.Repository.FullName, proj.name)
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

	if err := cloneRepo(push.Repository.CloneURL, push.HeadCommit.Id, toDir); err != nil {
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
		if err := repo.Update(); err != nil {
			log.Printf("WARNING: Could neither clone nor update repo. %s", err)
		}
	}

	if err := repo.UpdateVersion(version); err != nil {
		log.Printf("Failed to checkout %q: %s", version, err)
		return err
	}

	return nil
}

type project struct {
	name   string
	repo   string
	secret string
}

func loadProjectConfig(name, namespace string) (*project, error) {
	kc, err := libk8s.KubeClient()
	proj := &project{}
	if err != nil {
		return proj, err
	}

	// The project config is stored in a secret.
	secret, err := kc.CoreV1().Secrets(namespace).Get(name)
	if err != nil {
		return proj, err
	}

	proj.name = secret.Name
	proj.repo = string(secret.Data["repository"])
	proj.secret = string(secret.Data["secret"])

	return proj, nil
}

func shortSHA(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[0:54]
}
