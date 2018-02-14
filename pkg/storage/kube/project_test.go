package kube

import (
	"testing"

	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConfigureProject(t *testing.T) {
	secret := &v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: "brigadeTest",
		},
		Type: secretTypeBuild,
		Data: map[string][]byte{
			"repository":        []byte("myrepo"),
			"defaultScript":     []byte(`console.log("hello default script")`),
			"sharedSecret":      []byte("mysecret"),
			"github.token":      []byte("like a fish needs a bicycle"),
			"sshKey":            []byte("hello$world"),
			"namespace":         []byte("zooropa"),
			"secrets":           []byte(`{"bar":"baz","foo":"bar"}`),
			"worker.registry":   []byte("deis"),
			"worker.name":       []byte("brigade-worker"),
			"worker.tag":        []byte("canary"),
			"worker.pullPolicy": []byte("Always"),
			// Intentionally skip cloneURL, test that this is ""
		},
	}

	proj, err := NewProjectFromSecret(secret, "defaultNS")
	if err != nil {
		t.Fatal(err)
	}

	if proj.ID != "brigadeTest" {
		t.Error("ID is not correct")
	}
	if proj.Repo.CloneURL != "" {
		t.Errorf("Expected empty cloneURL, got %s", proj.Repo.CloneURL)
	}
	if proj.Repo.Name != "myrepo" {
		t.Error("Repo is not correct")
	}
	if proj.DefaultScript != `console.log("hello default script")` {
		t.Errorf("Unexpected DefaultScript: %q", proj.DefaultScript)
	}
	if proj.SharedSecret != "mysecret" {
		t.Error("SharedSecret is not correct")
	}
	if proj.Github.Token != "like a fish needs a bicycle" {
		t.Error("Fish cannot find its bicycle")
	}
	if proj.Repo.SSHKey != "hello\nworld" {
		t.Errorf("Unexpected SSHKey: %q", proj.Repo.SSHKey)
	}
	if v, ok := proj.Secrets["bar"]; !ok {
		t.Error("Could not find key bar in Secrets")
	} else if v != "baz" {
		t.Errorf("Expected baz, got %q", v)
	}
	if v, ok := proj.Secrets["NO SUCH KEY"]; ok {
		t.Fatal("unexpected key")
	} else if v != "" {
		t.Fatal("Expected empty string for non-existent key")
	}
	if proj.Worker.Registry != "deis" {
		t.Fatalf("unexpected Worker.Registry: %s != deis", proj.Worker.Registry)
	}
	if proj.Worker.Name != "brigade-worker" {
		t.Fatalf("unexpected Worker.Name: %s != brigade-worker", proj.Worker.Name)
	}
	if proj.Worker.Tag != "canary" {
		t.Fatalf("unexpected Worker.Tag: %s != canary", proj.Worker.Tag)
	}
	if proj.Worker.PullPolicy != "Always" {
		t.Fatalf("unexpected Worker.PullPolicy: %s != Always", proj.Worker.PullPolicy)
	}
}

func TestDef(t *testing.T) {
	if got := def([]byte{}, "default"); got != "default" {
		t.Error("Expected default value")
	}
	if got := def([]byte("hello"), "world"); got != "hello" {
		t.Error("Expected non-default value")
	}
}
