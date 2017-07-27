package webhook

import (
	"testing"

	"github.com/deis/acid/pkg/acid"
)

func TestExecuteScriptData(t *testing.T) {
	script := `
events.exec = function(e)  {
	console.log("Hello")
}
`
	proj := createExampleProject()
	commit := "1234567890abcdef"
	code, h := executeScriptData(commit, []byte(script), proj)
	if code != 200 {
		t.Fatalf("Expected 200 code, got %d", code)
	}
	if s := h["status"]; s != "completed" {
		t.Errorf("Expected completed, got %s", s)
	}
}

func createExampleProject() *acid.Project {
	return &acid.Project{
		ID:   "acid-org/proj",
		Name: "org/proj",
		Repo: acid.Repo{
			Name:     "github.com/org/proj",
			CloneURL: "https://example.com/git",
			SSHKey:   "SSHKEY",
		},
		Kubernetes: acid.Kubernetes{
			Namespace:  "foo",
			VCSSidecar: "foo:latest",
		},
	}
}
