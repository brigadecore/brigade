package webhook

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
	"github.com/Azure/brigade/pkg/storage/mock"

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

func TestGenericWebhook(t *testing.T) {
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

func TestGenericWebhookHandlerBadBody(t *testing.T) {
	store := newEmptyTestStore()
	router := newMockRouterSimpleEvent(store)

	httpRequest := httptest.NewRequest("POST", "/simpleevent/fakeProject/fakeCode", bytes.NewBuffer(nil))
	httpRequest.Header.Add("Content-Type", "application/json")

	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, httpRequest)

	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected error 400, got %d", rw.Result().StatusCode)
	}
}

func TestGenericWebhookHandlerWrongProject(t *testing.T) {
	store := newTestStoreWithFakeProject2()
	router := newMockRouterSimpleEvent(store)

	httpRequest := httptest.NewRequest("POST", "/simpleevent/brigade-fakeProject/fakeCode", bytes.NewBuffer([]byte(exampleSimpleEvent)))
	httpRequest.Header.Add("Content-Type", "application/json")

	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, httpRequest)

	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected error 400, got %d", rw.Result().StatusCode)
	}
}

func TestGenericWebhookHandlerCorrectProjectEmptySecret(t *testing.T) {
	store := newTestStoreWithFakeProject()
	router := newMockRouterSimpleEvent(store)

	httpRequest := httptest.NewRequest("POST", "/simpleevent/brigade-fakeProject/fakeCode", bytes.NewBuffer([]byte(exampleSimpleEvent)))
	httpRequest.Header.Add("Content-Type", "application/json")

	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, httpRequest)

	if rw.Result().StatusCode != http.StatusUnauthorized {
		t.Errorf("expected error 401, got %d", rw.Result().StatusCode)
	}
}

func TestGenericWebhookHandlerCorrectProjectCorrectSecretSimpleEvent(t *testing.T) {
	getTestInfra := func(postdata string) (*mock.Store, *gin.Engine, *httptest.ResponseRecorder, *http.Request) {
		store := newTestStoreWithFakeProject()
		store.ProjectList[0].GenericGatewaySecret = "fakeCode"
		router := newMockRouterSimpleEvent(store)
		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest("POST", "/simpleevent/brigade-fakeProject/fakeCode", bytes.NewBuffer([]byte(postdata)))
		httpRequest.Header.Add("Content-Type", "application/json")
		return store, router, rw, httpRequest
	}

	// this test checks both ref and commit in JSON payload
	store, router, rw, httpRequest := getTestInfra(exampleSimpleEvent)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusOK {
		t.Errorf("expected error 200, got %d", rw.Result().StatusCode)
	}
	checkBuild(t, store, "refs/heads/changes", "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28", []byte(exampleSimpleEvent))

	// this test checks only ref in JSON payload
	exampleSimpleEvent2 := `
	{
		"ref": "refs/heads/changes"
	}
	`
	store, router, rw, httpRequest = getTestInfra(exampleSimpleEvent2)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusOK {
		t.Errorf("expected error 200, got %d", rw.Result().StatusCode)
	}
	checkBuild(t, store, "refs/heads/changes", "", []byte(exampleSimpleEvent2))

	// this test checks only commit in JSON payload
	exampleSimpleEvent3 := `
	{
		"commit": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"
	}
	`
	store, router, rw, httpRequest = getTestInfra(exampleSimpleEvent3)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusOK {
		t.Errorf("expected error 200, got %d", rw.Result().StatusCode)
	}
	checkBuild(t, store, "", "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28", []byte(exampleSimpleEvent3))

	exampleSimpleEvent4 := `
	{
		"val1": "refs/heads/changes",
		"val2": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"
	}
	`
	store, router, rw, httpRequest = getTestInfra(exampleSimpleEvent4)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusOK {
		t.Errorf("expected error 200, got %d", rw.Result().StatusCode)
	}
	checkBuild(t, store, "master", "", []byte(exampleSimpleEvent4))

	exampleSimpleEvent5 := `
	{
		"val1": "refs/heads/changes",
		"val2": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"
		CORRUPT
	}
	`
	_, router, rw, httpRequest = getTestInfra(exampleSimpleEvent5)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected error 400, got %d", rw.Result().StatusCode)
	}

	exampleSimpleEvent6 := `{}`
	store, router, rw, httpRequest = getTestInfra(exampleSimpleEvent6)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusOK {
		t.Errorf("expected error 200, got %d", rw.Result().StatusCode)
	}
	checkBuild(t, store, "master", "", []byte(exampleSimpleEvent6))

	exampleSimpleEvent7 := ``
	_, router, rw, httpRequest = getTestInfra(exampleSimpleEvent7)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected error 400, got %d", rw.Result().StatusCode)
	}
}

func checkBuild(t *testing.T, store *mock.Store, expectedRef string, expectedCommit string, payload []byte) {
	// timeout check here is necessary because handler ultimately runs in a goroutine
	c := make(chan struct{})
	go func() {
		for {
			if len(store.Builds) == 0 {
				time.Sleep(50 * time.Millisecond)
			} else {
				c <- struct{}{}
				break
			}
		}
	}()

	select {
	case <-c:
		break
	case <-time.After(3 * time.Second):
		t.Errorf("No new Builds were created, expectedRef %s and expectedCommit %s", expectedRef, expectedCommit)
		return
	}

	build := store.Builds[0]
	if build.Revision.Ref != expectedRef || build.Revision.Commit != expectedCommit {
		t.Errorf("Wrong Revision, expected %#v got %v#", exampleSimpleEvent, build.Revision)
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

	events := router.Group("/simpleevent")
	{
		events.Use(gin.Logger())
		events.POST("/:projectID/:secret", handler)
	}

	return router
}

func newEmptyTestStore() *mock.Store {
	return &mock.Store{}
}

func newTestStoreWithFakeProject() *mock.Store {
	return &mock.Store{
		ProjectList: []*brigade.Project{{
			ID: "brigade-fakeProject",
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
