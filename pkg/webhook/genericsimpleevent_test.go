package webhook

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/brigadecore/brigade/pkg/storage"
	"github.com/brigadecore/brigade/pkg/storage/mock"

	gin "gopkg.in/gin-gonic/gin.v1"
)

func newTestGenericWebhookSimpleEventHandler(store storage.Store) *genericWebhookSimpleEvent {
	return &genericWebhookSimpleEvent{store}
}

func newGenericProject() *brigade.Project {
	return &brigade.Project{
		ID:   "brigade-1234",
		Name: "dgkanatsios/o365-notify",
		Repo: brigade.Repo{
			Name:     "github.com/dgkanatsios/o365-notify",
			CloneURL: "https://github.com/dgkanatsios/o365-notify",
		},
		Secrets: map[string]string{
			"mysecret": "value",
		},
		DefaultScript: `const { events, Job } = require("brigadier")
		events.on("exec", () => {
		  var job = new Job("do-nothing", "alpine:3.8")
		  job.tasks = [
			"echo Hello",
			"echo World"
		  ]
		  job.run()
		})`,
	}
}

func TestGenericWebhookSimpleEventHandler(t *testing.T) {
	proj := newGenericProject()
	store := newTestStore()
	h := newTestGenericWebhookSimpleEventHandler(store)

	revision := &brigade.Revision{
		Commit: "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28",
	}

	if err := h.genericWebhookSimpleEvent(proj, []byte(exampleSimpleEvent), revision); err != nil {
		t.Errorf("failed generic gateway event: %s", err)
	}

	if payload := string(store.builds[0].Payload); payload != exampleSimpleEvent {
		t.Errorf("unexpected payload: %s", payload)
	}

	if provider := string(store.builds[0].Provider); provider != "GenericWebhook" {
		t.Errorf("unexpected provider: %s", provider)
	}

	if etype := string(store.builds[0].Type); etype != "simpleevent" {
		t.Errorf("unexpected type: %s", etype)
	}
}

func TestGenericWebHookSimpleEvent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		description    string
		url            string
		statusExpected int
		store          *mock.Store
		payload        string
		revision       *brigade.Revision
	}{
		{
			description:    "Wrong project",
			url:            "/simpleevents/v1/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusBadRequest,
			store:          newTestStoreWithFakeProject2(),
			payload:        exampleSimpleEvent,
		},
		{
			description:    "Correct project, empty secret",
			url:            "/simpleevents/v1/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusUnauthorized,
			store:          newTestStoreWithFakeProject(),
			payload:        exampleSimpleEvent,
		},
		{
			description:    "Both ref and commit in JSON payload",
			url:            "/simpleevents/v1/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusOK,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload:        exampleSimpleEvent,
			revision: &brigade.Revision{
				Ref:    "refs/heads/changes",
				Commit: "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28",
			},
		},
		{
			description:    "Only ref in JSON payload",
			url:            "/simpleevents/v1/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusOK,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload:        `{"ref": "refs/heads/changes"}`,
			revision: &brigade.Revision{
				Ref: "refs/heads/changes",
			},
		},
		{
			description:    "Only commit in JSON payload",
			url:            "/simpleevents/v1/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusOK,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload:        `{"commit": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"}`,
			revision: &brigade.Revision{
				Commit: "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28",
			},
		},
		{
			description:    "Random values in JSON payload",
			url:            "/simpleevents/v1/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusOK,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload:        `{"val1": "refs/heads/changes", "val2": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"}`,
			revision: &brigade.Revision{
				Ref: "master",
			},
		},
		{
			description:    "Corrupt values in JSON payload",
			url:            "/simpleevents/v1/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusBadRequest,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload:        `{"val1": "refs/heads/changes", "val2": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28" CORRUPT}`,
			revision:       nil,
		},
		{
			description:    "Empty POST data",
			url:            "/simpleevents/v1/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusOK,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload:        ``,
			revision:       &brigade.Revision{Ref: "master"},
		},
		{
			description:    "POST data is an empty JSON object",
			url:            "/simpleevents/v1/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusOK,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload:        `{}`,
			revision:       &brigade.Revision{Ref: "master"},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			router := newMockRouterSimpleEvent(test.store)
			httpRequest := httptest.NewRequest("POST", test.url, bytes.NewBuffer([]byte(test.payload)))
			httpRequest.Header.Add("Content-Type", "application/json")
			rw := httptest.NewRecorder()
			router.ServeHTTP(rw, httpRequest)
			if rw.Result().StatusCode != test.statusExpected {
				t.Errorf("expected error %d, got %d", test.statusExpected, rw.Result().StatusCode)
			}

			// we got a 200, so let's make sure we got a proper Build created
			if rw.Result().StatusCode == http.StatusOK {
				checkBuild(t, test.store, test.revision.Ref, test.revision.Commit, []byte(test.payload))
			}
		})
	}
}

func checkBuild(t *testing.T, store *mock.Store, expectedRef string, expectedCommit string, payload []byte) {
	// timeout check in the method is necessary because handler ultimately runs in a goroutine
	// we might get rid of this as soon as we switch to synchronous handlers
	c := make(chan struct{})
	stopChan := make(chan struct{})

	go func() {
		for {
			select {
			default:
				if len(store.Builds) == 0 {
					time.Sleep(50 * time.Millisecond)
				} else {
					c <- struct{}{} // signal that we do have a Build
					break
				}
			case <-stopChan: // calling goroutine signals that we should exit, so return
				return
			}
		}
	}()

	select {
	case <-c: // we do have a Build, so exit select and continue on checking the Builds
		break
	case <-time.After(3 * time.Second):
		t.Errorf("No new Builds were created, expectedRef %s and expectedCommit %s", expectedRef, expectedCommit)
		stopChan <- struct{}{} // signal that infinite loop goroutine should be stopped
		return
	}

	build := store.Builds[0]
	if build.Revision.Ref != expectedRef || build.Revision.Commit != expectedCommit {
		t.Errorf("Wrong Revision, expected %#v got %#v", exampleSimpleEvent, build.Revision)
	}

	if !bytes.Equal(build.Payload, payload) {
		t.Errorf("Wrong payload, expected %s got %s", string(payload), string(build.Payload))
	}
}

const exampleSimpleEvent = `
{
	"ref": "refs/heads/changes",
	"commit": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"
}
`

func newMockRouterSimpleEvent(store storage.Store) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	handler := NewGenericWebhookSimpleEvent(store)

	events := router.Group("/simpleevents/v1")
	events.Use(gin.Logger())
	events.POST("/:projectID/:secret", handler)

	return router
}

func newTestStoreWithFakeProject() *mock.Store {
	return &mock.Store{
		ProjectList: []*brigade.Project{{
			ID: "brigade-fakeProject",
		}},
	}
}

func newTestStoreWithFakeProjectAndSecret(secret string) *mock.Store {
	return &mock.Store{
		ProjectList: []*brigade.Project{{
			ID:                   "brigade-fakeProject",
			GenericGatewaySecret: secret,
		}},
	}
}

func newTestStoreWithFakeProject2() *mock.Store {
	return &mock.Store{
		ProjectList: []*brigade.Project{{
			ID: "brigade-fakeProject2",
		}},
	}
}
