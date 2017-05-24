package webhook

import (
	"bytes"
	"testing"
)

// Canary
var _ StatusSetter = &GitHubNotifier{}
var _ StatusSetter = &mockStatusSetter{}

type mockStatusSetter struct {
	owner, repo, sha  string
	state             StatusCode
	ctx, desc, target string
}

func newMockStatusSetter() *mockStatusSetter {
	return &mockStatusSetter{}
}

func (m *mockStatusSetter) SetTarget(owner, repo, sha string) error {
	m.owner = owner
	m.repo = repo
	m.sha = sha
	return nil
}

func (m *mockStatusSetter) SetStatus(state StatusCode, ctx, desc, target string) error {
	m.ctx, m.desc, m.target = ctx, desc, target
	return nil
}

func TestGitHubNotifier_url(t *testing.T) {
	expect := "https://api.github.com/repos/technosophos/zolver/statuses/c0ff33"
	g := NewGitHubNotifier("chocolate chip pancakes")
	g.SetTarget("technosophos", "zolver", "c0ff33")
	if got := g.url(); got != expect {
		t.Errorf("Expected %q, got %q", expect, got)
	}
}

func TestGitHubNotifier_body(t *testing.T) {
	expect := `{"state":"failure","target_url":"a","description":"b","context":"c"}`
	g := NewGitHubNotifier("eggs benedict")
	data, err := g.body(GitHubStatus{StatusFailure, "a", "b", "c"})

	if err != nil {
		t.Fatal(err)
	}

	if got := data.(*bytes.Buffer).String(); got != expect {
		t.Errorf("Expected:\n%s\nGot:\n%s", expect, got)
	}
}
