package webhook

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

// mock8s provides a mock for libk8s.
const mock8s = "../../testdata/mock8s.js"

func TestExecScripts(t *testing.T) {
	ph := &PushHook{}
	script := []byte(`console.log('loaded')`)
	if err := execScripts(ph, "", script); err != nil {
		t.Fatal(err)
	}
}

func mustReadScript(t *testing.T, filename string) []byte {
	script, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	return script
}

func TestExecScripts_Runner(t *testing.T) {
	// This test essentially does a parsing chek on runner.js, too.
	ref := "c0ff334411"
	ph := &PushHook{
		Ref: ref,
		Repository: Repository{
			Name:        "repo",
			FullName:    "owner/repo",
			Description: "Test fixture",
			CloneURL:    "https://example.com/clone",
			SSHURL:      "ssh://git@example.com/clone",
			GitURL:      "git://git@example.com/clone",
			Owner: User{
				Name:     "owner",
				Email:    "owner@example.com",
				Username: "owner",
			},
		},
		HeadCommit: Commit{
			Id: ref,
		},
		Commits: []Commit{
			{
				Id: ref,
			},
		},
		Pusher: User{
			Name:     "owner",
			Email:    "owner@example.com",
			Username: "owner",
		},
	}

	tests := []struct {
		name   string
		script []byte
		sshKey string
		fail   bool
	}{
		{"log", []byte(`console.log(sshKey)`), "foo", false},
		{"basic", mustReadScript(t, "testdata/job_no_sshkey.js"), "", false},
		{"with-sshkey", mustReadScript(t, "testdata/job_sshkey.js"), "my-ssh-key", false},
		{"waitgroup", mustReadScript(t, "testdata/waitgroup.js"), "", false},
	}

	mock := mustReadScript(t, mock8s)
	runner := mustReadScript(t, filepath.Join("..", "..", runnerJS))
	for _, tt := range tests {
		if err := execScripts(ph, tt.sshKey, mock, runner, tt.script); err != nil {
			if tt.fail {
				continue
			}
			t.Fatalf("Script %s failed with : %s", tt.name, err)
		} else if tt.fail {
			t.Errorf("Expected test %s to fail.", tt.name)
		}
	}
}
