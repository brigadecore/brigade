package webhook

import "testing"

func TestConfigureProject(t *testing.T) {
	data := map[string][]byte{
		"repository":  []byte("myrepo"),
		"secret":      []byte("mysecret"),
		"githubToken": []byte("like a fish needs a bicycle"),
		"sshKey":      []byte("hello$world"),
		"namespace":   []byte("zooropa"),
		"env":         []byte(`{"bar":"baz","foo":"bar"}`),
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
	if proj.Secret != "mysecret" {
		t.Error("Secret is not correct")
	}
	if proj.GitHubToken != "like a fish needs a bicycle" {
		t.Error("Fish cannot find its bicycle")
	}
	if proj.SSHKey != "hello\nworld" {
		t.Errorf("Unexpected SSHKey: %q", proj.SSHKey)
	}
	if v, ok := proj.Env["bar"]; !ok {
		t.Error("Could not find key bar")
	} else if v != "baz" {
		t.Errorf("Expected baz, got %q", v)
	}
	if v, ok := proj.Env["NO SUCH KEY"]; ok {
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
