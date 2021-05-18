package meta

import (
	"encoding/json"
	"fmt"
)

// ErrAuthentication represents an error asserting an principal's identity.
type ErrAuthentication struct {
	// Reason is a natural language explanation for why authentication failed.
	Reason string `json:"reason,omitempty"`
}

func (e *ErrAuthentication) Error() string {
	if e.Reason == "" {
		return "Could not authenticate the request."
	}
	return fmt.Sprintf("Could not authenticate the request: %s", e.Reason)
}

// MarshalJSON amends ErrAuthentication instances with type metadata.
func (e ErrAuthentication) MarshalJSON() ([]byte, error) {
	type Alias ErrAuthentication
	return json.Marshal(
		struct {
			TypeMeta `json:",inline"`
			Alias    `json:",inline"`
		}{
			TypeMeta: TypeMeta{
				APIVersion: APIVersion,
				Kind:       "AuthenticationError",
			},
			Alias: (Alias)(e),
		},
	)
}

// ErrAuthorization represents an error wherein a principal was not authorized
// to perform the requested operation.
type ErrAuthorization struct {
	// Reason is a natural language explanation for why authorization failed.
	Reason string `json:"reason,omitempty"`
}

func (e *ErrAuthorization) Error() string {
	if e.Reason == "" {
		return "The request is not authorized."
	}
	return fmt.Sprintf("The request is not authorized: %s", e.Reason)
}

// MarshalJSON amends ErrAuthorization instances with type metadata.
func (e ErrAuthorization) MarshalJSON() ([]byte, error) {
	type Alias ErrAuthorization
	return json.Marshal(
		struct {
			TypeMeta `json:",inline"`
			Alias    `json:",inline"`
		}{
			TypeMeta: TypeMeta{
				APIVersion: APIVersion,
				Kind:       "AuthorizationError",
			},
			Alias: (Alias)(e),
		},
	)
}

// ErrBadRequest represents an error wherein an invalid request has been
// rejected by the API server.
type ErrBadRequest struct {
	// Reason is a natural language explanation for why the request is invalid.
	Reason string `json:"reason,omitempty"`
	// Details may further qualify why a request is invalid. For instance, if
	// the Reason field states that request validation failed, the Details field,
	// may enumerate specific request schema violations.
	Details []string `json:"details,omitempty"`
}

func (e *ErrBadRequest) Error() string {
	if len(e.Details) == 0 {
		return fmt.Sprintf("Bad request: %s", e.Reason)
	}
	msg := fmt.Sprintf("Bad request: %s:", e.Reason)
	for i, detail := range e.Details {
		msg = fmt.Sprintf("%s\n  %d. %s", msg, i, detail)
	}
	return msg
}

// MarshalJSON amends ErrBadRequest instances with type metadata.
func (e ErrBadRequest) MarshalJSON() ([]byte, error) {
	type Alias ErrBadRequest
	return json.Marshal(
		struct {
			TypeMeta `json:",inline"`
			Alias    `json:",inline"`
		}{
			TypeMeta: TypeMeta{
				APIVersion: APIVersion,
				Kind:       "BadRequestError",
			},
			Alias: (Alias)(e),
		},
	)
}

// ErrNotFound represents an error wherein a resource presumed to exist could
// not be located.
type ErrNotFound struct {
	// Type identifies the type of the resource that could not be located.
	Type string `json:"type,omitempty"`
	// ID is the identifier of the resource of type Type that could not be
	// located.
	ID string `json:"id,omitempty"`
	// Reason is a natural language explanation around why the resource could not
	// be located.
	Reason string `json:"reason,omitempty"`
}

func (e *ErrNotFound) Error() string {
	if e.Type == "" && e.ID == "" && e.Reason != "" {
		return e.Reason
	}

	msg := fmt.Sprintf("%s %q not found", e.Type, e.ID)
	if e.Reason != "" {
		return msg + fmt.Sprintf(": %s", e.Reason)
	}
	return msg + "."
}

// MarshalJSON amends ErrNotFound instances with type metadata.
func (e ErrNotFound) MarshalJSON() ([]byte, error) {
	type Alias ErrNotFound
	return json.Marshal(
		struct {
			TypeMeta `json:",inline"`
			Alias    `json:",inline"`
		}{
			TypeMeta: TypeMeta{
				APIVersion: APIVersion,
				Kind:       "NotFoundError",
			},
			Alias: (Alias)(e),
		},
	)
}

// ErrConflict represents an error wherein a request cannot be completed because
// it would violate some constrain of the system, for instance creating a new
// resource with an identifier already used by another resource of the same
// type.
type ErrConflict struct {
	// Type identifies the type of the resource that the conflict applies to.
	Type string `json:"type,omitempty"`
	// ID is the identifier of the resource that has encountered a conflict.
	ID string `json:"id,omitempty"`
	// Reason is a natural language explanation of the conflict.
	Reason string `json:"reason,omitempty"`
}

func (e *ErrConflict) Error() string {
	return e.Reason
}

// MarshalJSON amends MarshalJSON instances with type metadata.
func (e ErrConflict) MarshalJSON() ([]byte, error) {
	type Alias ErrConflict
	return json.Marshal(
		struct {
			TypeMeta `json:",inline"`
			Alias    `json:",inline"`
		}{
			TypeMeta: TypeMeta{
				APIVersion: APIVersion,
				Kind:       "ConflictError",
			},
			Alias: (Alias)(e),
		},
	)
}

// ErrInternalServer represents a condition wherein the API server has
// encountered an unexpected error and does not wish to communicate further
// details of that error to the client.
type ErrInternalServer struct{}

func (e *ErrInternalServer) Error() string {
	return "An internal server error occurred."
}

// MarshalJSON amends ErrInternalServer instances with type metadata.
func (e ErrInternalServer) MarshalJSON() ([]byte, error) {
	type Alias ErrInternalServer
	return json.Marshal(
		struct {
			TypeMeta `json:",inline"`
			Alias    `json:",inline"`
		}{
			TypeMeta: TypeMeta{
				APIVersion: APIVersion,
				Kind:       "InternalServerError",
			},
			Alias: (Alias)(e),
		},
	)
}

// ErrNotSupported represents an error wherein a request cannot be completed
// because the API server explicitly does not support it. This can occur, for
// instance, if a client attempts to authenticate to the API server using an
// authentication mechanism that is explicitly disabled by the API server.
type ErrNotSupported struct {
	// Details is a natural language explanation of why the request was is not
	// supported by the API server.
	Details string `json:"reason,omitempty"`
}

func (e *ErrNotSupported) Error() string {
	return e.Details
}

// MarshalJSON amends ErrNotSupported instances with type metadata.
func (e ErrNotSupported) MarshalJSON() ([]byte, error) {
	type Alias ErrNotSupported
	return json.Marshal(
		struct {
			TypeMeta `json:",inline"`
			Alias    `json:",inline"`
		}{
			TypeMeta: TypeMeta{
				APIVersion: APIVersion,
				Kind:       "NotSupportedError",
			},
			Alias: (Alias)(e),
		},
	)
}
