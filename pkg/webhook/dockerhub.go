package webhook

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"gopkg.in/gin-gonic/gin.v1"

	"github.com/deis/acid/pkg/acid"
)

type dockerPushHook struct {
	store   store
	getFile fileGetter
}

// NewDockerPushHook creates a new Docker Push handler for webhooks.
func NewDockerPushHook(s store) *dockerPushHook {
	return &dockerPushHook{
		store:   s,
		getFile: getFile,
	}
}

// Handle handles a Push webhook event from DockerHub or a compatible agent.
func (s *dockerPushHook) Handle(c *gin.Context) {
	orgName := c.Param("org")
	projName := c.Param("project")
	commit := c.Param("commit")
	pname := fmt.Sprintf("%s/%s", orgName, projName)

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Failed to read body: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Malformed body"})
		return
	}
	defer c.Request.Body.Close()

	proj, err := s.store.GetProject(pname)
	if err != nil {
		log.Printf("Project %q not found. No secret loaded. %s", pname, err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "project not found"})
		return
	}

	// This will clone the repo before responding to the webhook. We need
	// to make sure that this doesn't cause the hook to hang up.
	acidJS, err := s.getFile(commit, "./acid.js", proj)
	if err != nil {
		log.Printf("aborting DockerImagePush event due to error: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "acidjs not found"})
		return
	}

	go s.notifyDockerImagePush(proj, commit, body, acidJS)
	c.JSON(200, gin.H{"status": "Success"})
}

func (s *dockerPushHook) notifyDockerImagePush(proj *acid.Project, commit string, payload, acidJS []byte) {
	if err := s.doDockerImagePush(proj, commit, payload, acidJS); err != nil {
		log.Printf("failed dockerimagepush event: %s", err)
	}

}

func (s *dockerPushHook) doDockerImagePush(proj *acid.Project, commit string, payload, acidJS []byte) error {
	b := &acid.Build{
		ProjectID: proj.ID,
		Type:      "imagePush",
		Provider:  "dockerhub",
		Commit:    commit,
		Payload:   payload,
		Script:    acidJS,
	}
	return s.store.CreateBuild(b)
}
