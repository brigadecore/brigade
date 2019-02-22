package webhook

import (
	"encoding/json"
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

	// try to unmarshal Revision data, if they do exist
	revision := &brigade.Revision{}
	err = json.Unmarshal(payload, &revision)
	if err != nil {
		log.Printf("Failed to convert POST data into JSON: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Malformed POST data - Invalid JSON"})
		return
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
	if b.Revision == nil || (b.Revision.Commit == "" && b.Revision.Ref == "") {
		b.Revision = &brigade.Revision{Ref: "master"}
	}

	return g.store.CreateBuild(b)
}
