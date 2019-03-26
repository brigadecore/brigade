package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brigadecore/brigade/pkg/storage/mock"
)

func TestNewRouter(t *testing.T) {
	t.Parallel()
	s := mock.New()
	s.ProjectList[0].ID = "brigade-4625a05cf6914e556aa254cb2af234203744de2f"
	s.ProjectList[0].Name = "brigadecore/empty-testbed"
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

	tests := []struct {
		testfile string
		route400 string
		route401 string
	}{
		{
			testfile: "./testdata/simpleevent.json",
			route400: "/simpleevents/v1/brigade-4625a05cf6914e556aa254cb2af234203744de2f_WRONG_URL/mysecret",
			route401: "/simpleevents/v1/brigade-4625a05cf6914e556aa254cb2af234203744de2f/mysecret2",
		},
		{
			testfile: "./testdata/cloudevent.json",
			route400: "/cloudevents/v02/brigade-4625a05cf6914e556aa254cb2af234203744de2f_WRONG_URL/mysecret",
			route401: "/cloudevents/v02/brigade-4625a05cf6914e556aa254cb2af234203744de2f/mysecret2",
		},
	}

	for _, test := range tests {
		body, err := ioutil.ReadFile(test.testfile)
		if err != nil {
			t.Fatal(err)
		}

		route400 := test.route400
		res, err = http.Post(ts.URL+route400, "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != 400 {
			t.Fatalf("Expected 400 status, got: %s", res.Status)
		}

		route401 := test.route401
		res, err = http.Post(ts.URL+route401, "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != 401 {
			t.Fatalf("Expected 401 status, got: %s", res.Status)
		}
	}

}
