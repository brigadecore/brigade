package webhook

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"gopkg.in/gin-gonic/gin.v1"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
)

type dockerPushHook struct {
	store   storage.Store
	getFile fileGetter
}

// NewDockerPushHook creates a new Docker Push handler for webhooks.
func NewDockerPushHook(s storage.Store) *dockerPushHook {
	return &dockerPushHook{
		store:   s,
		getFile: getFileFromGithub,
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
	brigadeJS, err := s.getFile(commit, "./brigade.js", proj)
	if err != nil {
		log.Printf("aborting DockerImagePush event due to error: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "brigadejs not found"})
		return
	}

	go s.notifyDockerImagePush(proj, commit, body, brigadeJS)
	c.JSON(200, gin.H{"status": "Success"})
}

func (s *dockerPushHook) notifyDockerImagePush(proj *brigade.Project, commit string, payload, brigadeJS []byte) {
	if err := s.doDockerImagePush(proj, commit, payload, brigadeJS); err != nil {
		log.Printf("failed dockerimagepush event: %s", err)
	}

}

func (s *dockerPushHook) doDockerImagePush(proj *brigade.Project, commit string, payload, brigadeJS []byte) error {
	b := &brigade.Build{
		ProjectID: proj.ID,
		Type:      "imagePush",
		Provider:  "dockerhub",
		Commit:    commit,
		Payload:   payload,
		Script:    brigadeJS,
	}
	return s.store.CreateBuild(b)
}
