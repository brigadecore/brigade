package js

import (
	"encoding/json"
	"testing"
)

func TestEvent(t *testing.T) {
	e := &Event{
		Type:      "push",
		Provider:  "github",
		Commit:    "c0ff33c0ffeec0ff33c0ffee",
		Payload:   map[string]string{"hello": "world"},
		ProjectID: "acid-c0ff33c0ffee",
		Repo: Repo{
			Name:     "technosophos/coffee",
			CloneURL: "https://example.com/coffee.git",
			SSHKey:   "my voice is my passport. Verify me.",
		},
		Kubernetes: Kubernetes{
			Namespace: "frenchpress	",
		},
	}

	obj, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}

	sandbox, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, script := range []string{
		"var myEvent = " + string(obj),
		`myEvent.type == "push"`,
		`myEvent.repo.cloneURL == "https://example.com/coffee.git"`,
		`myEvent.kubernetes.namespace == "frenchpress"`,
	} {
		if err := sandbox.ExecString(script); err != nil {
			t.Fatalf("error executing %q: %s", script, err)
		}
	}

}
