package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

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
	router.POST("/webhook/push", pushWebhook)

	router.Run(":7744")
}

func pushWebhook(c *gin.Context) {
	push := &PushHook{}
	if err := c.BindJSON(push); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}

	// Start up a build
	if err := build(push); err != nil {
		log.Printf("error on pushWebhook: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Complete"})
}

func build(push *PushHook) error {
	toDir := filepath.Join("_cache", push.Repository.Name)
	if err := os.MkdirAll(toDir, 0755); err != nil {
		log.Printf("error making %s: %s", toDir, err)
		return err
	}

	if err := cloneRepo(push.Repository.CloneURL, push.HeadCommit.Id, toDir); err != nil {
		log.Printf("error cloning %s to %s: %s", push.Repository.CloneURL, toDir, err)
		return err
	}

	// Normally, here we would load the acid.js file from the repo. But we are
	// hard coding for now.
	acidScript, err := ioutil.ReadFile(acidJS)
	if err != nil {
		return err
	}

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
		log.Printf("WARNING: %s", err)
		//return err
	}

	if err := repo.UpdateVersion(version); err != nil {
		return err
	}

	return nil
}
