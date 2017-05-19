package webhook

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestExecScripts(t *testing.T) {
	ph := &PushHook{}
	script := []byte(`console.log('loaded')`)
	if err := execScripts(ph, script); err != nil {
		t.Fatal(err)
	}
}

func TestExecScripts_Runner(t *testing.T) {
	// This test essentially does a parsing chek on runner.js, too.
	ph := &PushHook{}
	script := []byte(`console.log(secName)`)

	runner, err := ioutil.ReadFile(filepath.Join("..", "..", runnerJS))
	if err != nil {
		t.Fatal(err)
	}

	if err := execScripts(ph, runner, script); err != nil {
		t.Fatal(err)
	}
}
