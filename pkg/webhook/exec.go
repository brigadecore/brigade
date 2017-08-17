package webhook

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/deis/acid/pkg/acid"
	"github.com/deis/acid/pkg/config"

	"gopkg.in/gin-gonic/gin.v1"
)

type execHook struct {
	store store
}

func NewExecHook(s store) *execHook {
	return &execHook{s}
}

// Handle takes an uploaded Acid script and some configuration and runs the script.
func (e *execHook) Handle(c *gin.Context) {
	signature := c.Request.Header.Get(hubSignatureHeader)
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

	proj, err := e.store.GetProject(pname, namespace)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "project not found"})
		return
	}

	if err := validateSignature(signature, proj.SharedSecret, rawScript); err != nil {
		log.Printf("Error processing exec event: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"status": "authentication error"})
		return
	}

	c.JSON(e.executeScriptData(commit, rawScript, proj))
}

func (e *execHook) executeScriptData(commit string, script []byte, proj *acid.Project) (int, gin.H) {
	b := &acid.Build{
		Type:     "exec",
		Provider: "client",
		Commit:   commit,
		Script:   script,
	}

	// Right now, we do this sychnronously since we have no backchannel.
	if err := e.store.CreateBuild(b, proj); err != nil {
		return http.StatusInternalServerError, gin.H{"error": err.Error()}
	}
	return 200, gin.H{"status": "completed"}
}
