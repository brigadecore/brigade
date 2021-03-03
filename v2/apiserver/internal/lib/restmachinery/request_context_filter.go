package restmachinery

import (
	"context"
	"net/http"
)

// requestContextFilter is a filter that can decorate an HTTP handler function
// such that requests are examined and IFF the request utilizes an HTTP method
// that commonly mutates the system (for instance, POST commonly is used for
// creates and PUT and PATCH are commonly used for updates), then the request's
// context is REPLACED with the background context. The usefulness of this lies
// in creating a guarantee that operations that mutate the system are
// uninterruptible by impatient clients that might hang up in the middle of a
// request that is running long, perhaps due to retry logic with exponential
// backoff. The client can hang up if they wish, but operations that mutate the
// system will continue in the background until they have succeeded or failed
// independently of any client action. This filter, when used, MUST be the first
// (outermost) filter in the chain of filters/handlers. By being first in the
// chain, this filter does not need to be concerned with discovering and copying
// values that may have been poked into the context by other filters (for
// example, a principal object added to the context by an authentication
// filter).
type requestContextFilter struct{}

func (r *requestContextFilter) Decorate(
	handle http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodDelete:
			fallthrough
		case http.MethodPatch:
			fallthrough
		case http.MethodPost:
			fallthrough
		case http.MethodPut:
			handle(w, req.WithContext(context.Background()))
		default:
			handle(w, req)
		}
	}
}
