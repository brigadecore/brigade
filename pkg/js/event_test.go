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
			SSHURL:   "ssh://foo@example.com/coffee.git",
			GitURL:   "git://foo@example.com/coffee.git",
			SSHKey:   "my voice is my passport. Verify me.",
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
		"console.log(myEvent.type)",
		"console.log(myEvent.repo.sshURL)",
	} {
		if err := sandbox.ExecString(script); err != nil {
			t.Fatalf("error executing %q: %s", script, err)
		}
	}

}
