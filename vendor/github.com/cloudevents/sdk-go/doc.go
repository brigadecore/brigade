/*
Package cloudevents provides primitives to work with CloudEvents specification: https://github.com/cloudevents/spec.


Parsing Event from HTTP Request:
	// req is *http.Request
	event, err := cloudEvents.FromHTTPRequest(req)
	if err != nil {
		panic("Unable to parse event from http Request: " + err.String())
	}


Creating a minimal CloudEvent in version 0.1:
    import "github.com/cloudevents/sdk-go/v01"
	event := v01.Event{
		EventType:        "com.example.file.created",
		Source:           "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
		EventID:          "ea35b24ede421",
	}


The goal of this package is to provide support for all released versions of CloudEvents, ideally while maintaining
the same API. It will use semantic versioning with following rules:
* MAJOR version increments when backwards incompatible changes is introduced.
* MINOR version increments when backwards compatible feature is introduced INCLUDING support for new CloudEvents version.
* PATCH version increments when a backwards compatible bug fix is introduced.
*/
package cloudevents
