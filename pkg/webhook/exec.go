package webhook

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/deis/acid/pkg/acid"
	"github.com/deis/acid/pkg/config"
	"github.com/deis/acid/pkg/js"

	"gopkg.in/gin-gonic/gin.v1"
)

type execHook struct {
	store store
}

func NewExecHook(s store) *execHook {
	return &execHook{
		store: s,
	}
}

// Handle takes an uploaded Acid script and some configuration and runs the script.
func (e *execHook) Handle(c *gin.Context) {
	signature := c.Request.Header.Get(hubSignature)
	namespace, _ := config.AcidNamespace(c)
	orgName := c.Param("org")
	projName := c.Param("project")
	// curl -X POST http://example.com/events/exec/technosophos/myproj/master
	// events.exec = function(e) {}
	commit := c.Param("commit")
	pname := fmt.Sprintf("%s/%s", orgName, projName)

	rawScript, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Failed to read body: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "Malformed body"})
		return
	}
	defer c.Request.Body.Close()

	proj, err := e.store.Get(pname, namespace)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "project not found"})
		return
	}

	if err := validateSignature(signature, proj.SharedSecret, []byte(rawScript)); err != nil {
		log.Printf("Error processing exec event: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"status": "authentication error"})
		return
	}

	c.JSON(executeScriptData(commit, rawScript, proj))
}

func executeScriptData(commit string, rawScript []byte, proj *acid.Project) (int, gin.H) {
	e := &js.Event{
		Type:     "exec",
		Provider: "client",
		Commit:   commit,
	}
	p := &js.Project{
		ID:   proj.ID, //"acid-" + storage.ShortSHA(proj.Repo),
		Name: proj.Name,
		Repo: js.Repo{
			Name:     proj.Repo.Name,
			CloneURL: proj.Repo.CloneURL,
			SSHKey:   strings.Replace(proj.Repo.SSHKey, "\n", "$", -1),
		},
		Payload: "",
		Kubernetes: js.Kubernetes{
			Namespace: proj.Kubernetes.Namespace,
			// By putting the sidecar image here, we are allowing an acid.js
			// to override it.
			VCSSidecar: proj.Kubernetes.VCSSidecar,
		},
		Secrets: proj.Secrets,
	}

	// Right now, we do this sychnronously since we have no backchannel.
	if err := js.HandleEvent(e, p, rawScript); err != nil {
		return http.StatusInternalServerError, gin.H{"error": err.Error()}
	}
	return 200, gin.H{"status": "completed"}
}
