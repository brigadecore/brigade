package worker

import (
	"regexp"
	"testing"

	"github.com/deis/acid/pkg/worker/workertest"
)

// Canary for MockExecutor. We put this here to avoid circular dependency.
var _ Executor = &workertest.MockExecutor{}

func TestHandleEvent(t *testing.T) {
	exec := &workertest.MockExecutor{}
	DefaultExecutor = exec

	e := &Event{
		Type:     "push",
		Provider: "github",
		Commit:   "c0ff33c0ffeec0ff33c0ffee",
	}
	p := &Project{
		ID:   "acid-11111111111111111111111111111111",
		Name: "github.com/technosophos/coffeesnob",
		Repo: Repo{
			Name:     "technosophos/coffeesnob",
			CloneURL: "https://github.com/technosophos/coffeesnob.git",
		},
		Kubernetes: Kubernetes{
			Namespace:  "acid-builds",
			VCSSidecar: "acid-sidecar:latest",
		},
		Secrets: map[string]string{
			"db_admin": "Ishmael",
		},
	}
	script := []byte(`console.log("Yay")`)

	if err := HandleEvent(e, p, script); err != nil {
		t.Fatal(err)
	}

	pod := exec.LastPod

	// TODO: Add more tests here
	if belongs := pod.Labels["belongsto"]; belongs != "technosophos-coffeesnob" {
		t.Errorf("Unexpected belongsto: %s", belongs)
	}

	jMatch := regexp.MustCompile("acid-worker-[0-9]+-[a-f0-9]{8}")
	if !jMatch.MatchString(pod.Name) {
		t.Errorf("pod name did not match regexp: %q", pod.Name)
	}
	if !jMatch.MatchString(pod.Labels["jobname"]) {
		t.Errorf("jobname did not match regexp: %q", pod.Labels["jobname"])
	}
	if pod.Labels["commit"] != e.Commit {
		t.Errorf("unexpected commit: %q", pod.Labels["commit"])
	}
}
