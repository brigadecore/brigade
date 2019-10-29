package brigade

import (
	"fmt"
	"io"
	"sort"

	"github.com/brigadecore/brigade/pkg/storage"

	brigademodel "github.com/slok/brigadeterm/pkg/model/brigade"
)

// Service is the service where all the brigade data will be retrieved
// and prepared with the information that this applications needs.
type Service interface {
	// GetProjectBuilds will get one project.
	GetProject(projectID string) (*brigademodel.Project, error)
	// GetProjectLastBuild will get projects last builds.
	GetProjectLastBuilds(projectID string, quantity int) ([]*brigademodel.Build, error)
	// GetProjects will get all the projects that are on brigade.
	GetProjects() ([]*brigademodel.Project, error)
	// GetBuild will get one build.
	GetBuild(buildID string) (*brigademodel.Build, error)
	// GetProjectBuilds will get all the builds of a project in descendant or ascendant order.
	GetProjectBuilds(project *brigademodel.Project, desc bool) ([]*brigademodel.Build, error)
	// RerunBuild will take a buildID and rerun that build. The build needs to exist
	// if the build doesn't exist it will error.
	RerunBuild(buildID string) error
	// GetBuildJobs will get all the jobs of a build in descendant or ascendant order.
	GetBuildJobs(BuildID string, desc bool) ([]*brigademodel.Job, error)
	// GetJob will get a job.
	GetJob(jobID string) (*brigademodel.Job, error)
	// GetJobLog will get a job log.
	GetJobLog(jobID string) (io.ReadCloser, error)
	// GetJobLogStream will get a job log stream that will be updated when new logs
	// are created, in other words this stream is a real time stream of a job log.
	GetJobLogStream(jobID string) (io.ReadCloser, error)
}

// repository will use kubernetes as repository for the brigade objects.
type service struct {
	client storage.Store
}

// NewService returns a new brigade service.
func NewService(brigadestore storage.Store) Service {
	return &service{
		client: brigadestore,
	}
}

func (s *service) GetProject(projectID string) (*brigademodel.Project, error) {
	prj, err := s.client.GetProject(projectID)

	if err != nil {
		return nil, err
	}
	res := brigademodel.Project(*prj)
	return &res, nil
}

func (s *service) GetProjectLastBuilds(projectID string, quantity int) ([]*brigademodel.Build, error) {
	prj, err := s.client.GetProject(projectID)

	if err != nil {
		return nil, err
	}

	// Get the available builds.
	builds, err := s.GetProjectBuilds(prj, true)
	if err != nil {
		return nil, err
	}
	if len(builds) == 0 {
		return nil, fmt.Errorf("no builds available")
	}

	// Get last one.
	if len(builds) > quantity {
		builds = builds[:quantity]
	}
	lastBuilds := make([]*brigademodel.Build, len(builds))

	for i, b := range builds {
		lb := brigademodel.Build(*b)
		lastBuilds[i] = &lb
	}

	return lastBuilds, nil
}

// GetProjects satisfies Service interface.
func (s *service) GetProjects() ([]*brigademodel.Project, error) {
	prjs, err := s.client.GetProjects()

	if err != nil {
		return nil, err
	}

	// Sort projects by name.
	sort.Slice(prjs, func(i, j int) bool {
		return prjs[i].Name < prjs[j].Name
	})

	prjList := make([]*brigademodel.Project, len(prjs))
	for i, prj := range prjs {
		p := brigademodel.Project(*prj)
		prjList[i] = &p
	}

	return prjList, nil
}

func (s *service) GetBuild(buildID string) (*brigademodel.Build, error) {
	bld, err := s.client.GetBuild(buildID)

	if err != nil {
		return nil, err
	}
	res := brigademodel.Build(*bld)
	return &res, nil
}

// GetAllProjects satisfies Service interface.
func (s *service) GetProjectBuilds(project *brigademodel.Project, desc bool) ([]*brigademodel.Build, error) {
	pr, err := s.client.GetProject(project.ID)
	if err != nil {
		return []*brigademodel.Build{}, err
	}

	builds, err := s.client.GetProjectBuilds(pr)
	if err != nil {
		return []*brigademodel.Build{}, err
	}

	res := make([]*brigademodel.Build, len(builds))
	for i, build := range builds {
		bl := brigademodel.Build(*build)
		res[i] = &bl
	}

	// TODO: Should we improve the sorting algorithm?
	// Make a first sort by ID so in equality of time always is the same order on every case.
	// Doesn't matter the order at all is for consistency when start time is the same.
	sort.SliceStable(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})

	// Split phantom jobs and builds.
	phantomBs := []*brigademodel.Build{}
	goodBs := []*brigademodel.Build{}
	for _, b := range res {
		b := b
		if b.Worker == nil {
			phantomBs = append(phantomBs, b)
		} else {
			goodBs = append(goodBs, b)
		}
	}

	// Order builds in descending order (last ones first).
	sort.SliceStable(goodBs, func(i, j int) bool {
		if desc {
			return goodBs[i].Worker.StartTime.After(goodBs[j].Worker.StartTime)
		}
		return goodBs[i].Worker.StartTime.Before(goodBs[j].Worker.StartTime)
	})

	// Append phantom to the end.
	for _, pb := range phantomBs {
		pb := pb
		goodBs = append(goodBs, pb)
	}

	return goodBs, nil
}

// GetBuildJobs satisfies Service interface.
func (s *service) GetBuildJobs(BuildID string, desc bool) ([]*brigademodel.Job, error) {
	bl, err := s.client.GetBuild(BuildID)
	if err != nil {
		return []*brigademodel.Job{}, err
	}

	jobs, err := s.client.GetBuildJobs(bl)
	if err != nil {
		return []*brigademodel.Job{}, err
	}
	res := make([]*brigademodel.Job, len(jobs))
	for i, job := range jobs {
		j := brigademodel.Job(*job)
		res[i] = &j
	}

	// TODO: Should we improve the sorting algorithm?
	// Make a first sort by ID so in equality of time always is the same order on every case.
	// Doesn't matter the order at all is for consistency when start time is the same.
	sort.SliceStable(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})

	// Order jobs in ascending order (first ones first).
	sort.SliceStable(res, func(i, j int) bool {
		if desc {
			return res[i].StartTime.After(res[j].StartTime)
		}
		return res[i].StartTime.Before(res[j].StartTime)
	})

	return res, nil
}

func (s *service) GetJob(jobID string) (*brigademodel.Job, error) {
	j, err := s.client.GetJob(jobID)

	if err != nil {
		return nil, err
	}
	res := brigademodel.Job(*j)
	return &res, nil
}

// GetJobLog satisfies Service interface.
func (s *service) GetJobLog(jobID string) (io.ReadCloser, error) {
	job, err := s.client.GetJob(jobID)
	if err != nil {
		return nil, err
	}

	// This Brigade methos is not a live stream so it's lot faster than
	// `GetJobLogStreamFollow` that it uses `follow: true` in the options
	// of kube client to get the logs.
	rc, err := s.client.GetJobLogStream(job)
	if err != nil {
		return nil, err
	}

	return rc, nil
}

// GetJobLog satisfies Service interface.
func (s *service) GetJobLogStream(jobID string) (io.ReadCloser, error) {
	job, err := s.client.GetJob(jobID)
	if err != nil {
		return nil, err
	}

	rc, err := s.client.GetJobLogStreamFollow(job)
	if err != nil {
		return nil, err
	}

	return rc, nil
}

// RerunBuild satisfies Service interface.
func (s *service) RerunBuild(buildID string) error {
	if buildID == "" {
		return fmt.Errorf("the build ID can't be empty")
	}

	bld, err := s.client.GetBuild(buildID)
	if err != nil {
		return err
	}

	// Replace some existing data so a new build is created
	// based on the original one.
	bld.ID = ""
	bld.Worker = nil

	err = s.client.CreateBuild(bld)
	if err != nil {
		return err
	}

	return nil
}
