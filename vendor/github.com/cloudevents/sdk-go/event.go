package cloudevents

import (
	"net/url"
	"time"
)

// Version01 holds a version string for CloudEvents specification version 0.1. See also EventV01 interface
// https://github.com/cloudevents/spec/blob/v0.1/spec.md
const Version01 = "0.1"
const Version02 = "0.2"

// Event interface is a generic abstraction over all possible versions and implementations of CloudEvents.
type Event interface {
	// CloudEventVersion returns the version of Event specification followed by the underlying implementation.
	CloudEventVersion() string
	// Get takes a property name and, if it exists, returns the value of that property. The ok return value can
	// be used to verify if the property exists.
	Get(property string) (value interface{}, ok bool)
	// GetInt is a convenience method that wraps Get to provide a type checked return value. Ok will be false
	// if the property does not exist or the value cannot be converted to an int32.
	GetInt(property string) (value int32, ok bool)
	// GetString is a convenience method that wraps Get to provide a type checked return value. Ok will be false
	// if the property does not exist or the value cannot be converted to a string.
	GetString(property string) (value string, ok bool)
	// GetBinary is a convenience method that wraps Get to provide a type checked return value. Ok will be false
	// if the property does not exist or the value cannot be converted to a binary array.
	GetBinary(property string) (value []byte, ok bool)
	// GetMap is a convenience method that wraps Get to provide a type checked return value. Ok will be false
	// if the property does not exist or the value cannot be converted to a map.
	GetMap(propery string) (value map[string]interface{}, ok bool)
	// GetTime is a convenience method that wraps Get to provide a type checked return value. Ok will be false
	// if the property does not exist or the value cannot be converted or parsed into a time.Time.
	GetTime(property string) (value *time.Time, ok bool)
	// GetURL is a convenience method that wraps Get to provide a type checked return value. Ok will be false
	// if the property does not exist or the value cannot be converted or parsed into a url.URL.
	GetURL(property string) (value url.URL, ok bool)
	// Set sets the property value
	Set(property string, value interface{})
	// Properties returns a map of all event properties as keys and their mandatory status as values
	Properties() map[string]bool
}
