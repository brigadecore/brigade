package meta

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrAuthentication(t *testing.T) {
	testCases := []struct {
		name       string
		err        *ErrAuthentication
		assertions func(err *ErrAuthentication, errStr string)
	}{
		{
			name: "without reason",
			err:  &ErrAuthentication{},
			assertions: func(err *ErrAuthentication, errStr string) {
				require.Contains(t, errStr, "Could not authenticate the request")
			},
		},
		{
			name: "with reason",
			err: &ErrAuthentication{
				Reason: "i don't have to answer to you",
			},
			assertions: func(err *ErrAuthentication, errStr string) {
				require.Contains(t, errStr, "Could not authenticate the request")
				require.Contains(t, errStr, err.Reason)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			errStr := testCase.err.Error()
			testCase.assertions(testCase.err, errStr)
		})
	}
}

func TestErrAuthenticationMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, &ErrAuthentication{}, "AuthenticationError")
}

func TestErrAuthorization(t *testing.T) {
	testCases := []struct {
		name       string
		err        *ErrAuthorization
		assertions func(err *ErrAuthorization, errStr string)
	}{
		{
			name: "without reason",
			err:  &ErrAuthorization{},
			assertions: func(err *ErrAuthorization, errStr string) {
				require.Contains(t, errStr, "The request is not authorized")
			},
		},
		{
			name: "with reason",
			err: &ErrAuthorization{
				Reason: "i don't have to answer to you",
			},
			assertions: func(err *ErrAuthorization, errStr string) {
				require.Contains(t, errStr, "The request is not authorized")
				require.Contains(t, errStr, err.Reason)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			errStr := testCase.err.Error()
			testCase.assertions(testCase.err, errStr)
		})
	}
}

func TestErrAuthorizationMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, &ErrAuthorization{}, "AuthorizationError")
}

func TestErrBadRequest(t *testing.T) {
	testCases := []struct {
		name       string
		err        *ErrBadRequest
		assertions func(err *ErrBadRequest)
	}{
		{
			name: "without details",
			err: &ErrBadRequest{
				Reason: "i don't have to answer to you",
			},
			assertions: func(err *ErrBadRequest) {
				require.Contains(t, err.Error(), err.Reason)
				for _, detail := range err.Details {
					require.NotContains(t, err.Error(), detail)
				}
			},
		},
		{
			name: "with details",
			err: &ErrBadRequest{
				Reason:  "i don't have to answer to you",
				Details: []string{"the", "devil", "is", "in", "the", "details"},
			},
			assertions: func(err *ErrBadRequest) {
				require.Contains(t, err.Error(), err.Reason)
				for _, detail := range err.Details {
					require.Contains(t, err.Error(), detail)
				}
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.assertions(testCase.err)
		})
	}
}

func TestErrBadRequestMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, &ErrBadRequest{}, "BadRequestError")
}

func TestErrNotFound(t *testing.T) {
	err := &ErrNotFound{
		Type: "User",
		ID:   "tony@starkindustries.com",
	}
	require.Contains(t, err.Error(), "not found")
	require.Contains(t, err.Error(), err.Type)
	require.Contains(t, err.Error(), err.ID)
}

func TestErrNotFoundMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, &ErrNotFound{}, "NotFoundError")
}

func TestErrConflict(t *testing.T) {
	err := &ErrConflict{
		Type:   "User",
		ID:     "tony@starkindustries.com",
		Reason: "i don't have to answer to you",
	}
	require.Contains(t, err.Error(), err.Reason)
}

func TestErrConflictMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, &ErrConflict{}, "ConflictError")
}

func TestErrInternalServer(t *testing.T) {
	err := &ErrInternalServer{}
	require.Contains(t, err.Error(), "internal server error")
}

func TestErrInternalServerMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, &ErrInternalServer{}, "InternalServerError")
}

func TestErrNotSupported(t *testing.T) {
	err := &ErrNotSupported{
		Details: "i don't have to answer to you",
	}
	require.Contains(t, err.Error(), err.Details)
}

func TestErrNotSupportedMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, &ErrNotSupported{}, "NotSupportedError")
}
