package webhook

import (
	"testing"

	"github.com/deis/acid/pkg/acid"
)

func TestDoDockerImagePush(t *testing.T) {
	script := `events.dockerhub = function(e) {
		if (e.payload.push_data.tag == "latest") {
			throw "Unexpected test: " + e.payload.push_data.tag
		}
	}`
	proj := &acid.Project{
		ID:   "acid-1234",
		Name: "org/proj",
		Repo: acid.Repo{
			Name:     "example.com/org/proj",
			CloneURL: "http://example.com/org/project.git",
		},
		Kubernetes: acid.Kubernetes{
			Namespace:  "namespace",
			VCSSidecar: "sidecar:latest",
		},
		Secrets: map[string]string{
			"mysecret": "value",
		},
	}

	commit := "e1e10"

	if err := doDockerImagePush([]byte(exampleWebhook), proj, commit, []byte(script)); err != nil {
		t.Errorf("failed docker image push: %s", err)
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
