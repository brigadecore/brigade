# Go SDK for [CloudEvents](https://github.com/cloudevents/spec)

[![go-doc](https://godoc.org/github.com/cloudevents/sdk-go?status.svg)](https://godoc.org/github.com/cloudevents/sdk-go)

**NOTE: This SDK is still considered work in progress, things might (and will) break with every update.**

## New for v0.2
For this release extensions have been moved to top level properties. Previously extensions were defined in an extensions map, which was itself a top level property. All CloudEvent properties can be accessed using the generic Get method, or the type checked versions, e.g. GetString, GetMap, etc., but only the well known properties allow for direct field access. The marshallers handle packing and unpacking the extensions into an internal map.

This release also makes significant changes to the CloudEvent property names. All property names on the wire are now lower case with no separator characters. This ensures that these names are recognized across transports, which have different standards for property names. This release also removes the redundant 'event' prefix on property names. So EventType becomes Type and EventID become ID, etc. One special case is CloudEventsVersion, which becomes SpecVersion. 

## Working with CloudEvents
Package cloudevents provides primitives to work with CloudEvents specification: https://github.com/cloudevents/spec.

Parsing Event from HTTP Request:
```go
import "github.com/cloudevents/sdk-go"
	marshaller := v02.NewDefaultHTTPMarshaller()
	// req is *http.Request
	event, err := marshaller.FromRequest(req)
	if err != nil {
		panic("Unable to parse event from http Request: " + err.String())
	}
	fmt.Printf("type: %s", event.Get("type")
```

Creating a minimal CloudEvent in version 0.2:
```go
import "github.com/cloudevents/sdk-go/v02"
	event := v02.Event{
		Type:        "com.example.file.created",
		Source:           "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
		ID:          "ea35b24ede421",
	}
```

Creating HTTP request from CloudEvent:
```
marshaller := v02.NewDefaultHTTPMarshaller()
var req *http.Request
err := marshaller.ToRequest(req)
if err != nil {
	panic("Unable to marshal event into http Request: " + err.String())
}
```

The goal of this package is to provide support for all released versions of CloudEvents, ideally while maintaining
the same API. It will use semantic versioning with following rules:
* MAJOR version increments when backwards incompatible changes is introduced.
* MINOR version increments when backwards compatible feature is introduced INCLUDING support for new CloudEvents version.
* PATCH version increments when a backwards compatible bug fix is introduced.


## TODO list

- [ ] Add encoders registry, where SDK user can register their custom content-type encoders/decoders
- [ ] Add more tests for edge cases

## Existing Go for CloudEvents

Existing projects that added support for CloudEvents in Go are listed below. It's our goal to identify existing patterns
of using CloudEvents in Go-based project and design the SDK to support these patterns (where it makes sense).
- https://github.com/knative/pkg/tree/master/cloudevents
- https://github.com/vmware/dispatch/blob/master/pkg/events/cloudevent.go
- https://github.com/serverless/event-gateway/tree/master/event
