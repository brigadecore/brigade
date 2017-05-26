package lib

import (
	"path/filepath"
	"testing"
)

func TestScripts(t *testing.T) {
	jsFiles, err := filepath.Glob("*.js")
	if err != nil {
		t.Fatal(err)
	}

	if len(jsFiles) != len(Scripts) {
		t.Fatalf("Expected %d (*.js) scripts, got %d (Scripts)", len(jsFiles), len(Scripts))
	}
}
