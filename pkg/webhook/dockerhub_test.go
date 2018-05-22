package webhook

import (
	"testing"

	"github.com/Azure/brigade/pkg/brigade"
)

func newProject() *brigade.Project {
	return &brigade.Project{
		ID:   "brigade-1234",
		Name: "org/proj",
		Repo: brigade.Repo{
			Name:     "example.com/org/proj",
			CloneURL: "http://example.com/org/project.git",
		},
		Kubernetes: brigade.Kubernetes{
			Namespace:        "namespace",
			VCSSidecar:       "sidecar:latest",
			BuildStorageSize: "50Mi",
		},
		Secrets: map[string]string{
			"mysecret": "value",
		},
	}
}

func TestDoDockerImagePush(t *testing.T) {
	proj := newProject()

	commit := "e1e10"
	store := &testStore{}
	hook := &dockerPushHook{
		store: store,
	}

	if err := hook.doDockerImagePush(proj, commit, []byte(exampleWebhook)); err != nil {
		t.Errorf("failed docker image push: %s", err)
	}
	script := string(store.builds[0].Script)
	if script != "" {
		t.Errorf("unexpected build script: %s", script)
	}
}

func TestDoDockerImagePush_WithDefaultScript(t *testing.T) {
	proj := newProject()
	proj.DefaultScript = `console.log("hello default script")`

	commit := "e1e10"
	store := &testStore{}
	hook := &dockerPushHook{store: store}

	if err := hook.doDockerImagePush(proj, commit, []byte(exampleWebhook)); err != nil {
		t.Errorf("failed docker image push: %s", err)
	}
	script := string(store.builds[0].Script)
	if script != proj.DefaultScript {
		t.Errorf("unexpected build script: %s", script)
	}
}

const exampleWebhook = `
{
  "callback_url": "https://registry.hub.docker.com/u/svendowideit/testhook/hook/2141b5bi5i5b02bec211i4eeih0242eg11000a/",
  "push_data": {
    "images": [
        "27d47432a69bca5f2700e4dff7de0388ed65f9d3fb1ec645e2bc24c223dc1cc3",
        "51a9c7c1f8bb2fa19bcd09789a34e63f35abb80044bc10196e304f6634cc582c",
        "..."
    ],
    "pushed_at": 1.417566161e+09,
    "pusher": "trustedbuilder",
    "tag": "latest"
  },
  "repository": {
    "comment_count": "0",
    "date_created": 1.417494799e+09,
    "description": "",
    "dockerfile": "FROM scratch",
    "full_description": "Docker Hub based automated build from a GitHub repo",
    "is_official": false,
    "is_private": true,
    "is_trusted": true,
    "name": "testhook",
    "namespace": "svendowideit",
    "owner": "svendowideit",
    "repo_name": "svendowideit/testhook",
    "repo_url": "https://registry.hub.docker.com/u/svendowideit/testhook/",
    "star_count": 0,
    "status": "Active"
  }
}
`
