package commands

import (
	"os"
	"reflect"
	"testing"

	"github.com/brigadecore/brigade/pkg/brigade"
)

const testProjectSecret = "./testdata/project_secret.json"

func TestInitProjectVCS(t *testing.T) {
	p := newProjectVCS()
	if p.Name != defaultProjectVCS.Name {
		t.Fatal("newProject is not cloning default project")
	}

	p.Name = "overrideName"
	if p.Name == defaultProjectVCS.Name {
		t.Fatal("newProject returned the pointer to the default project.")
	}
}

func TestInitProjectNoVCS(t *testing.T) {
	p := newProjectVCS()     // VCS is the default
	setDefaultValuesNoVCS(p) // switch the defaults to a no VCS project

	if p.Kubernetes.VCSSidecar != "NONE" {
		t.Fatal("VCSSidecar should be NONE")
	}

	if reflect.DeepEqual(p.Repo, brigade.Repo{}) {
		t.Fatal("Project Repo should be empty/unset")
	}
}

func TestParseSecret(t *testing.T) {
	f, err := os.Open(testProjectSecret)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sec, err := parseSecret(f)
	if err != nil {
		t.Fatal(err)
	}
	expect := "brigade-407900363c01e6153bc1a91792055b898e20a29f1387b72a0b6f00"
	if sec.Name != expect {
		t.Fatalf("Expected name %s, got %s", expect, sec.Name)
	}
}

func TestLoadProjectConfig(t *testing.T) {
	proj, err := loadProjectConfig(testProjectSecret, newProjectVCS())
	if err != nil {
		t.Fatal(err)
	}

	// We just spot-check a few values. The kube package tests every field.
	if proj.Name != "technosophos/-whale-eyes-" {
		t.Error("Expected project name to be whale eyes")
	}
	if proj.Kubernetes.BuildStorageSize != "50Mi" {
		t.Error("Expected Kubernetes BuilStorageSize to be 50Mi")
	}

	if proj.Github.Token != "not with a bang but a whimper" {
		t.Errorf("Expected Github secret to be set")
	}

	if proj.Worker.PullPolicy != "Always" {
		t.Errorf("expected worker pull policy to be Always.")
	}
}

func TestLoadFileValidator(t *testing.T) {
	if err := loadFileValidator(testProjectSecret); err != nil {
		t.Fatal(err)
	}
	if err := loadFileValidator("sir/not/appearing/in/this/film"); err == nil {
		t.Fatal("expected load of non-existent file to produce an eror")
	}
}

func TestLoadFileStr(t *testing.T) {
	if data := loadFileStr(testProjectSecret); data == "" {
		t.Fatal("Data should have been loaded")
	}
	if data := loadFileStr("sir/not/appearing"); len(data) > 0 {
		t.Fatal("Expected empty string for nonexistent file")
	}
}

func TestIsHTTP(t *testing.T) {
	tests := map[string]bool{
		"http://foo.bar":    true,
		"https://foo.bar":   true,
		"http@foo.bar":      false,
		"":                  false,
		"HTTP://foo.bar":    true,
		"git@foo.bar":       false,
		"ssh://git@foo.bar": false,
	}

	for url, expect := range tests {
		if isHTTP(url) != expect {
			t.Errorf("Unexpected result for %q", url)
		}
	}
}

func TestReplaceNewlines(t *testing.T) {
	given := "foo\nbar\nbaz"
	expect := "foo$bar$baz"
	if got := replaceNewlines(given); got != expect {
		t.Fatalf("Expected %q, got %q", expect, got)
	}
}

func TestGenericGatewaySecretValidator(t *testing.T) {
	s1 := "asdf"
	if err := genericGatewaySecretValidator(s1); err != nil {
		t.Fatal("Expected nil, got error")
	}
	s2 := "AsDf1"
	if err := genericGatewaySecretValidator(s2); err != nil {
		t.Fatal("Expected nil, got error")
	}
	s3 := "jfdkfd^&"
	if err := genericGatewaySecretValidator(s3); err == nil {
		t.Fatal("Expected error, got nil")
	}
	s4 := ""
	if err := genericGatewaySecretValidator(s4); err != nil {
		t.Fatal("Expected nil, got error")
	}
	s5 := " "
	if err := genericGatewaySecretValidator(s5); err == nil {
		t.Fatal("Expected error, got nil")
	}
}
