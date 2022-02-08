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
			"$id": "github.com/brigadecore/brigade/v2/test.schema.json",

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
			r *http.Response,
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
				r *http.Response,
			) {
				assert.True(t, result)
				// OK is a response's default status. If nothing went wrong, that
				// shouldn't have been overridden.
				require.Equal(t, http.StatusOK, r.StatusCode)
				// If nothing went wrong, we also should not have written any response
				// yet.
				bodyBytes, err := ioutil.ReadAll(r.Body)
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
				r *http.Response,
			) {
				assert.False(t, result)
				// Marshaling should have failed
				require.Equal(t, http.StatusBadRequest, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Contains(
					t,
					string(bodyBytes),
					"Request body contains malformed JSON",
				)
				// We're not expecting our body object to have been populated
				require.NotNil(t, obj)
				tt, ok := obj.(*testType)
				require.True(t, ok)
				require.Empty(t, tt.Foo)
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
				r *http.Response,
			) {
				assert.False(t, result)
				// Validation should have failed
				require.Equal(t, http.StatusBadRequest, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
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
				r *http.Response,
			) {
				assert.False(t, result)
				// Validation should have failed
				require.Equal(t, http.StatusBadRequest, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Contains(
					t,
					string(bodyBytes),
					"Request body contains malformed JSON",
				)
				// We're not expecting our body object to have been populated
				require.NotNil(t, obj)
				tt, ok := obj.(*testType)
				require.True(t, ok)
				require.Empty(t, tt.Foo)
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
				r *http.Response,
			) {
				assert.True(t, result)
				// OK is a response's default status. If nothing went wrong, that
				// shouldn't have been overridden.
				require.Equal(t, http.StatusOK, r.StatusCode)
				// If nothing went wrong, we also should not have written any response
				// yet.
				bodyBytes, err := ioutil.ReadAll(r.Body)
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
				r *http.Response,
			) {
				assert.True(t, result)
				// OK is a response's default status. If nothing went wrong, that
				// shouldn't have been overridden.
				require.Equal(t, http.StatusOK, r.StatusCode)
				// If nothing went wrong, we also should not have written any response
				// yet.
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Empty(t, bodyBytes)
				// We're expecting our body object to have been populated
				require.NotNil(t, obj)
				tt, ok := obj.(*testType)
				require.True(t, ok)
				require.Equal(t, "bar", tt.Foo)
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
				r *http.Response,
			) {
				assert.False(t, result)
				// Validation should have failed
				require.Equal(t, http.StatusBadRequest, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
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
				r *http.Response,
			) {
				assert.False(t, result)
				// Validation should have failed
				require.Equal(t, http.StatusBadRequest, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Contains(
					t,
					string(bodyBytes),
					"Request body failed JSON validation",
				)
				// We're not expecting our body object to have been populated
				require.NotNil(t, obj)
				tt, ok := obj.(*testType)
				require.True(t, ok)
				require.Empty(t, tt.Foo)
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
				r *http.Response,
			) {
				assert.True(t, result)
				// OK is a response's default status. If nothing went wrong, that
				// shouldn't have been overridden.
				require.Equal(t, http.StatusOK, r.StatusCode)
				// If nothing went wrong, we also should not have written any response
				// yet.
				bodyBytes, err := ioutil.ReadAll(r.Body)
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
				r *http.Response,
			) {
				assert.True(t, result)
				// OK is a response's default status. If nothing went wrong, that
				// shouldn't have been overridden.
				require.Equal(t, http.StatusOK, r.StatusCode)
				// If nothing went wrong, we also should not have written any response
				// yet.
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Empty(t, bodyBytes)
				// We're expecting our body object to have been populated
				require.NotNil(t, obj)
				tt, ok := obj.(*testType)
				require.True(t, ok)
				require.Equal(t, "bar", tt.Foo)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r, schema, obj := testCase.setup(t)
			rr := httptest.NewRecorder()
			result := ReadAndValidateRequestBody(rr, r, schema, obj)
			res := rr.Result()
			defer res.Body.Close()
			testCase.assertions(t, result, obj, res)
		})
	}
}

func TestServeRequest(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func() InboundRequest
		assertions func(*http.Response)
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
					ReqBodySchemaLoader: testSchema,
				}
			},
			assertions: func(r *http.Response) {
				// Validation should fail
				require.Equal(t, http.StatusBadRequest, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
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
					EndpointLogic: func() (interface{}, error) {
						return nil, &meta.ErrAuthentication{}
					},
				}
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusUnauthorized, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
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
					EndpointLogic: func() (interface{}, error) {
						return nil, &meta.ErrAuthorization{}
					},
				}
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusForbidden, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
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
					EndpointLogic: func() (interface{}, error) {
						return nil, &meta.ErrBadRequest{}
					},
				}
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusBadRequest, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
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
					EndpointLogic: func() (interface{}, error) {
						return nil, &meta.ErrNotFound{}
					},
				}
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusNotFound, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
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
					EndpointLogic: func() (interface{}, error) {
						return nil, &meta.ErrConflict{}
					},
				}
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusConflict, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
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
					EndpointLogic: func() (interface{}, error) {
						return nil, &meta.ErrNotSupported{}
					},
				}
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusNotImplemented, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
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
					EndpointLogic: func() (interface{}, error) {
						return nil, &meta.ErrInternalServer{}
					},
				}
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusInternalServerError, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
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
					EndpointLogic: func() (interface{}, error) {
						return nil, errors.New("something went wrong")
					},
				}
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusInternalServerError, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
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
					EndpointLogic: func() (interface{}, error) {
						return testType{Foo: "bar"}, nil
					},
				}
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusOK, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "foo")
				require.Contains(t, string(bodyBytes), "bar")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req := testCase.setup()
			rr := httptest.NewRecorder()
			req.W = rr
			ServeRequest(req)
			res := rr.Result()
			defer res.Body.Close()
			testCase.assertions(res)
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
			res := rr.Result()
			defer res.Body.Close()
			require.Equal(
				t,
				testStatusCode,
				res.StatusCode,
			)
			bodyBytes, err := ioutil.ReadAll(res.Body)
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
		assertions func(*http.Response)
	}{

		{
			name: "endpoint logic returns authn error",
			req: InboundWebUIRequest{
				EndpointLogic: func() (interface{}, error) {
					return nil, &meta.ErrAuthentication{}
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusUnauthorized, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
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
				EndpointLogic: func() (interface{}, error) {
					return nil, &meta.ErrAuthorization{}
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusForbidden, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "The request is not authorized")
			},
		},

		{
			name: "endpoint logic returns bad request error",
			req: InboundWebUIRequest{
				EndpointLogic: func() (interface{}, error) {
					return nil, &meta.ErrBadRequest{}
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusBadRequest, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "Bad request")
			},
		},

		{
			name: "endpoint logic returns not found error",
			req: InboundWebUIRequest{
				EndpointLogic: func() (interface{}, error) {
					return nil, &meta.ErrNotFound{}
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusNotFound, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "not found")
			},
		},

		{
			name: "endpoint logic returns conflict error",
			req: InboundWebUIRequest{
				EndpointLogic: func() (interface{}, error) {
					return nil, &meta.ErrConflict{}
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusConflict, r.StatusCode)
			},
		},

		{
			name: "endpoint logic returns not implemented error",
			req: InboundWebUIRequest{
				EndpointLogic: func() (interface{}, error) {
					return nil, &meta.ErrNotSupported{}
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusNotImplemented, r.StatusCode)
			},
		},

		{
			name: "endpoint logic returns internal server error",
			req: InboundWebUIRequest{
				EndpointLogic: func() (interface{}, error) {
					return nil, &meta.ErrInternalServer{}
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusInternalServerError, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
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
				EndpointLogic: func() (interface{}, error) {
					return nil, errors.New("something went wrong")
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusInternalServerError, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
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
				EndpointLogic: func() (interface{}, error) {
					return []byte("lorem ipsum"), nil
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusOK, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "lorem ipsum")
			},
		},

		{
			name: "endpoint logic returns string",
			req: InboundWebUIRequest{
				EndpointLogic: func() (interface{}, error) {
					return "lorem ipsum", nil
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusOK, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "lorem ipsum")
			},
		},

		{
			name: "endpoint logic returns fmt.Stringer",
			req: InboundWebUIRequest{
				EndpointLogic: func() (interface{}, error) {
					return &testType{Foo: "lorem ipsum"}, nil
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusOK, r.StatusCode)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Contains(t, string(bodyBytes), "lorem ipsum")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			testCase.req.W = rr
			ServeWebUIRequest(testCase.req)
			res := rr.Result()
			defer res.Body.Close()
			testCase.assertions(res)
		})
	}
}
