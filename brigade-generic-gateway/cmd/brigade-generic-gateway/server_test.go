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
	s.ProjectList[0].ID = "brigade-4625a05cf6914e556aa254cb2af234203744de2f"
	s.ProjectList[0].Name = "deis/empty-testbed"
	s.ProjectList[0].GenericGatewaySecret = "mysecret"
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

	body, err := ioutil.ReadFile("./testdata/simpleevent.json")
	if err != nil {
		t.Fatal(err)
	}

	route400 := "/simpleevent/brigade-4625a05cf6914e556aa254cb2af234203744de2f_WRONG_URL/mysecret"
	res, err = http.Post(ts.URL+route400, "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 400 {
		t.Fatalf("Expected 400 status, got: %s", res.Status)
	}

	route401 := "/simpleevent/brigade-4625a05cf6914e556aa254cb2af234203744de2f/mysecret2"
	res, err = http.Post(ts.URL+route401, "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 401 {
		t.Fatalf("Expected 401 status, got: %s", res.Status)
	}
}
