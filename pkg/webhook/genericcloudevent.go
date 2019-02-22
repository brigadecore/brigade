package webhook

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"

	gin "gopkg.in/gin-gonic/gin.v1"

	cloudevents "github.com/cloudevents/sdk-go/v02"
)

type genericWebhookCloudEvent struct {
	store storage.Store
}

// NewGenericWebhookCloudEvent creates a go-restful handler for generic Gateway that will handle CloudEvents.
func NewGenericWebhookCloudEvent(s storage.Store) gin.HandlerFunc {
	h := &genericWebhookCloudEvent{store: s}
	return h.Handle
}

// Handle handles a generic Gateway CloudEvent.
func (g *genericWebhookCloudEvent) Handle(c *gin.Context) {
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

	event := &cloudevents.Event{}

	err = json.Unmarshal(payload, &event)
	if err != nil {
		log.Printf("Failed to convert POST data into JSON: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Malformed POST data - Invalid JSON"})
		return
	}

	// CloudEvents required fields are type, specversion, source, id
	// as per https://github.com/cloudevents/spec/blob/v0.2/spec.md
	if event.ID == "" || event.Type == "" || event.SpecVersion == "" || event.Source.String() == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "CloudEvent should have non empty type, specversion, source, id"})
		return
	}

	// only support 0.2 of the CloudEvent spec for now
	if event.SpecVersion != "0.2" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "Brigade supports only '0.2' as CloudEvent specversion"})
		return
	}

	go g.notifyGenericWebhookCloudEvent(proj, payload, event)
	c.JSON(200, gin.H{"status": "Success"})
}

func (g *genericWebhookCloudEvent) notifyGenericWebhookCloudEvent(proj *brigade.Project, payload []byte, event *cloudevents.Event) {
	if err := g.genericWebhookCloudEvent(proj, payload, event); err != nil {
		log.Printf("failed genericWebhook Cloud Event: %s", err)
	}
}

func (g *genericWebhookCloudEvent) genericWebhookCloudEvent(proj *brigade.Project, payload []byte, event *cloudevents.Event) error {
	var revision brigade.Revision
	if event.Data != nil {
		data := event.Data.(map[string]interface{})
		if data["ref"] != nil {
			revision.Ref, _ = data["ref"].(string)
		}
		if data["commit"] != nil {
			revision.Commit, _ = data["commit"].(string)
		}
	}

	// set a default Revision if user has not provided any information about commit or ref
	// otherwise, sidecar fails with 'fatal: empty string is not a valid pathspec. please use . instead if you meant to match all paths'
	if revision.Commit == "" && revision.Ref == "" {
		revision.Ref = "master"
	}

	// create a Build for the specified Revision
	b := &brigade.Build{
		ProjectID: proj.ID,
		Type:      "cloudevent",
		Provider:  "GenericWebhook",
		Payload:   payload,
		Revision:  &revision,
	}

	return g.store.CreateBuild(b)
}
