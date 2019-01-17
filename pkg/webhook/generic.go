package webhook

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"

	gin "gopkg.in/gin-gonic/gin.v1"
)

type genericWebhook struct {
	store storage.Store
}

type genericWebhookData struct {
	Ref    string `json:"ref"`
	Commit string `json:"commit"`
}

// NewGenericWebhook creates a go-restful handler for generic Gateway.
func NewGenericWebhook(s storage.Store) gin.HandlerFunc {
	h := &genericWebhook{store: s}
	return h.Handle
}

// Handle handles a generic Gateway event.
func (g *genericWebhook) Handle(c *gin.Context) {
	projectID := c.Param("projectID")
	secret := c.Param("secret")

	proj, err := g.store.GetProject(projectID)

	if err != nil {
		log.Printf("Project %q not found. No secret loaded. %s", projectID, err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "project not found"})
		return
	}

	// if the secret is "" (probably i) due to a Brigade upgrade or ii) user did not create a Generic Gateway secret during `brig project create`)
	// refuse to serve it, so Brigade admin will be forced to update the project with a non-empty secret
	if proj.GenericGatewaySecret == "" {
		log.Printf("Secret for project %s is empty, please update it and try again", projectID)
		c.JSON(http.StatusUnauthorized, gin.H{"status": "secret for this Brigade Project is empty, refusing to serve, please inform your Brigade admin"})
		return
	}

	// compare secrets
	if secret != proj.GenericGatewaySecret {
		log.Printf("Secret %s for project %s is wrong", secret, projectID)
		c.JSON(http.StatusUnauthorized, gin.H{"status": "secret is wrong"})
		return
	}

	gwData := &genericWebhookData{}

	err = c.BindJSON(gwData)
	if err != nil {
		log.Printf("Failed to convert POST data into JSON: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Malformed POST data"})
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		log.Printf("Failed to read body: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Malformed body"})
		return
	}
	defer c.Request.Body.Close()

	go g.notifyGenericWebhookEvent(proj, body, gwData)
	c.JSON(200, gin.H{"status": "Success"})
}

func (g *genericWebhook) notifyGenericWebhookEvent(proj *brigade.Project, payload []byte, gwData *genericWebhookData) {
	if err := g.genericWebhookEvent(proj, payload, gwData); err != nil {
		log.Printf("failed genericWebhook event: %s", err)
	}
}

func (g *genericWebhook) genericWebhookEvent(proj *brigade.Project, payload []byte, gwData *genericWebhookData) error {
	revision := &brigade.Revision{}
	revision.Commit = gwData.Commit
	revision.Ref = gwData.Ref

	// get brigade.js
	script, err := GetFileContents(proj, revision.Ref, "brigade.js")
	if err != nil {
		log.Printf("Error getting file: %s", err)
		return fmt.Errorf("Error getting file: %s", err)
	}

	// create a Build for the specified Revision
	b := &brigade.Build{
		ProjectID: proj.ID,
		Type:      "webhook",
		Provider:  "GenericWebhook",
		Payload:   payload,
		Revision:  revision,
		Script:    script,
	}
	if proj.DefaultScript != "" {
		b.Script = []byte(proj.DefaultScript)
	}
	return g.store.CreateBuild(b)
}
