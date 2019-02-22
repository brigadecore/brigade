package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"

	gin "gopkg.in/gin-gonic/gin.v1"
)

type genericWebhookSimpleEvent struct {
	store storage.Store
}

// NewGenericWebhookSimpleEvent creates a go-restful handler for generic Gateway.
func NewGenericWebhookSimpleEvent(s storage.Store) gin.HandlerFunc {
	h := &genericWebhookSimpleEvent{store: s}
	return h.Handle
}

// Handle handles a generic Gateway event.
func (g *genericWebhookSimpleEvent) Handle(c *gin.Context) {
	projectID := c.Param("projectID")
	secret := c.Param("secret")

	proj, err := g.store.GetProject(projectID)

	if err != nil {
		log.Printf("Project %q not found. No secret loaded. %s", projectID, err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "project not found"})
		return
	}

	err = validateGenericGatewaySecret(proj, secret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": err.Error()})
		return
	}

	payload, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Failed to read body: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Malformed body"})
		return
	}
	defer c.Request.Body.Close()

	revision := &brigade.Revision{}

	// try to unmarshal Revision data, if payload string is not empty
	if string(payload) != "" {
		err = json.Unmarshal(payload, &revision)
		if err != nil {
			log.Printf("Failed to convert POST data into JSON: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"status": "Malformed POST data - Invalid JSON"})
			return
		}
	}

	go g.notifyGenericWebhookSimpleEvent(proj, payload, revision)
	c.JSON(200, gin.H{"status": "Success. Build created"})
}

func (g *genericWebhookSimpleEvent) notifyGenericWebhookSimpleEvent(proj *brigade.Project, payload []byte, revision *brigade.Revision) {
	if err := g.genericWebhookSimpleEvent(proj, payload, revision); err != nil {
		log.Printf("failed genericWebhook SimpleEvent: %s", err)
	}
}

func (g *genericWebhookSimpleEvent) genericWebhookSimpleEvent(proj *brigade.Project, payload []byte, revision *brigade.Revision) error {
	b := &brigade.Build{
		ProjectID: proj.ID,
		Type:      "simpleevent",
		Provider:  "GenericWebhook",
		Payload:   payload,
		Revision:  revision,
	}

	// set a default Revision if user has not provided any information about commit or ref
	// otherwise, sidecar fails with 'fatal: empty string is not a valid pathspec. please use . instead if you meant to match all paths'
	// if the project has no VCS integration (e.g. the sidecar is set to 'NONE'), then this "master" will just be ignored by the worker
	if b.Revision == nil || (b.Revision.Commit == "" && b.Revision.Ref == "") {
		b.Revision = &brigade.Revision{Ref: "master"}
	}

	return g.store.CreateBuild(b)
}

// validateGenericGatewaySecret will return an error if given Project does not have a GenericGatewaySecret or if the provided secret is wrong
// Otherwise, it will simply return nil
func validateGenericGatewaySecret(proj *brigade.Project, secret string) error {
	// if the secret is "" (probably i) due to a Brigade upgrade or ii) user did not create a Generic Gateway secret during `brig project create`)
	// refuse to serve it, so Brigade admin will be forced to update the project with a non-empty secret
	if proj.GenericGatewaySecret == "" {
		log.Printf("Secret for project %s is empty, please update it and try again", proj.ID)
		return fmt.Errorf("secret for this Brigade Project is empty, refusing to serve, please inform your Brigade admin")
	}

	// compare secrets
	if secret != proj.GenericGatewaySecret {
		log.Printf("Secret %s for project %s is wrong", secret, proj.ID)
		return fmt.Errorf("secret is wrong")
	}

	return nil
}
