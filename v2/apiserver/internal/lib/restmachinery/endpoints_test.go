package restmachinery

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

var testSchema = gojsonschema.NewBytesLoader(
	[]byte(`
		{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"$id": "github.com/lovethedrake/drakecore/config.schema.json",

			"title": "Project",
			"type": "object",
			"required": ["foo"],
			"properties": {
				"foo": {
					"type": "string"
				}
			}
		}
		`,
	),
)

type testType struct {
	Foo string `json:"foo"`
}

func (t *testType) String() string {
	return t.Foo
}

func TestReadAndValidateRequestBody(t *testing.T) {

	testCases := []struct {
		name  string
		setup func(
			t *testing.T,
		) (*http.Request, gojsonschema.JSONLoader, interface{})
		assertions func(
			t *testing.T,
			result bool,
			obj interface{},
			rr *httptest.ResponseRecorder,
		)
	}{

		{
			name: "malformed JSON request body + no validation + no body object",
			setup: func(
				t *testing.T,
			) (*http.Request, gojsonschema.JSONLoader, interface{}) {
				// An empty request body is definitely not well-formed JSON
				r, err :=
					http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte{}))
				require.NoError(t, err)
				return r, nil, nil
			},
			assertions: func(
				t *testing.T,
				result bool,
				obj interface{},
				rr *httptest.ResponseRecorder,
			) {
				assert.True(t, result)
				// OK is a response's default status. If nothing went wrong, that
				// shouldn't have been overridden.
				require.Equal(t, http.StatusOK, rr.Result().StatusCode)
				// If nothing went wrong, we also should not have written any response
				// yet.
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Empty(t, bodyBytes)
				// We're expecting our body object to still be nil
				require.Nil(t, obj)
			},
		},

		{
			name: "malformed JSON request body + no validation + body object",
			setup: func(
				t *testing.T,
			) (*http.Request, gojsonschema.JSONLoader, interface{}) {
				// An empty request body is definitely not well-formed JSON
				r, err :=
					http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte{}))
				require.NoError(t, err)
				return r, nil, &testType{}
			},
			assertions: func(
				t *testing.T,
				result bool,
				obj interface{},
				rr *httptest.ResponseRecorder,
			) {
				assert.False(t, result)
				// Marshaling should have failed
				require.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(
					t,
					string(bodyBytes),
					"Request body contains malformed JSON",
				)
				// We're not expecting our body object to have been populated
				require.NotNil(t, obj)
				require.Empty(t, obj.(*testType).Foo)
			},
		},

		{
			name: "malformed JSON request body + validation + no body object",
			setup: func(
				t *testing.T,
			) (*http.Request, gojsonschema.JSONLoader, interface{}) {
				// An empty request body is definitely not well-formed JSON
				r, err :=
					http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte{}))
				require.NoError(t, err)
				return r, testSchema, nil
			},
			assertions: func(
				t *testing.T,
				result bool,
				obj interface{},
				rr *httptest.ResponseRecorder,
			) {
				assert.False(t, result)
				// Validation should have failed
				require.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(
					t,
					string(bodyBytes),
					"Request body contains malformed JSON",
				)
				// We're expecting our body object to still be nil
				require.Nil(t, obj)
			},
		},

		{
			name: "malformed JSON request body + validation + body object",
			setup: func(
				t *testing.T,
			) (*http.Request, gojsonschema.JSONLoader, interface{}) {
				// An empty request body is definitely not well-formed JSON
				r, err :=
					http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte{}))
				require.NoError(t, err)
				return r, testSchema, &testType{}
			},
			assertions: func(
				t *testing.T,
				result bool,
				obj interface{},
				rr *httptest.ResponseRecorder,
			) {
				assert.False(t, result)
				// Validation should have failed
				require.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(
					t,
					string(bodyBytes),
					"Request body contains malformed JSON",
				)
				// We're not expecting our body object to have been populated
				require.NotNil(t, obj)
				require.Empty(t, obj.(*testType).Foo)
			},
		},

		{
			name: "well-formed JSON request body + no validation + no body object",
			setup: func(
				t *testing.T,
			) (*http.Request, gojsonschema.JSONLoader, interface{}) {
				r, err :=
					http.NewRequest(
						http.MethodGet,
						"/",
						bytes.NewBuffer([]byte(`{"foo": "bar"}`)),
					)
				require.NoError(t, err)
				return r, nil, nil
			},
			assertions: func(
				t *testing.T,
				result bool,
				obj interface{},
				rr *httptest.ResponseRecorder,
			) {
				assert.True(t, result)
				// OK is a response's default status. If nothing went wrong, that
				// shouldn't have been overridden.
				require.Equal(t, http.StatusOK, rr.Result().StatusCode)
				// If nothing went wrong, we also should not have written any response
				// yet.
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Empty(t, bodyBytes)
				// We're expecting our body object to still be nil
				require.Nil(t, obj)
			},
		},

		{
			name: "well-formed JSON request body + no validation + body object",
			setup: func(
				t *testing.T,
			) (*http.Request, gojsonschema.JSONLoader, interface{}) {
				r, err :=
					http.NewRequest(
						http.MethodGet,
						"/",
						bytes.NewBuffer([]byte(`{"foo": "bar"}`)),
					)
				require.NoError(t, err)
				return r, nil, &testType{}
			},
			assertions: func(
				t *testing.T,
				result bool,
				obj interface{},
				rr *httptest.ResponseRecorder,
			) {
				assert.True(t, result)
				// OK is a response's default status. If nothing went wrong, that
				// shouldn't have been overridden.
				require.Equal(t, http.StatusOK, rr.Result().StatusCode)
				// If nothing went wrong, we also should not have written any response
				// yet.
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Empty(t, bodyBytes)
				// We're expecting our body object to have been populated
				require.NotNil(t, obj)
				require.Equal(t, "bar", obj.(*testType).Foo)
			},
		},

		{
			name: "well-formed, invalid JSON request body + validation + no body object", // nolint: lll
			setup: func(
				t *testing.T,
			) (*http.Request, gojsonschema.JSONLoader, interface{}) {
				r, err :=
					http.NewRequest(
						http.MethodGet,
						"/",
						bytes.NewBuffer([]byte("{}")),
					)
				require.NoError(t, err)
				return r, testSchema, nil
			},
			assertions: func(
				t *testing.T,
				result bool,
				obj interface{},
				rr *httptest.ResponseRecorder,
			) {
				assert.False(t, result)
				// Validation should have failed
				require.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(
					t,
					string(bodyBytes),
					"Request body failed JSON validation",
				)
				// We're expecting our body object to still be nil
				require.Nil(t, obj)
			},
		},

		{
			name: "well-formed, invalid JSON request body + validation + body object",
			setup: func(
				t *testing.T,
			) (*http.Request, gojsonschema.JSONLoader, interface{}) {
				r, err :=
					http.NewRequest(
						http.MethodGet,
						"/",
						bytes.NewBuffer([]byte("{}")),
					)
				require.NoError(t, err)
				return r, testSchema, &testType{}
			},
			assertions: func(
				t *testing.T,
				result bool,
				obj interface{},
				rr *httptest.ResponseRecorder,
			) {
				assert.False(t, result)
				// Validation should have failed
				require.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(
					t,
					string(bodyBytes),
					"Request body failed JSON validation",
				)
				// We're not expecting our body object to have been populated
				require.NotNil(t, obj)
				require.Empty(t, obj.(*testType).Foo)
			},
		},

		{
			name: "well-formed, valid JSON request body + validation + no body object", // nolint: lll
			setup: func(
				t *testing.T,
			) (*http.Request, gojsonschema.JSONLoader, interface{}) {
				r, err :=
					http.NewRequest(
						http.MethodGet,
						"/",
						bytes.NewBuffer([]byte(`{"foo": "bar"}`)),
					)
				require.NoError(t, err)
				return r, testSchema, nil
			},
			assertions: func(
				t *testing.T,
				result bool,
				obj interface{},
				rr *httptest.ResponseRecorder,
			) {
				assert.True(t, result)
				// OK is a response's default status. If nothing went wrong, that
				// shouldn't have been overridden.
				require.Equal(t, http.StatusOK, rr.Result().StatusCode)
				// If nothing went wrong, we also should not have written any response
				// yet.
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Empty(t, bodyBytes)
				// We're expecting our body object to still be nil
				require.Nil(t, obj)
			},
		},

		{
			name: "well-formed, valid JSON request body + validation + body object",
			setup: func(
				t *testing.T,
			) (*http.Request, gojsonschema.JSONLoader, interface{}) {
				r, err :=
					http.NewRequest(
						http.MethodGet,
						"/",
						bytes.NewBuffer([]byte(`{"foo": "bar"}`)),
					)
				require.NoError(t, err)
				return r, testSchema, &testType{}
			},
			assertions: func(
				t *testing.T,
				result bool,
				obj interface{},
				rr *httptest.ResponseRecorder,
			) {
				assert.True(t, result)
				// OK is a response's default status. If nothing went wrong, that
				// shouldn't have been overridden.
				require.Equal(t, http.StatusOK, rr.Result().StatusCode)
				// If nothing went wrong, we also should not have written any response
				// yet.
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Empty(t, bodyBytes)
				// We're expecting our body object to have been populated
				require.NotNil(t, obj)
				require.Equal(t, "bar", obj.(*testType).Foo)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r, schema, obj := testCase.setup(t)
			rr := httptest.NewRecorder()
			result := ReadAndValidateRequestBody(rr, r, schema, obj)
			testCase.assertions(t, result, obj, rr)
		})
	}
}

func TestServeRequest(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func() InboundRequest
		assertions func(req InboundRequest)
	}{

		{
			name: "request fails validation",
			setup: func() InboundRequest {
				// An empty request body is definitely not well-formed JSON
				r, err :=
					http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte("")))
				require.NoError(t, err)
				return InboundRequest{
					R:                   r,
					W:                   httptest.NewRecorder(),
					ReqBodySchemaLoader: testSchema,
				}
			},
			assertions: func(req InboundRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				// Validation should fail
				require.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "BadRequestError")
			},
		},

		{
			name: "endpoint logic returns authn error",
			setup: func() InboundRequest {
				r, err :=
					http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte("")))
				require.NoError(t, err)
				return InboundRequest{
					R: r,
					W: httptest.NewRecorder(),
					EndpointLogic: func() (interface{}, error) {
						return nil, &meta.ErrAuthentication{}
					},
				}
			},
			assertions: func(req InboundRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "AuthenticationError")
			},
		},

		{
			name: "endpoint logic returns authz error",
			setup: func() InboundRequest {
				r, err :=
					http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte("")))
				require.NoError(t, err)
				return InboundRequest{
					R: r,
					W: httptest.NewRecorder(),
					EndpointLogic: func() (interface{}, error) {
						return nil, &meta.ErrAuthorization{}
					},
				}
			},
			assertions: func(req InboundRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusForbidden, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "AuthorizationError")
			},
		},

		{
			name: "endpoint logic returns bad request error",
			setup: func() InboundRequest {
				r, err :=
					http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte("")))
				require.NoError(t, err)
				return InboundRequest{
					R: r,
					W: httptest.NewRecorder(),
					EndpointLogic: func() (interface{}, error) {
						return nil, &meta.ErrBadRequest{}
					},
				}
			},
			assertions: func(req InboundRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "BadRequestError")
			},
		},

		{
			name: "endpoint logic returns not found error",
			setup: func() InboundRequest {
				r, err :=
					http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte("")))
				require.NoError(t, err)
				return InboundRequest{
					R: r,
					W: httptest.NewRecorder(),
					EndpointLogic: func() (interface{}, error) {
						return nil, &meta.ErrNotFound{}
					},
				}
			},
			assertions: func(req InboundRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusNotFound, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "NotFoundError")
			},
		},

		{
			name: "endpoint logic returns conflict error",
			setup: func() InboundRequest {
				r, err :=
					http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte("")))
				require.NoError(t, err)
				return InboundRequest{
					R: r,
					W: httptest.NewRecorder(),
					EndpointLogic: func() (interface{}, error) {
						return nil, &meta.ErrConflict{}
					},
				}
			},
			assertions: func(req InboundRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusConflict, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "ConflictError")
			},
		},

		{
			name: "endpoint logic returns not implemented error",
			setup: func() InboundRequest {
				r, err :=
					http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte("")))
				require.NoError(t, err)
				return InboundRequest{
					R: r,
					W: httptest.NewRecorder(),
					EndpointLogic: func() (interface{}, error) {
						return nil, &meta.ErrNotSupported{}
					},
				}
			},
			assertions: func(req InboundRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusNotImplemented, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "NotSupportedError")
			},
		},

		{
			name: "endpoint logic returns internal server error",
			setup: func() InboundRequest {
				r, err :=
					http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte("")))
				require.NoError(t, err)
				return InboundRequest{
					R: r,
					W: httptest.NewRecorder(),
					EndpointLogic: func() (interface{}, error) {
						return nil, &meta.ErrInternalServer{}
					},
				}
			},
			assertions: func(req InboundRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusInternalServerError, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "InternalServerError")
			},
		},

		{
			name: "endpoint logic returns unanticipated error",
			setup: func() InboundRequest {
				r, err :=
					http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte("")))
				require.NoError(t, err)
				return InboundRequest{
					R: r,
					W: httptest.NewRecorder(),
					EndpointLogic: func() (interface{}, error) {
						return nil, errors.New("something went wrong")
					},
				}
			},
			assertions: func(req InboundRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusInternalServerError, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "InternalServerError")
			},
		},

		{
			name: "endpoint logic returns response object",
			setup: func() InboundRequest {
				r, err :=
					http.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte("")))
				require.NoError(t, err)
				return InboundRequest{
					R: r,
					W: httptest.NewRecorder(),
					EndpointLogic: func() (interface{}, error) {
						return testType{Foo: "bar"}, nil
					},
				}
			},
			assertions: func(req InboundRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusOK, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "foo")
				require.Contains(t, string(bodyBytes), "bar")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req := testCase.setup()
			req.W = httptest.NewRecorder()
			ServeRequest(req)
			testCase.assertions(req)
		})
	}
}

func TestWriteAPIResponse(t *testing.T) {
	testCases := []struct {
		name     string
		response interface{}
	}{
		{
			name:     "response is bytes",
			response: `{"foo": "bar"}`,
		},
		{
			name: "response is a struct",
			response: testType{
				Foo: "bar",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			const testStatusCode = http.StatusNotFound
			rr := httptest.NewRecorder()
			WriteAPIResponse(rr, testStatusCode, testCase.response)
			require.Equal(t, testStatusCode, rr.Result().StatusCode)
			bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
			require.NoError(t, err)
			require.Contains(t, string(bodyBytes), "foo")
			require.Contains(t, string(bodyBytes), "bar")
		})
	}
}

func TestServeWebUIRequest(t *testing.T) {
	testCases := []struct {
		name       string
		req        InboundWebUIRequest
		assertions func(req InboundWebUIRequest)
	}{

		{
			name: "endpoint logic returns authn error",
			req: InboundWebUIRequest{
				W: httptest.NewRecorder(),
				EndpointLogic: func() (interface{}, error) {
					return nil, &meta.ErrAuthentication{}
				},
			},
			assertions: func(req InboundWebUIRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(
					t,
					string(bodyBytes),
					"Could not authenticate the request",
				)
			},
		},

		{
			name: "endpoint logic returns authz error",
			req: InboundWebUIRequest{
				W: httptest.NewRecorder(),
				EndpointLogic: func() (interface{}, error) {
					return nil, &meta.ErrAuthorization{}
				},
			},
			assertions: func(req InboundWebUIRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusForbidden, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "The request is not authorized")
			},
		},

		{
			name: "endpoint logic returns bad request error",
			req: InboundWebUIRequest{
				W: httptest.NewRecorder(),
				EndpointLogic: func() (interface{}, error) {
					return nil, &meta.ErrBadRequest{}
				},
			},
			assertions: func(req InboundWebUIRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "Bad request")
			},
		},

		{
			name: "endpoint logic returns not found error",
			req: InboundWebUIRequest{
				W: httptest.NewRecorder(),
				EndpointLogic: func() (interface{}, error) {
					return nil, &meta.ErrNotFound{}
				},
			},
			assertions: func(req InboundWebUIRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusNotFound, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "not found")
			},
		},

		{
			name: "endpoint logic returns conflict error",
			req: InboundWebUIRequest{
				W: httptest.NewRecorder(),
				EndpointLogic: func() (interface{}, error) {
					return nil, &meta.ErrConflict{}
				},
			},
			assertions: func(req InboundWebUIRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusConflict, rr.Result().StatusCode)
			},
		},

		{
			name: "endpoint logic returns not implemented error",
			req: InboundWebUIRequest{
				W: httptest.NewRecorder(),
				EndpointLogic: func() (interface{}, error) {
					return nil, &meta.ErrNotSupported{}
				},
			},
			assertions: func(req InboundWebUIRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusNotImplemented, rr.Result().StatusCode)
			},
		},

		{
			name: "endpoint logic returns internal server error",
			req: InboundWebUIRequest{
				W: httptest.NewRecorder(),
				EndpointLogic: func() (interface{}, error) {
					return nil, &meta.ErrInternalServer{}
				},
			},
			assertions: func(req InboundWebUIRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusInternalServerError, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(
					t,
					string(bodyBytes),
					"An internal server error occurred",
				)
			},
		},

		{
			name: "endpoint logic returns unanticipated error",
			req: InboundWebUIRequest{
				W: httptest.NewRecorder(),
				EndpointLogic: func() (interface{}, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(req InboundWebUIRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusInternalServerError, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(
					t,
					string(bodyBytes),
					"An internal server error occurred",
				)
			},
		},

		{
			name: "endpoint logic returns bytes",
			req: InboundWebUIRequest{
				W: httptest.NewRecorder(),
				EndpointLogic: func() (interface{}, error) {
					return []byte("lorem ipsum"), nil
				},
			},
			assertions: func(req InboundWebUIRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusOK, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "lorem ipsum")
			},
		},

		{
			name: "endpoint logic returns string",
			req: InboundWebUIRequest{
				W: httptest.NewRecorder(),
				EndpointLogic: func() (interface{}, error) {
					return "lorem ipsum", nil
				},
			},
			assertions: func(req InboundWebUIRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusOK, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "lorem ipsum")
			},
		},

		{
			name: "endpoint logic returns fmt.Stringer",
			req: InboundWebUIRequest{
				W: httptest.NewRecorder(),
				EndpointLogic: func() (interface{}, error) {
					return &testType{Foo: "lorem ipsum"}, nil
				},
			},
			assertions: func(req InboundWebUIRequest) {
				rr := req.W.(*httptest.ResponseRecorder)
				require.Equal(t, http.StatusOK, rr.Result().StatusCode)
				bodyBytes, err := ioutil.ReadAll(rr.Result().Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "lorem ipsum")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ServeWebUIRequest(testCase.req)
			testCase.assertions(testCase.req)
		})
	}
}
