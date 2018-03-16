// +build integration

package tests

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-github/github"
	"k8s.io/api/core/v1"

	"github.com/Azure/brigade/pkg/storage/kube"
	"github.com/Azure/brigade/pkg/webhook"
)

var (
	kubeconfig string
	namespace  string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&namespace, "namespace", os.Getenv("BRIGADE_NAMESPACE"), "kubernetes namespace")
}

func generate() (payload []byte, hmac string) {
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "required arg: Git SHA")
		os.Exit(1)
	}
	commit := flag.Arg(0)

	eventType := "push"

	data, err := ioutil.ReadFile("./testdata/test-repo-push.json")
	if err != nil {
		panic(err)
	}

	event, err := github.ParseWebHook(eventType, data)
	if err != nil {
		panic(err)
	}

	var repo string

	switch event := event.(type) {
	case *github.PushEvent:
		event.HeadCommit.ID = github.String(commit)
		repo = event.Repo.GetFullName()
	case *github.PullRequestEvent:
		event.PullRequest.Head.SHA = github.String(commit)
		repo = event.Repo.GetFullName()
	}

	out, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		panic(err)
	}

	clientset, err := kube.GetClient("", kubeconfig)
	if err != nil {
		panic(err)
	}

	if namespace == "" {
		namespace = v1.NamespaceDefault
	}

	proj, err := kube.New(clientset, namespace).GetProject(repo)
	if err != nil {
		panic(err)
	}
	hmac = webhook.SHA1HMAC([]byte(proj.SharedSecret), out)
	return out, hmac
}

func TestFunctional(t *testing.T) {
	payload, hmac := generate()

	requests := []*http.Request{{
		Method: "POST",
		URL:    &url.URL{Scheme: "http", Host: "localhost:7744", Path: "/events/github"},
		Body:   ioutil.NopCloser(bytes.NewReader(payload)),
		Header: http.Header{
			"X-Github-Event":  []string{"push"},
			"X-Hub-Signature": []string{hmac},
		},
	}}

	for _, request := range requests {
		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Error(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s %s: expected status code '200', got '%d'\n", request.Method, request.URL.String(), resp.StatusCode)
		}
	}
}
