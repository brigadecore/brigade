package lib

import (
	"testing"
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

		// TODO: Import Otto and run a test parse on each script.
	}
}
