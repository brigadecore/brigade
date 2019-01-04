package webhook

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"

	restful "github.com/emicklei/go-restful"
)

type genericWebhook struct {
	store storage.Store
}

type genericWebhookData struct {
	Ref    string `json:"ref"`
	Commit string `json:"commit"`
}

// NewGenericWebhook creates a go-restful handler for generic webhook.
func NewGenericWebhook(s storage.Store) func(request *restful.Request, response *restful.Response) {
	h := &genericWebhook{store: s}
	return h.Handle
}

// Handle handles a generic webhook event.
func (g *genericWebhook) Handle(request *restful.Request, response *restful.Response) {
	log.Printf("Parameters: %s\n", request.PathParameters())

	projectID := request.PathParameter("projectID")
	secret := request.PathParameter("secret")

	proj, err := g.store.GetProject(projectID)

	if err != nil {
		log.Printf("Project %q not found. No secret loaded. %s", projectID, err)
		response.WriteErrorString(http.StatusBadRequest, "{\"status\": \"project not found\"}")
		return
	}

	// if the secret is "" (probably due to a Brigade upgrade)
	// refuse to serve it, so Brigade admin will be forced to update the project with a non-empty secret
	if proj.GenericWebhookSecret == "" {
		log.Printf("Secret for project %s is empty, please update it and try again", projectID)
		response.WriteErrorString(http.StatusUnauthorized, "{\"status\": \"secret is empty, please update it and try again\"}")
		return
	}

	// compare secrets
	if secret != proj.GenericWebhookSecret {
		log.Printf("Secret %s for project %s is wrong", secret, projectID)
		response.WriteErrorString(http.StatusUnauthorized, "{\"status\": \"secret is wrong\"}")
		return
	}

	gwData := &genericWebhookData{}
	err = request.ReadEntity(gwData)

	if err != nil {
		log.Printf("Failed to read GenericWebHookData: %s", err)
		response.WriteErrorString(http.StatusBadRequest, "{\"status\": \"Malformed genericWebhookData\"}")
		return
	}

	log.Printf("gwData: %#v\n", gwData)

	body, err := ioutil.ReadAll(request.Request.Body)
	if err != nil {
		log.Printf("Failed to read body: %s", err)
		response.WriteErrorString(http.StatusBadRequest, "{\"status\": \"Malformed body\"}")
		return
	}
	defer request.Request.Body.Close()

	go g.notifyGenericWebhookEvent(proj, body, gwData)
	response.Write([]byte("{\"status\": \"Success\"}"))
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
