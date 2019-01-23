package api

import (
	"testing"

	"github.com/Azure/brigade/pkg/storage/mock"
)

func TestGetBuildSummariesForProjects(t *testing.T) {
	project := &Project{
		store: mock.New(),
	}

	projects, err := project.store.GetProjects()
	if err != nil {
		t.Fatalf("error: %s", err.Error())
	}
	projectSummaries := project.getBuildSummariesForProjects(projects)

	if projectSummaries[0].LastBuild.ID != "build-id1" {
		t.Fatal("wrong BuildID in getBuildSummariesForProjects")
	}
}
