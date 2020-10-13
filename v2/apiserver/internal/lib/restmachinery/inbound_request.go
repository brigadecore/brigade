package restmachinery

import (
	"net/http"

	"github.com/xeipuuv/gojsonschema"
)

// InboundRequest represents an inbound REST API request.
type InboundRequest struct {
	// W is the http.ResponseWriter that can be used to send a response to the
	// client.
	W http.ResponseWriter
	// R is the raw, inbound *http.Request.
	R *http.Request
	// ReqBodySchemaLoader optionally specifies a gojsonschema.JSONLoader that,
	// when applicable, can be used to validate JSON in the request body.
	ReqBodySchemaLoader gojsonschema.JSONLoader
	// ReqBodyObj optionally specifies an object into which, when applicable, the
	// JSON in the request body will be unmarshaled.
	ReqBodyObj interface{}
	// EndpointLogic implements endpoint-specific logic. Typically it will only
	// invoke a transport-agnostic function of the service layer.
	EndpointLogic func() (interface{}, error)
	// SuccessCode optionally specifies the HTTP status code for the response to
	// the client when the EndpointLogic has executed without error. If not
	// specified (i.e. left 0), 200 (OK) is assumed.
	SuccessCode int
}

// InboundWebUIRequest represents an inbound non-API (i.e. web UI) request,
// presumably from a human user and not an API client.
//
// TODO: Figure out where this can be moved to, because it doesn't belong in
// the restmachinery package, although it isn't immediately clear how to cleanly
// separate this.
type InboundWebUIRequest struct {
	W             http.ResponseWriter
	EndpointLogic func() (interface{}, error)
	SuccessCode   int
}
