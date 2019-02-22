package webhook

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
	"github.com/Azure/brigade/pkg/storage/mock"

	gin "gopkg.in/gin-gonic/gin.v1"

	cloudevents "github.com/cloudevents/sdk-go/v02"
)

func newTestGenericWebhookHandlerCloudEvent(store storage.Store) *genericWebhookCloudEvent {
	return &genericWebhookCloudEvent{store}
}

func TestGenericWebhookCloudEventHandler(t *testing.T) {
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

func TestGenericWebhookHandlerCloudEvent(t *testing.T) {
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
			url:            "/cloudevents/v02/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusBadRequest,
			store:          newTestStoreWithFakeProject2(),
			payload:        exampleCloudEvent,
		},
		{
			description:    "Correct project, empty secret",
			url:            "/cloudevents/v02/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusUnauthorized,
			store:          newTestStoreWithFakeProject(),
			payload:        exampleCloudEvent,
		},
		{
			description:    "both ref and commit in JSON payload",
			url:            "/cloudevents/v02/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusOK,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload: `
			{
				"type":   "com.example.file.created",
				"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
				"id":     "ea35b24ede421",
				"specversion": "0.2",
				"data": {
					"ref": "refs/heads/changes",
					"commit": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"
				}
			}`,
			revision: &brigade.Revision{Ref: "refs/heads/changes", Commit: "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"},
		},
		{
			description:    "only ref in JSON payload",
			url:            "/cloudevents/v02/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusOK,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload: `
			{
				"type":   "com.example.file.created",
				"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
				"id":     "ea35b24ede421",
				"specversion": "0.2",
				"data": {
					"ref": "refs/heads/changes"
				}
			}`,
			revision: &brigade.Revision{Ref: "refs/heads/changes"},
		},
		{
			description:    "only commit in JSON payload",
			url:            "/cloudevents/v02/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusOK,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload: `
			{
				"type":   "com.example.file.created",
				"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
				"id":     "ea35b24ede421",
				"specversion": "0.2",
				"data": {
					"commit": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"
				}
			}`,
			revision: &brigade.Revision{Commit: "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"},
		},
		{
			description:    "custom JSON object inside commit",
			url:            "/cloudevents/v02/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusOK,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload: `
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
			}`,
			revision: &brigade.Revision{Ref: "master"},
		},
		{
			description:    "custom JSON object inside ref",
			url:            "/cloudevents/v02/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusOK,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload: `
			{
				"type":   "com.example.file.created",
				"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
				"id":     "ea35b24ede421",
				"specversion": "0.2",
				"data": {
					"ref": {
						"hello": "another JSON object! hooray!"
					}
				}
			}`,
			revision: &brigade.Revision{Ref: "master"},
		},
		{
			description:    "random values inside data",
			url:            "/cloudevents/v02/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusOK,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload: `
			{
				"type":   "com.example.file.created",
				"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
				"id":     "ea35b24ede421",
				"specversion": "0.2",
				"data": {
					"val1": "refs/heads/changes",
					"val2": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"
				}
			}`,
			revision: &brigade.Revision{Ref: "master"},
		},
		{
			description:    "corrupt JSON",
			url:            "/cloudevents/v02/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusBadRequest,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload: `
			{
				"type":   "com.example.file.created",
				"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
				"id":     "ea35b24ede421",
				"specversion": "0.2",
				"data": {
					"val1": "refs/heads/changes",
					"val2": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28"
				}CORRUPT
			}`,
			revision: nil,
		},
		{
			description:    "missing Spec version",
			url:            "/cloudevents/v02/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusBadRequest,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload: `
			{
				"type":   "com.example.file.created",
				"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
				"id":     "ea35b24ede421"
			}`,
			revision: nil,
		},
		{
			description:    "empty POST data",
			url:            "/cloudevents/v02/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusBadRequest,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload:        ``,
			revision:       nil,
		},
		{
			description:    "wrong spec version",
			url:            "/cloudevents/v02/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusBadRequest,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload: `
			{
				"type":   "com.example.file.created",
				"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
				"id":     "ea35b24ede421",
				"specversion": "0.1"
			}`,
			revision: nil,
		},
		{
			description:    "missing type",
			url:            "/cloudevents/v02/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusBadRequest,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload: `
			{
				"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
				"id":     "ea35b24ede421",
				"specversion": "0.2"
			}`,
			revision: nil,
		},
		{
			description:    "missing source",
			url:            "/cloudevents/v02/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusBadRequest,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload: `
			{
				"type":   "com.example.file.created",
				"id":     "ea35b24ede421",
				"specversion": "0.2"
			}`,
			revision: nil,
		},
		{
			description:    "missing id",
			url:            "/cloudevents/v02/brigade-fakeProject/fakeCode",
			statusExpected: http.StatusBadRequest,
			store:          newTestStoreWithFakeProjectAndSecret("fakeCode"),
			payload: `
			{
				"type":   "com.example.file.created",
				"source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
				"specversion": "0.2"
			}`,
			revision: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			router := newMockRouterCloudEvent(test.store)
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

	events := router.Group("/cloudevents/v02")
	events.Use(gin.Logger())
	events.POST("/:projectID/:secret", handler)

	return router
}
