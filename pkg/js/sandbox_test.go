package js

import (
	"io/ioutil"
	"testing"
)

func TestSandbox_ExecAll(t *testing.T) {
	globals := map[string]interface{}{
		"github": map[string]string{
			"org":     "deis",
			"project": "acid",
		},
	}
	script1 := []byte(`console.log(github.org + "/" + github.project)`)
	script2 := []byte(`console.log(github.project+ "/" + github.org)`)

	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range globals {
		s.Variable(k, v)
	}

	if err := s.ExecAll(script1, script2); err != nil {
		t.Fatalf("JavaScript failed to execute with error %s", err)
	}
}

func TestSandbox_ExecString(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if err := s.ExecString(`console.log("hello there")`); err != nil {
		t.Fatal(err)
	}

	if err := s.ExecString(`invalid.f(`); err == nil {
		t.Fatal("exepected invalid JS to produce an error")
	}
}

func TestHandleEvent(t *testing.T) {
	acidjs := `
	exports.fired = false
	events.push = function(e) {
		if (e.repo.name != "owner/repo") {
			throw "expected owner/repo, got " + e.repo.name
		}
		exports.fired = true
	}
	`
	e := createTestEvent()
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	s.LoadPrecompiled("js/mock8s.js")

	if err := s.HandleEvent(e, []byte(acidjs)); err != nil {
		t.Fatal(err)
	}

	checkRun := `if (!exports.fired) { throw "Expected event to fire" }`
	if err := s.ExecString(checkRun); err != nil {
		t.Fatal(err)
	}
}

func TestHandleEvent_JS(t *testing.T) {
	e := createTestEvent()

	tests := []struct {
		name   string
		script []byte
		sshKey string
		fail   bool
	}{
		{"log", []byte(`events.push = function() {console.log("hello") }`), "foo", false},
		{"log", []byte(`events.push = function(e) {console.log(e.sshKey) }`), "foo", false},
		{"empty", mustReadScript(t, "testdata/empty_event.js"), "", false},
		{"empty_tasks", mustReadScript(t, "testdata/empty_tasks.js"), "", false},
		{"basic", mustReadScript(t, "testdata/job_no_sshkey.js"), "", false},
		{"with-sshkey", mustReadScript(t, "testdata/job_sshkey.js"), "my-ssh-key", false},
		{"waitgroup", mustReadScript(t, "testdata/waitgroup.js"), "", false},
		{"with-secrets", mustReadScript(t, "testdata/job_secrets.js"), "", false},
	}

	for _, tt := range tests {
		t.Logf("Running %s", tt.name)
		s, err := New()
		if err != nil {
			t.Fatal(err)
		}

		// Load a kubernetes mock
		s.LoadPrecompiled("js/mock8s.js")

		if err := s.HandleEvent(e, tt.script); err != nil {
			if tt.fail {
				continue
			}
			t.Fatalf("Script %s failed with : %s", tt.name, err)
		} else if tt.fail {
			t.Errorf("Expected test %s to fail.", tt.name)
		}
	}
}

func mustReadScript(t *testing.T, filename string) []byte {
	script, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	return script
}

func createTestEvent() *Event {
	ref := "c0ff334411"
	ph := &testPushHook{
		Ref: ref,
		Repository: testRepository{
			Name:        "repo",
			FullName:    "owner/repo",
			Description: "Test fixture",
			CloneURL:    "https://example.com/clone",
		},
		HeadCommit: testCommit{
			Id: ref,
		},
	}
	e := &Event{
		Type:      "push",
		Provider:  "github",
		Commit:    ph.Ref,
		Payload:   ph,
		ProjectID: "acid-c0ff33c0ffee",
		Repo: Repo{
			Name:     ph.Repository.FullName,
			CloneURL: ph.Repository.CloneURL,
			SSHKey:   "my voice is my passport. Verify me.",
		},
		Kubernetes: Kubernetes{
			Namespace:  "pandas",
			VCSSidecar: "mySidecar:latest",
		},
	}
	return e
}

// The following structs are vaguely reflective of GitHub's push hook

type testPushHook struct {
	Ref        string                 `json:"ref"`
	Repository testRepository         `json:"repository"`
	HeadCommit testCommit             `json:"head_commit"`
	Sender     map[string]interface{} `json:"sender"`
}

type testCommit struct {
	Id string `json:"id"`
}

type testRepository struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	CloneURL    string `json:"clone_url"`
}
