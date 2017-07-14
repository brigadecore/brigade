package webhook

import "testing"

func TestConfigureProject(t *testing.T) {
	data := map[string][]byte{
		"repository":   []byte("myrepo"),
		"sharedSecret": []byte("mysecret"),
		"githubToken":  []byte("like a fish needs a bicycle"),
		"sshKey":       []byte("hello$world"),
		"namespace":    []byte("zooropa"),
		"secrets":      []byte(`{"bar":"baz","foo":"bar"}`),
		// Intentionally skip cloneURL, test that this is ""
	}
	proj := &Project{Name: "acidTest"}
	if err := configureProject(proj, data, "defaultNS"); err != nil {
		t.Fatal(err)
	}

	if proj.CloneURL != "" {
		t.Errorf("Expected empty cloneURL, got %s", proj.CloneURL)
	}

	if proj.Repo != "myrepo" {
		t.Error("Repo is not correct")
	}
	if proj.SharedSecret != "mysecret" {
		t.Error("SharedSecret is not correct")
	}
	if proj.GitHubToken != "like a fish needs a bicycle" {
		t.Error("Fish cannot find its bicycle")
	}
	if proj.SSHKey != "hello\nworld" {
		t.Errorf("Unexpected SSHKey: %q", proj.SSHKey)
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
}

func TestDef(t *testing.T) {
	if got := def([]byte{}, "default"); got != "default" {
		t.Error("Expected default value")
	}
	if got := def([]byte("hello"), "world"); got != "hello" {
		t.Error("Expected non-default value")
	}
}
