package commands

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// ensure namespace is not set by local environment.
	globalNamespace = "default"

	os.Exit(m.Run())
}
