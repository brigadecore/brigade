package cloudevents

// RequiredPropertyError is return when a property of an event that is required by specification is not set
type RequiredPropertyError string

func (e RequiredPropertyError) Error() string {
	return "missing required property " + string(e)
}

// VersionMismatchError is returned when expected CloudEvent version does not match the actual one, e.g.
// when using GetEventV01 or when using transport bindings.
type VersionMismatchError string

func (e VersionMismatchError) Error() string {
	return "provided event is not CloudEvent or does not implement expected version: " + string(e)
}

// VersionNotSupportedError is returned when provided version is not supported by this library.
type VersionNotSupportedError string

func (e VersionNotSupportedError) Error() string {
	return "provided version " + string(e) + " is not supported"
}

// ContentTypeNotSupportedError is returned when povided event's content type is not supported by this library.
type ContentTypeNotSupportedError string

func (e ContentTypeNotSupportedError) Error() string {
	return "provided content type " + string(e) + " is not supported"
}

// IllegalArgumentError is returned when an argument passed to a method is an illegal value
type IllegalArgumentError string

func (e IllegalArgumentError) Error() string {
	return "argument " + string(e) + "is illegal"
}
