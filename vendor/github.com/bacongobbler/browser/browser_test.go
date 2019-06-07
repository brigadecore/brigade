package browser

import (
	"net/url"
	"reflect"
	"testing"
)

type TestCase struct {
	Name            string
	Input           string
	ExpectedErrType reflect.Type
}

func TestInvalidURL(t *testing.T) {
	tests := []TestCase{
		{
			Name:            "random string",
			Input:           "this is not a URL",
			ExpectedErrType: reflect.TypeOf(&url.Error{}),
		},
		{
			Name:            "open a file",
			Input:           "LICENSE",
			ExpectedErrType: reflect.TypeOf(&url.Error{}),
		},
	}
	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			if err := Open(tc.Input); reflect.TypeOf(err) != tc.ExpectedErrType {
				t.Errorf("expected err type '%s', got '%s'", tc.ExpectedErrType, reflect.TypeOf(err))
			}
		})
	}
}
