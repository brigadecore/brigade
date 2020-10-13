package restmachinery

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// Endpoints is an interface to be implemented by all REST API endpoints.
type Endpoints interface {
	// Register is invoked during Server initialization, giving endpoint
	// implementations an opportunity to register path/function mappings with
	// the provided *mux.Router.
	Register(router *mux.Router)
}

// ReadAndValidateRequestBody extracts the body of the provided raw HTTP request
// and, if applicable, validates it against the provided
// gojsonschema.JSONLoader. The JSON request body is then unmarshaled into the
// provided object. Validation errors and all other errors are dealt with by
// sending an appropriate status code and response body to the client. The
// function returns a bool indicating success (true) or failure (false). This is
// a lower-level function not often used directly, but more often invoked by the
// ServeRequest function.
func ReadAndValidateRequestBody(
	w http.ResponseWriter,
	r *http.Request,
	bodySchemaLoader gojsonschema.JSONLoader,
	bodyObj interface{},
) bool {
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// Log it in case something is actually wrong...
		log.Println(errors.Wrap(err, "error reading request body"))
		// But we're going to assume this is because the request body is missing, so
		// we'll treat it as a bad request.
		WriteAPIResponse(
			w,
			http.StatusBadRequest,
			&meta.ErrBadRequest{
				Reason: "Could not read request body",
			},
		)
		return false
	}
	if bodySchemaLoader != nil {
		var validationResult *gojsonschema.Result
		validationResult, err = gojsonschema.Validate(
			bodySchemaLoader,
			gojsonschema.NewBytesLoader(bodyBytes),
		)
		if err != nil {
			// Log it in case something is actually wrong...
			log.Printf("error validating request body: %s", err)
			// But as long as the schema itself was valid, the most likely scenario
			// here is that the request body wasn't well-formed JSON, so we'll treat
			// this as a bad request.
			WriteAPIResponse(
				w,
				http.StatusBadRequest,
				&meta.ErrBadRequest{
					Reason: "Request body contains malformed JSON",
				},
			)
			return false
		}
		if !validationResult.Valid() {
			// We don't bother to log this because this is DEFINITELY a bad request.
			verrStrs := make([]string, len(validationResult.Errors()))
			for i, verr := range validationResult.Errors() {
				verrStrs[i] = verr.String()
			}
			WriteAPIResponse(
				w,
				http.StatusBadRequest,
				&meta.ErrBadRequest{
					Reason:  "Request body failed JSON validation",
					Details: verrStrs,
				},
			)
			return false
		}
	}
	if bodyObj != nil {
		if err = json.Unmarshal(bodyBytes, bodyObj); err != nil {
			WriteAPIResponse(
				w,
				http.StatusBadRequest,
				&meta.ErrBadRequest{
					Reason: "Request body contains malformed JSON",
				},
			)
			return false
		}
	}
	return true
}

// ServeRequest handles an inbound REST API request as specified by the given
// InboundRequest. Handling includes, if applicable, request body validation,
// unmarshaling, execution of endpoint-specific logic, and response marshaling.
// Any errors are dealt with by sending an appropriate status code and response
// body to the client.
func ServeRequest(req InboundRequest) {
	if req.ReqBodySchemaLoader != nil || req.ReqBodyObj != nil {
		if !ReadAndValidateRequestBody(
			req.W,
			req.R,
			req.ReqBodySchemaLoader,
			req.ReqBodyObj,
		) {
			return
		}
	}
	respBodyObj, err := req.EndpointLogic()
	if err != nil {
		switch e := errors.Cause(err).(type) {
		case *meta.ErrAuthentication:
			WriteAPIResponse(req.W, http.StatusUnauthorized, e)
		case *meta.ErrAuthorization:
			WriteAPIResponse(req.W, http.StatusForbidden, e)
		case *meta.ErrBadRequest:
			WriteAPIResponse(req.W, http.StatusBadRequest, e)
		case *meta.ErrNotFound:
			WriteAPIResponse(req.W, http.StatusNotFound, e)
		case *meta.ErrConflict:
			WriteAPIResponse(req.W, http.StatusConflict, e)
		case *meta.ErrNotSupported:
			WriteAPIResponse(req.W, http.StatusNotImplemented, e)
		case *meta.ErrInternalServer:
			WriteAPIResponse(req.W, http.StatusInternalServerError, e)
		default:
			log.Println(err)
			WriteAPIResponse(
				req.W,
				http.StatusInternalServerError,
				&meta.ErrInternalServer{},
			)
		}
		return
	}
	WriteAPIResponse(req.W, req.SuccessCode, respBodyObj)
}

// WriteAPIResponse sends a response to the client with the specified HTTP
// status code and response body. The response body may be specified as raw
// bytes or as an object which will be marshaled to obtain response body bytes.
func WriteAPIResponse(
	w http.ResponseWriter,
	statusCode int,
	response interface{},
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	responseBody, ok := response.([]byte)
	if !ok {
		var err error
		if responseBody, err = json.Marshal(response); err != nil {
			log.Println(errors.Wrap(err, "error marshaling response body"))
		}
	}
	if _, err := w.Write(responseBody); err != nil {
		log.Println(errors.Wrap(err, "error writing response body"))
	}
}

// ServeWebUIRequest handles an inbound non-API (i.e. web UI) request,
// presumably from a human user and not an API client, as specified by the given
// ServeWebUIRequest. Handling includes execution of endpoint-specific logic.
// Any errors are dealt with by sending an appropriate status code and response
// body to the client.
//
// TODO: Figure out where this can be moved to, because it doesn't belong in
// the restmachinery package, although it isn't immediately clear how to cleanly
// separate this.
func ServeWebUIRequest(req InboundWebUIRequest) {
	respBodyObj, err := req.EndpointLogic()
	if err != nil {
		switch e := errors.Cause(err).(type) {
		case *meta.ErrAuthentication:
			http.Error(req.W, e.Error(), http.StatusUnauthorized)
		case *meta.ErrAuthorization:
			http.Error(req.W, e.Error(), http.StatusForbidden)
		case *meta.ErrBadRequest:
			http.Error(req.W, e.Error(), http.StatusBadRequest)
		case *meta.ErrNotFound:
			http.Error(req.W, e.Error(), http.StatusNotFound)
		case *meta.ErrConflict:
			http.Error(req.W, e.Error(), http.StatusConflict)
		case *meta.ErrNotSupported:
			http.Error(req.W, e.Error(), http.StatusNotImplemented)
		case *meta.ErrInternalServer:
			http.Error(req.W, e.Error(), http.StatusInternalServerError)
		default:
			log.Println(e)
			http.Error(
				req.W,
				(&meta.ErrInternalServer{}).Error(),
				http.StatusInternalServerError,
			)
		}
		return
	}
	req.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	req.W.WriteHeader(req.SuccessCode)
	var responseBody []byte
	switch r := respBodyObj.(type) {
	case []byte:
		responseBody = r
	case string:
		responseBody = []byte(r)
	case fmt.Stringer:
		responseBody = []byte(r.String())
	}
	if _, err := req.W.Write(responseBody); err != nil {
		log.Println(errors.Wrap(err, "error writing response body"))
	}
}
