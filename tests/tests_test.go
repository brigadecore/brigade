// +build integration

package tests

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
)

func TestFunctional(t *testing.T) {

	host := os.Getenv("BRIGADE_BRIGADE_GITHUB_GW_SERVICE_HOST")
	if host == "" {
		host = "localhost"
	}

	githubPushFile, err := os.Open("testdata/test-repo-generated.json")
	if err != nil {
		t.Fatal(err)
	}
	defer githubPushFile.Close()
	hubSignature, err := ioutil.ReadFile("testdata/test-repo-generated.hash")
	if err != nil {
		t.Fatal(err)
	}
	requests := []*http.Request{
		{
			Method: "POST",
			URL:    &url.URL{Scheme: "http", Host: host + ":7744", Path: "/events/github"},
			Body:   githubPushFile,
			Header: http.Header{
				"X-Github-Event":  []string{"push"},
				"X-Hub-Signature": []string{string(hubSignature)},
			},
		},
	}

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
