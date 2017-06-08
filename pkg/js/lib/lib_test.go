package lib

import (
	"testing"
)

func TestScript(t *testing.T) {
	b, err := Script("js/runner.js")
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Error("Expected script to have contents. Got empty []byte.")
	}
}
