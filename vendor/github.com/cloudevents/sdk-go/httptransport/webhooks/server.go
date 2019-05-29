package webhooks

import "net/http"

// WebHook is an HTTP middleware handling webhook requests according to HTTP Webhook spec
type WebHook struct {
}

// ServeHTTP implements http.Handler interface
func (w *WebHook) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	panic("not implemented")
}
