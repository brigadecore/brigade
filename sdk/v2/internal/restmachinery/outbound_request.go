package restmachinery

// OutboundRequest models of an outbound API call.
type OutboundRequest struct {
	// Method specifies the HTTP method to be used.
	Method string
	// Path specifies a path (relative to the root of the API) to be used.
	Path string
	// QueryParams optionally specifies any URL query parameters to be used.
	QueryParams map[string]string
	// AuthHeaders optionally specifies any authentication headers to be used.
	AuthHeaders map[string]string
	// Headers optionally specifies any miscellaneous HTTP headers to be used.
	Headers map[string]string
	// ReqBodyObj optionally provides an object that can be marshaled to create
	// the body of the HTTP request.
	ReqBodyObj interface{}
	// SuccessCode specifies what HTTP response code should indicate a successful
	// API call.
	SuccessCode int
	// RespObj optionally provides an object into which the HTTP response body can
	// be unmarshaled.
	RespObj interface{}
}
