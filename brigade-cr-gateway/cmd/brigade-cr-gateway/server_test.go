package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Azure/brigade/pkg/storage/mock"
)

func TestNewRouter(t *testing.T) {
	s := mock.New()
	s.ProjectList[0].Name = "pequod/stubbs"
	r := newRouter(s)

	if r == nil {
		t.Fail()
	}

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("Unexpected status on healthz: %s", res.Status)
	}

	body, err := ioutil.ReadFile("./testdata/dockerhub-push.json")
	if err != nil {
		t.Fatal(err)
	}

	// Basically, we're testing to make sure the route exists, but having it bail
	// before it hits the GitHub API.
	routes := []string{
		"/events/webhook/brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac",
		"/events/webhook/deis/empty-testbed",
		"/events/webhook/deis/empty-testbed/master",
	}
	for _, r := range routes {
		res, err = http.Post(ts.URL+r, "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != 400 {
			t.Fatalf("Expected bad status, got: %s", res.Status)
		}
	}
}
