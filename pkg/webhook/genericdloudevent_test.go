package webhook

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Azure/brigade/pkg/storage"
	"github.com/Azure/brigade/pkg/storage/mock"

	gin "gopkg.in/gin-gonic/gin.v1"

	cloudevents "github.com/cloudevents/sdk-go/v02"
)

func newTestGenericWebhookHandlerCloudEvent(store storage.Store) *genericWebhookCloudEvent {
	return &genericWebhookCloudEvent{store}
}

func TestGenericWebhookCloudEvent(t *testing.T) {
	proj := newGenericProject()
	store := newTestStore()
	h := newTestGenericWebhookHandlerCloudEvent(store)
	source, _ := url.Parse("/providers/Example.COM/storage/account#fileServices/default/{new-file}")
	event := &cloudevents.Event{
		Type:   "com.example.file.created",
		Source: *source,
		ID:     "ea35b24ede421",
	}

	if err := h.genericWebhookCloudEvent(proj, []byte(exampleCloudEvent), event); err != nil {
		t.Errorf("failed generic gateway cloud event: %s", err)
	}

	if payload := string(store.builds[0].Payload); payload != exampleCloudEvent {
		t.Errorf("unexpected payload: %s", payload)
	}

	if provider := string(store.builds[0].Provider); provider != "GenericWebhook" {
		t.Errorf("unexpected provider: %s", provider)
	}

	if etype := string(store.builds[0].Type); etype != "cloudevent" {
		t.Errorf("unexpected type: %s", etype)
	}
}

func TestGenericWebhookHandlerCloudEventBadBody(t *testing.T) {
	store := newEmptyTestStore()
	router := newMockRouterCloudEvent(store)

	httpRequest := httptest.NewRequest("POST", "/cloudevent/fakeProject/fakeCode", bytes.NewBuffer(nil))
	httpRequest.Header.Add("Content-Type", "application/json")

	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, httpRequest)

	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected error 400, got %d", rw.Result().StatusCode)
	}
}

func TestGenericWebhookHandlerCloudEventWrongProject(t *testing.T) {

	store := newTestStoreWithFakeProject2()
	router := newMockRouterCloudEvent(store)

	httpRequest := httptest.NewRequest("POST", "/cloudevent/brigade-fakeProject/fakeCode", bytes.NewBuffer([]byte(exampleCloudEvent)))
	httpRequest.Header.Add("Content-Type", "application/json")

	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, httpRequest)

	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected error 400, got %d", rw.Result().StatusCode)
	}
}

func TestGenericWebhookHandlerCloudEventCorrectProjectEmptySecret(t *testing.T) {
	store := newTestStoreWithFakeProject()
	router := newMockRouterCloudEvent(store)

	httpRequest := httptest.NewRequest("POST", "/cloudevent/brigade-fakeProject/fakeCode", bytes.NewBuffer([]byte(exampleCloudEvent)))
	httpRequest.Header.Add("Content-Type", "application/json")

	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, httpRequest)

	if rw.Result().StatusCode != http.StatusUnauthorized {
		t.Errorf("expected error 401, got %d", rw.Result().StatusCode)
	}
}

func TestGenericWebhookHandlerCorrectProjectCorrectSecretCloudEvent(t *testing.T) {
	getTestInfra := func(postdata string) (*mock.Store, *gin.Engine, *httptest.ResponseRecorder, *http.Request) {
		store := newTestStoreWithFakeProject()
		store.ProjectList[0].GenericGatewaySecret = "fakeCode"
		router := newMockRouterCloudEvent(store)
		rw := httptest.NewRecorder()
		httpRequest := httptest.NewRequest("POST", "/cloudevent/brigade-fakeProject/fakeCode", bytes.NewBuffer([]byte(postdata)))
		httpRequest.Header.Add("Content-Type", "application/json")
		return store, router, rw, httpRequest
	}

	////////////////////////////////////////////////////////////////////////////////////////

	exampleCloudEvent1 := `
	{
		"type":   "com.example.file.created",
		"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
		"id":     "ea35b24ede421",
		"specversion": "0.2",
		"data": {
			"ref": "refs/heads/changes",
			"commit": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"
		}
	}
	`

	// this test checks both ref and commit in JSON payload
	store, router, rw, httpRequest := getTestInfra(exampleCloudEvent1)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", rw.Result().StatusCode)
	}
	checkBuild(t, store, "refs/heads/changes", "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28", []byte(exampleCloudEvent1))

	////////////////////////////////////////////////////////////////////////////////////////
	// this test checks only ref in JSON payload

	exampleCloudEvent2 := `
	{
		"type":   "com.example.file.created",
		"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
		"id":     "ea35b24ede421",
		"specversion": "0.2",
		"data": {
			"ref": "refs/heads/changes"
		}
	}
	`
	store, router, rw, httpRequest = getTestInfra(exampleCloudEvent2)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", rw.Result().StatusCode)
	}
	checkBuild(t, store, "refs/heads/changes", "", []byte(exampleCloudEvent2))

	////////////////////////////////////////////////////////////////////////////////////////
	// this test checks only commit in JSON payload

	exampleCloudEvent3 := `
	{
		"type":   "com.example.file.created",
		"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
		"id":     "ea35b24ede421",
		"specversion": "0.2",
		"data": {
			"commit": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"
		}
	}
	`
	store, router, rw, httpRequest = getTestInfra(exampleCloudEvent3)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", rw.Result().StatusCode)
	}
	checkBuild(t, store, "", "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28", []byte(exampleCloudEvent3))

	////////////////////////////////////////////////////////////////////////////////////////
	// this test checks for a JSON object inside commit

	exampleCloudEvent4 := `
	{
		"type":   "com.example.file.created",
		"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
		"id":     "ea35b24ede421",
		"specversion": "0.2",
		"data": {
			"commit": {
				"hello": "another JSON object! hooray!"
			}
		}
	}
	`
	store, router, rw, httpRequest = getTestInfra(exampleCloudEvent4)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusOK {
		t.Errorf("expected error 200, got %d", rw.Result().StatusCode)
	}
	checkBuild(t, store, "master", "", []byte(exampleCloudEvent4))

	////////////////////////////////////////////////////////////////////////////////////////
	// this test checks for a JSON object inside ref

	exampleCloudEvent5 := `
	{
		"type":   "com.example.file.created",
		"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
		"id":     "ea35b24ede421",
		"specversion": "0.2",
		"data": {
			"ref": {
				"hello": "another JSON object! hooray!"
			},
			"commit": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"
		}
	}
	`
	store, router, rw, httpRequest = getTestInfra(exampleCloudEvent5)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusOK {
		t.Errorf("expected error 200, got %d", rw.Result().StatusCode)
	}
	checkBuild(t, store, "", "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28", []byte(exampleCloudEvent5))

	////////////////////////////////////////////////////////////////////////////////////////
	//this test tests for random values inside Data

	exampleCloudEvent6 := `
	{
		"type":   "com.example.file.created",
		"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
		"id":     "ea35b24ede421",
		"specversion": "0.2",
		"data": {
			"val1": "refs/heads/changes",
			"val2": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"
		}
	}
	`

	store, router, rw, httpRequest = getTestInfra(exampleCloudEvent6)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusOK {
		t.Errorf("expected error 200, got %d", rw.Result().StatusCode)
	}
	checkBuild(t, store, "master", "", []byte(exampleCloudEvent6))

	////////////////////////////////////////////////////////////////////////////////////////
	//corrupt JSON

	exampleCloudEvent7 := `
	{
		"type":   "com.example.file.created",
		"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
		"id":     "ea35b24ede421",
		"specversion": "0.2",
		"data": {
			"val1": "refs/heads/changes",
			"val2": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"
			CORRUPT
		}
	}
	`

	_, router, rw, httpRequest = getTestInfra(exampleCloudEvent7)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected error 400, got %d", rw.Result().StatusCode)
	}

	////////////////////////////////////////////////////////////////////////////////////////
	//missing specversion

	exampleCloudEvent8 := `
	{
		"type":   "com.example.file.created",
		"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
		"id":     "ea35b24ede421"
	}
	`
	_, router, rw, httpRequest = getTestInfra(exampleCloudEvent8)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected error 400, got %d", rw.Result().StatusCode)
	}

	////////////////////////////////////////////////////////////////////////////////////////
	//empty POST data

	exampleCloudEvent9 := ``
	_, router, rw, httpRequest = getTestInfra(exampleCloudEvent9)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected error 400, got %d", rw.Result().StatusCode)
	}

	////////////////////////////////////////////////////////////////////////////////////////
	//empty JSON

	exampleCloudEvent10 := ``
	_, router, rw, httpRequest = getTestInfra(exampleCloudEvent10)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected error 400, got %d", rw.Result().StatusCode)
	}

	////////////////////////////////////////////////////////////////////////////////////////
	//wrong specversion

	exampleCloudEvent11 := `
	{
		"type":   "com.example.file.created",
		"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
		"id":     "ea35b24ede421",
		"specversion": "0.1"
	}
	`
	_, router, rw, httpRequest = getTestInfra(exampleCloudEvent11)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected error 400, got %d", rw.Result().StatusCode)
	}

	////////////////////////////////////////////////////////////////////////////////////////
	// missing type

	exampleCloudEvent12 := `
	{
		"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
		"id":     "ea35b24ede421",
		"specversion": "0.2"
	}
	`
	_, router, rw, httpRequest = getTestInfra(exampleCloudEvent12)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected error 400, got %d", rw.Result().StatusCode)
	}

	////////////////////////////////////////////////////////////////////////////////////////
	//missing source

	exampleCloudEvent13 := `
	{
		"type":   "com.example.file.created",
		"id":     "ea35b24ede421",
		"specversion": "0.2"
	}
	`
	_, router, rw, httpRequest = getTestInfra(exampleCloudEvent13)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected error 400, got %d", rw.Result().StatusCode)
	}

	////////////////////////////////////////////////////////////////////////////////////////
	//missing id

	exampleCloudEvent14 := `
	{
		"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
		"type":   "com.example.file.created",
		"specversion": "0.2"
	}
	`
	_, router, rw, httpRequest = getTestInfra(exampleCloudEvent14)
	router.ServeHTTP(rw, httpRequest)
	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected error 400, got %d", rw.Result().StatusCode)
	}

}

const exampleCloudEvent = `
{
	"type":   "com.example.file.created",
	"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
	"id":     "ea35b24ede421",
	"specversion": "0.2"
}
`

func newMockRouterCloudEvent(store storage.Store) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	handler := NewGenericWebhookCloudEvent(store)

	events := router.Group("/cloudevent")
	events.Use(gin.Logger())
	events.POST("/:projectID/:secret", handler)

	return router
}
