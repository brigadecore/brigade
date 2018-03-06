package mock

import (
	"reflect"
	"testing"

	"github.com/uswitch/brigade/pkg/brigade"
	"github.com/uswitch/brigade/pkg/storage"
)

func TestStore(t *testing.T) {
	m := New()

	// Make sure we implement the interface.
	var _ storage.Store = m

	assertSame := func(label string, a, b interface{}) {
		if !reflect.DeepEqual(a, b) {
			t.Errorf("failed equality for %s", label)
		}
	}
	assertSame("project", StubProject, m.Project)
	assertSame("worker", StubWorker, m.Worker)
	assertSame("build", StubBuild, m.Build)
	assertSame("job", StubJob, m.Job)
	assertSame("log data", StubLogData, m.LogData)

	// Exercise the methods, too.
	p, _ := m.GetProjects()
	assertSame("GetProjects", []*brigade.Project{StubProject}, p)

	p2, _ := m.GetProject(StubProject.ID)
	assertSame("GetProject", StubProject, p2)

	b1, _ := m.GetProjectBuilds(StubProject)
	assertSame("GetProjectBuilds", StubBuild, b1[0])

	b2, _ := m.GetBuilds()
	assertSame("GetBuilds", StubBuild, b2[0])

	b3, _ := m.GetBuild(StubBuild.ID)
	assertSame("GetBuild", StubBuild, b3)

	j1, _ := m.GetBuildJobs(StubBuild)
	assertSame("GetBuildJobs", StubJob, j1[0])

	j2, _ := m.GetJob(StubJob.ID)
	assertSame("GetJob", StubJob, j2)

	w1, _ := m.GetWorker(StubBuild.ID)
	assertSame("GetWorker", StubWorker, w1)

	jl, _ := m.GetJobLog(StubJob)
	assertSame("GetJobLog", StubLogData, jl)

	wl, _ := m.GetWorkerLog(StubWorker)
	assertSame("GetJobLog", StubLogData, wl)
}
