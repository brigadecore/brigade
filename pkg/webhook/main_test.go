package webhook

import (
	"os"
	"testing"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/brigadecore/brigade/pkg/storage"

	gin "gopkg.in/gin-gonic/gin.v1"
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

func newTestStore() *testStore {
	return &testStore{
		proj: &brigade.Project{
			Name:         "baxterthehacker/public-repo",
			SharedSecret: "asdf",
		},
	}
}
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
