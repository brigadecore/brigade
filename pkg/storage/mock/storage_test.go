package mock

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
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

	jls, _ := m.GetJobLogStream(StubJob)
	bjls := new(bytes.Buffer)
	bjls.ReadFrom(jls)
	assertSame("GetJobLogStream", StubLogData, bjls.String())

	jlsf, _ := m.GetJobLogStreamFollow(StubJob)
	bjlsf := new(bytes.Buffer)
	bjlsf.ReadFrom(jlsf)
	assertSame("GetJobLogStreamFollow", StubLogData, bjlsf.String())

	wl, _ := m.GetWorkerLog(StubWorker)
	assertSame("GetWorkerLog", StubLogData, wl)

	wls, _ := m.GetWorkerLogStream(StubWorker)
	bwls := new(bytes.Buffer)
	bwls.ReadFrom(wls)
	assertSame("GetWorkerLogStream", StubLogData, bwls.String())

	wlsf, _ := m.GetWorkerLogStreamFollow(StubWorker)
	bwlsf := new(bytes.Buffer)
	bwlsf.ReadFrom(wlsf)
	assertSame("GetWorkerLogStreamFollow", StubLogData, bwlsf.String())

	if !m.BlockUntilAPICacheSynced(nil) {
		t.Fatal("expected to return true")
	}
}
