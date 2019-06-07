package cloudevents

import (
	"net/http"
	"reflect"
)

// HTTPMarshaller an interface with methods for creating CloudEvents
type HTTPMarshaller interface {
	FromRequest(req *http.Request) (Event, error)
	ToRequest(req *http.Request, event Event) error
}

// HTTPCloudEventConverter an interface for defining converters that can read/write CloudEvents from HTTP requests
type HTTPCloudEventConverter interface {
	CanRead(t reflect.Type, mediaType string) bool
	CanWrite(t reflect.Type, mediaType string) bool
	Read(t reflect.Type, req *http.Request) (Event, error)
	Write(t reflect.Type, req *http.Request, event Event) error
}
