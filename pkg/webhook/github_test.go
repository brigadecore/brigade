package webhook

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/github"
	"gopkg.in/gin-gonic/gin.v1"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
)

type testStore struct {
	proj   *brigade.Project
	builds []*brigade.Build
	err    error
	storage.Store
}

func (s *testStore) GetProject(name string) (*brigade.Project, error) {
	return s.proj, s.err
}

func (s *testStore) CreateBuild(build *brigade.Build) error {
	s.builds = append(s.builds, build)
	return s.err
}

func TestGithubHandler(t *testing.T) {
	store := &testStore{
		proj: &brigade.Project{
			Name:         "baxterthehacker/public-repo",
			SharedSecret: "asdf",
		},
	}

	s := NewGithubHook(store, false)
	s.getFile = func(commit, path string, proj *brigade.Project) ([]byte, error) {
		t.Logf("Getting file %s, for commit %s", path, commit)
		return []byte(""), nil
	}
	s.createStatus = func(commit string, proj *brigade.Project, status *github.RepoStatus) error {
		t.Logf("Creating status %v, for commit %s", status, commit)
		return nil
	}

	payload, err := ioutil.ReadFile("testdata/github-push-payload.json")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r, err := http.NewRequest("POST", "", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Add("X-GitHub-Event", "push")
	r.Header.Add("X-Hub-Signature", SHA1HMAC([]byte("asdf"), payload))

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = r

	s.Handle(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("unexpected error: %d\n%s", w.Code, w.Body.String())
	}
}

func TestTruncAt(t *testing.T) {
	if "foo" != truncAt("foo", 100) {
		t.Fatal("modified string that was fine.")
	}

	if got := truncAt("foobar", 6); got != "foobar" {
		t.Errorf("Unexpected truncation of foobar: %s", got)
	}

	if got := truncAt("foobar1", 6); got != "foo..." {
		t.Errorf("Unexpected truncation of foobar1: %s", got)
	}
}
