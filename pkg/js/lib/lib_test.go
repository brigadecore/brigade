package lib

import (
	"testing"

	"github.com/robertkrimen/otto/parser"
)

func TestScript(t *testing.T) {
	for _, script := range []string{
		"js/run.js",
		"js/run_mock.js",
		"js/job.js",
		"js/event.js",
		"js/waitgroup.js",
		"js/runner.js",
	} {
		b, err := Script(script)
		if err != nil {
			t.Fatal(err)
		}
		if len(b) == 0 {
			t.Error("Expected script to have contents. Got empty []byte.")
		}

		// Just ensure that the JS parses. The code in the files is tested
		// in pkg/js.
		if _, err := parser.ParseFile(nil, "", b, 0); err != nil {
			t.Errorf("parse error on %q: %s", script, err)
		}
	}
}
