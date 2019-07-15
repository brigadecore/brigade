package kube

import (
	"encoding/json"
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/brigadecore/brigade/pkg/brigade"
)

func TestGetProjects(t *testing.T) {
	k, s := fakeStore()
	createFakeProject(k, stubProjectSecret)
	projects, err := s.GetProjects()
	if err != nil {
		t.Fatal(err)
	}

	if len(projects) != 1 {
		t.Fatalf("expected one project, got %d", len(projects))
	}
}

func TestGetProject(t *testing.T) {
	k, s := fakeStore()
	createFakeProject(k, stubProjectSecret)
	proj, err := s.GetProject(stubProjectID)
	if err != nil {
		t.Fatal(err)
	}
	if proj.ID != stubProjectID {
		t.Error("Unexpected project ID: ", proj.ID)
	}
}

func TestCreateProject(t *testing.T) {
	k, s := fakeStore()
	secretsMap := map[string]string{"username": "hello", "password": "world"}
	n := "tennyson/light-brigade"
	proj := &brigade.Project{
		Name:         n,
		SharedSecret: "We Break for Seabeasts",
		Github: brigade.Github{
			Token:     "half-a-league",
			BaseURL:   "http://example.com",
			UploadURL: "http://up.example.com",
		},
		Kubernetes: brigade.Kubernetes{
			BuildStorageSize:  "50Mi",
			VCSSidecar:        "alpine:3.7",
			Namespace:         "brigade",
			BuildStorageClass: "3rdGrade",
			CacheStorageClass: "underwaterbasketweaving",
			ServiceAccount:    "project-sa",
		},
		DefaultScript:     "console.log('hi');",
		DefaultScriptName: "bernie",
		Repo: brigade.Repo{
			Name:     "git.example.com/tennyson/light-brigade",
			SSHKey:   "i know what you did last summer",
			CloneURL: "http://clown.example.com/clown.git",
		},
		Secrets: secretsMap,
		Worker: brigade.WorkerConfig{
			Registry:   "reggie",
			Name:       "bobby",
			Tag:        "millie",
			PullPolicy: "Always",
		},
		InitGitSubmodules:   true,
		AllowPrivilegedJobs: true,
		AllowHostMounts:     true,
		WorkerCommand:       "echo hello",
	}
	err := s.CreateProject(proj)
	if err != nil {
		t.Fatal(err)
	}

	id := brigade.ProjectID(n)
	secret, err := k.CoreV1().Secrets("default").Get(id, meta.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	expectedLabels := map[string]string{
		"app":       "brigade",
		"heritage":  "brigade",
		"component": "project",
	}
	for n, want := range expectedLabels {
		if got := secret.ObjectMeta.Labels[n]; got != want {
			t.Errorf("Expected %s to be %q, got %q", n, want, got)
		}
	}

	if pn := secret.ObjectMeta.Annotations["projectName"]; pn != n {
		t.Errorf("Expected %s for projectName, got %s", n, pn)
	}

	secretsJSON, err := json.Marshal(secretsMap)
	if err != nil {
		t.Fatal(err)
	}
	stringData := map[string]string{
		"sharedSecret":                 proj.SharedSecret,
		"github.token":                 proj.Github.Token,
		"github.baseURL":               proj.Github.BaseURL,
		"github.uploadURL":             proj.Github.UploadURL,
		"vcsSidecar":                   proj.Kubernetes.VCSSidecar,
		"namespace":                    proj.Kubernetes.Namespace,
		"serviceAccount":               proj.Kubernetes.ServiceAccount,
		"buildStorageSize":             proj.Kubernetes.BuildStorageSize,
		"kubernetes.cacheStorageClass": proj.Kubernetes.CacheStorageClass,
		"kubernetes.buildStorageClass": proj.Kubernetes.BuildStorageClass,
		"defaultScript":                proj.DefaultScript,
		"defaultScriptName":            proj.DefaultScriptName,
		"repository":                   proj.Repo.Name,
		"sshKey":                       proj.Repo.SSHKey,
		"cloneURL":                     proj.Repo.CloneURL,
		"secrets":                      string(secretsJSON),
		"worker.registry":              proj.Worker.Registry,
		"worker.name":                  proj.Worker.Name,
		"worker.tag":                   proj.Worker.Tag,
		"worker.pullPolicy":            proj.Worker.PullPolicy,
		"initGitSubmodules":            fmt.Sprintf("%t", proj.InitGitSubmodules),
		"imagePullSecrets":             proj.ImagePullSecrets,
		"allowPrivilegedJobs":          fmt.Sprintf("%t", proj.AllowPrivilegedJobs),
		"allowHostMounts":              fmt.Sprintf("%t", proj.AllowHostMounts),
		"workerCommand":                proj.WorkerCommand,
	}

	for key, want := range stringData {
		if got := secret.StringData[key]; got != want {
			t.Errorf("For key %s, got %q, want %q", key, got, want)
		}
	}
}

func TestReplaceProject(t *testing.T) {
	k, s := fakeStore()
	p := &brigade.Project{Name: "fakeName", ID: "brigade-fakeID", Github: brigade.Github{
		Token:     "half-a-league",
		BaseURL:   "http://example.com",
		UploadURL: "http://up.example.com",
	}}
	// create the project
	if err := s.CreateProject(p); err != nil {
		t.Fatal(err)
	}
	// make sure it's there
	if _, err := k.CoreV1().Secrets("default").Get("brigade-fakeID", meta.GetOptions{}); err != nil {
		t.Fatal(err)
	}
	// create another one with different ID
	pNonExistent := &brigade.Project{Name: "fakeName", ID: "brigade-fakeIDNotExistent"}
	// try to replace it, this will throw an error since it does not already exists
	if err := s.ReplaceProject(pNonExistent); err == nil {
		t.Fatal("Err should not be nil, since project does not exist")
	}
	// create another one, this time using same ID and a couple of added/changed properties
	p2 := &brigade.Project{Name: "fakeName", ID: "brigade-fakeID", DefaultScript: "new.js", Github: brigade.Github{
		Token:     "half-a-league2",
		BaseURL:   "http://example2.com",
		UploadURL: "http://up.example2.com",
	}}
	// replace it
	if err := s.ReplaceProject(p2); err != nil {
		t.Fatal(err)
	}
	// make sure it worked, get it as well
	updatedSecret, err := k.CoreV1().Secrets("default").Get("brigade-fakeID", meta.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	// check for new/added properties
	if updatedSecret.StringData["defaultScript"] != "new.js" {
		t.Fatalf("Wrong value in DefaultScript. It's %s whereas it should be 'new.js'", updatedSecret.StringData["defaultScript"])
	}

	if updatedSecret.StringData["github.baseURL"] != "http://example2.com" {
		t.Fatalf("Wrong value in Github.BaseURL. It's %s whereas it should be 'http://example2.com'", updatedSecret.StringData["github.baseURL"])
	}
}

func TestDeleteProject(t *testing.T) {
	k, s := fakeStore()
	p := &brigade.Project{ID: "fake", Name: "fake"}
	if err := s.CreateProject(p); err != nil {
		t.Fatal(err)
	}
	if _, err := k.CoreV1().Secrets("default").Get("fake", meta.GetOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteProject("fake"); err != nil {
		t.Fatal(err)
	}
}

func TestConfigureProject(t *testing.T) {
	secret := &v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name: "brigadeTest",
		},
		Type: secretTypeBuild,
		Data: map[string][]byte{
			"repository":        []byte("myrepo"),
			"defaultScript":     []byte(`console.log("hello default script")`),
			"defaultScriptName": []byte("global-cm-script"),
			"sharedSecret":      []byte("mysecret"),
			"github.token":      []byte("like a fish needs a bicycle"),
			"github.baseURL":    []byte("https://example.com/base"),
			"github.uploadURL":  []byte("https://example.com/upload"),
			"sshKey":            []byte("hello$world"),
			"namespace":         []byte("zooropa"),
			"secrets":           []byte(`{"bar":"baz","foo":"bar"}`),
			"worker.registry":   []byte("brigadecore"),
			"worker.name":       []byte("brigade-worker"),
			"worker.tag":        []byte("canary"),
			"worker.pullPolicy": []byte("Always"),
			// Intentionally skip cloneURL, test that this is ""
			"buildStorageSize":             []byte("50Mi"),
			"kubernetes.cacheStorageClass": []byte("hello"),
			"kubernetes.buildStorageClass": []byte("goodbye"),
			"allowPrivilegedJobs":          []byte("true"),
			// Default fo allowHostMounts is false. Testing that
			"initGitSubmodules": []byte("false"),
			"workerCommand":     []byte("echo hello"),
			"imagePullSecrets":  []byte("image pull secrets"),
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
	if proj.DefaultScriptName != "global-cm-script" {
		t.Errorf("Unexpected DefaultScriptName: %q", proj.DefaultScriptName)
	}
	if proj.SharedSecret != "mysecret" {
		t.Error("SharedSecret is not correct")
	}
	if proj.Github.Token != "like a fish needs a bicycle" {
		t.Error("Fish cannot find its bicycle")
	}
	if proj.Github.BaseURL != "https://example.com/base" {
		t.Errorf("Unexpected base URL: %s", proj.Github.BaseURL)
	}
	if proj.Github.UploadURL != "https://example.com/upload" {
		t.Errorf("Unexpected upload URL: %s", proj.Github.UploadURL)
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
	if proj.Worker.Registry != "brigadecore" {
		t.Fatalf("unexpected Worker.Registry: %s != brigadecore", proj.Worker.Registry)
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
	if proj.Kubernetes.BuildStorageSize != "50Mi" {
		t.Fatalf("buildStorageSize is wrong %s", proj.Kubernetes.BuildStorageSize)
	}
	if proj.Kubernetes.BuildStorageClass != "goodbye" {
		t.Errorf("buildStorageClass is wrong %s", proj.Kubernetes.BuildStorageClass)
	}
	if proj.Kubernetes.CacheStorageClass != "hello" {
		t.Errorf("cacheStorageClass is wrong")
	}
	if !proj.AllowPrivilegedJobs {
		t.Error("allowPrivilegedJobs should be true")
	}
	if proj.AllowHostMounts {
		t.Error("allowHostMounts should be false")
	}
	if proj.InitGitSubmodules {
		t.Error("initGitSubmodules should be false")
	}

	if proj.WorkerCommand != "echo hello" {
		t.Error("unexpected worker command")
	}

	if proj.ImagePullSecrets != "image pull secrets" {
		t.Error("unexpected image pull secrets")
	}
}

func TestDef(t *testing.T) {
	if got := def("", "default"); got != "default" {
		t.Error("Expected default value")
	}
	if got := def("hello", "world"); got != "hello" {
		t.Error("Expected non-default value")
	}
}
