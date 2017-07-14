package js

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestEvent(t *testing.T) {
	e := &Event{
		Type:     "push",
		Provider: "github",
		Commit:   "c0ff33c0ffeec0ff33c0ffee",
		Payload:  map[string]string{"hello": "world"},
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
		`myEvent.provider == "github"`,
		fmt.Sprintf("myEvent.commit == %q", e.Commit),
		`myEvent.payload.hello == "world"`,
	} {
		if err := sandbox.ExecString(script); err != nil {
			t.Fatalf("error executing %q: %s", script, err)
		}
	}

}
