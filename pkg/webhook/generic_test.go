package webhook

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
	"github.com/Azure/brigade/pkg/storage/mock"

	gin "gopkg.in/gin-gonic/gin.v1"
)

func newTestGenericWebhookHandler(store storage.Store) *genericWebhook {
	return &genericWebhook{store}
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
	}
}

func TestGenericWebhook(t *testing.T) {
	proj := newGenericProject()
	store := newTestStore()
	h := newTestGenericWebhookHandler(store)
	gwData := &genericWebhookData{
		Commit: "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28",
	}

	if err := h.genericWebhookEvent(proj, []byte(exampleGenericWebhook), gwData); err != nil {
		t.Errorf("failed generic gateway event: %s", err)
	}

	if payload := string(store.builds[0].Payload); payload != exampleGenericWebhook {
		t.Errorf("unexpected payload: %s", payload)
	}

	if provider := string(store.builds[0].Provider); provider != "GenericWebhook" {
		t.Errorf("unexpected provider: %s", provider)
	}

	if etype := string(store.builds[0].Type); etype != "webhook" {
		t.Errorf("unexpected type: %s", etype)
	}
}

func TestGenericWebhookHandlerBadBody(t *testing.T) {
	store := newEmptyTestStore()
	router := newMockRouter(store)

	httpRequest := httptest.NewRequest("POST", "/webhook/fakeProject/fakeCode", bytes.NewBuffer(nil))
	httpRequest.Header.Add("Content-Type", "application/json")

	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, httpRequest)

	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected error 400, got %d", rw.Result().StatusCode)
	}
}

func TestGenericWebhookHandlerWrongProject(t *testing.T) {

	store := newTestStoreWithFakeProject2()
	router := newMockRouter(store)

	httpRequest := httptest.NewRequest("POST", "/webhook/brigade-fakeProject/fakeCode", bytes.NewBuffer([]byte(exampleGenericWebhook)))
	httpRequest.Header.Add("Content-Type", "application/json")

	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, httpRequest)

	if rw.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("expected error 400, got %d", rw.Result().StatusCode)
	}
}

func TestGenericWebhookHandlerCorrectProjectEmptySecret(t *testing.T) {
	store := newTestStoreWithFakeProject()
	router := newMockRouter(store)

	httpRequest := httptest.NewRequest("POST", "/webhook/brigade-fakeProject/fakeCode", bytes.NewBuffer([]byte(exampleGenericWebhook)))
	httpRequest.Header.Add("Content-Type", "application/json")

	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, httpRequest)

	if rw.Result().StatusCode != http.StatusUnauthorized {
		t.Errorf("expected error 401, got %d", rw.Result().StatusCode)
	}
}

func TestGenericWebhookHandlerCorrectProjectCorrectSecret(t *testing.T) {
	store := newTestStoreWithFakeProject()
	store.ProjectList[0].GenericGatewaySecret = "fakeCode"
	router := newMockRouter(store)

	httpRequest := httptest.NewRequest("POST", "/webhook/brigade-fakeProject/fakeCode", bytes.NewBuffer([]byte(exampleGenericWebhook)))
	httpRequest.Header.Add("Content-Type", "application/json")

	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, httpRequest)

	if rw.Result().StatusCode != http.StatusOK {
		t.Errorf("expected error 200, got %d", rw.Result().StatusCode)
	}
}

const exampleGenericWebhook = `
{
	"ref": "refs/heads/changes",
	"commit": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"
}
`

func newMockRouter(store storage.Store) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	handler := NewGenericWebhook(store)

	events := router.Group("/webhook")
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
