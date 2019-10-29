package brigade_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	azurebrigade "github.com/brigadecore/brigade/pkg/brigade"
	mstore "github.com/slok/brigadeterm/pkg/mocks/github.com/brigadecore/brigade/pkg/storage"
	brigademodel "github.com/slok/brigadeterm/pkg/model/brigade"
	"github.com/slok/brigadeterm/pkg/service/brigade"
)

func TestGetProjectBuilds(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		desc      bool
		builds    []*azurebrigade.Build
		expBuilds []*brigademodel.Build
	}{
		{
			name: "A list of builds in ascending order should be ordered correctly by time and maintain the equals in time and the nil workers at the end of the list",
			desc: false,
			builds: []*azurebrigade.Build{
				&azurebrigade.Build{ID: "00"},
				&azurebrigade.Build{ID: "01", Worker: &azurebrigade.Worker{StartTime: now.Add(-10 * time.Hour)}},
				&azurebrigade.Build{ID: "02", Worker: &azurebrigade.Worker{StartTime: now.Add(-11 * time.Hour)}},
				&azurebrigade.Build{ID: "03", Worker: &azurebrigade.Worker{StartTime: now.Add(-20 * time.Hour)}},
				&azurebrigade.Build{ID: "04", Worker: &azurebrigade.Worker{StartTime: now.Add(-1 * time.Hour)}},
				&azurebrigade.Build{ID: "05"},
				&azurebrigade.Build{ID: "06", Worker: &azurebrigade.Worker{StartTime: now.Add(-10 * time.Hour)}},
				&azurebrigade.Build{ID: "07", Worker: &azurebrigade.Worker{StartTime: now.Add(-3 * time.Hour)}},
				&azurebrigade.Build{ID: "08", Worker: &azurebrigade.Worker{StartTime: now.Add(-4 * time.Hour)}},
				&azurebrigade.Build{ID: "09", Worker: &azurebrigade.Worker{StartTime: now.Add(-3 * time.Hour)}},
				&azurebrigade.Build{ID: "10", Worker: &azurebrigade.Worker{StartTime: now.Add(-5 * time.Hour)}},
			},
			expBuilds: []*brigademodel.Build{
				&brigademodel.Build{ID: "03", Worker: &azurebrigade.Worker{StartTime: now.Add(-20 * time.Hour)}},
				&brigademodel.Build{ID: "02", Worker: &azurebrigade.Worker{StartTime: now.Add(-11 * time.Hour)}},
				&brigademodel.Build{ID: "01", Worker: &azurebrigade.Worker{StartTime: now.Add(-10 * time.Hour)}},
				&brigademodel.Build{ID: "06", Worker: &azurebrigade.Worker{StartTime: now.Add(-10 * time.Hour)}},
				&brigademodel.Build{ID: "10", Worker: &azurebrigade.Worker{StartTime: now.Add(-5 * time.Hour)}},
				&brigademodel.Build{ID: "08", Worker: &azurebrigade.Worker{StartTime: now.Add(-4 * time.Hour)}},
				&brigademodel.Build{ID: "07", Worker: &azurebrigade.Worker{StartTime: now.Add(-3 * time.Hour)}},
				&brigademodel.Build{ID: "09", Worker: &azurebrigade.Worker{StartTime: now.Add(-3 * time.Hour)}},
				&brigademodel.Build{ID: "04", Worker: &azurebrigade.Worker{StartTime: now.Add(-1 * time.Hour)}},
				&azurebrigade.Build{ID: "00"},
				&azurebrigade.Build{ID: "05"},
			},
		},
		{
			name: "A list of builds in descenging order should be ordered correctly by time and maintain the equals in time",
			desc: true,
			builds: []*azurebrigade.Build{
				&azurebrigade.Build{ID: "00"},
				&azurebrigade.Build{ID: "01", Worker: &azurebrigade.Worker{StartTime: now.Add(-10 * time.Hour)}},
				&azurebrigade.Build{ID: "02", Worker: &azurebrigade.Worker{StartTime: now.Add(-11 * time.Hour)}},
				&azurebrigade.Build{ID: "03", Worker: &azurebrigade.Worker{StartTime: now.Add(-20 * time.Hour)}},
				&azurebrigade.Build{ID: "04", Worker: &azurebrigade.Worker{StartTime: now.Add(-1 * time.Hour)}},
				&azurebrigade.Build{ID: "05"},
				&azurebrigade.Build{ID: "06", Worker: &azurebrigade.Worker{StartTime: now.Add(-10 * time.Hour)}},
				&azurebrigade.Build{ID: "07", Worker: &azurebrigade.Worker{StartTime: now.Add(-3 * time.Hour)}},
				&azurebrigade.Build{ID: "08", Worker: &azurebrigade.Worker{StartTime: now.Add(-4 * time.Hour)}},
				&azurebrigade.Build{ID: "09", Worker: &azurebrigade.Worker{StartTime: now.Add(-3 * time.Hour)}},
				&azurebrigade.Build{ID: "10", Worker: &azurebrigade.Worker{StartTime: now.Add(-5 * time.Hour)}},
			},
			expBuilds: []*brigademodel.Build{
				&brigademodel.Build{ID: "04", Worker: &azurebrigade.Worker{StartTime: now.Add(-1 * time.Hour)}},
				&brigademodel.Build{ID: "07", Worker: &azurebrigade.Worker{StartTime: now.Add(-3 * time.Hour)}},
				&brigademodel.Build{ID: "09", Worker: &azurebrigade.Worker{StartTime: now.Add(-3 * time.Hour)}},
				&brigademodel.Build{ID: "08", Worker: &azurebrigade.Worker{StartTime: now.Add(-4 * time.Hour)}},
				&brigademodel.Build{ID: "10", Worker: &azurebrigade.Worker{StartTime: now.Add(-5 * time.Hour)}},
				&brigademodel.Build{ID: "01", Worker: &azurebrigade.Worker{StartTime: now.Add(-10 * time.Hour)}},
				&brigademodel.Build{ID: "06", Worker: &azurebrigade.Worker{StartTime: now.Add(-10 * time.Hour)}},
				&brigademodel.Build{ID: "02", Worker: &azurebrigade.Worker{StartTime: now.Add(-11 * time.Hour)}},
				&brigademodel.Build{ID: "03", Worker: &azurebrigade.Worker{StartTime: now.Add(-20 * time.Hour)}},
				&azurebrigade.Build{ID: "00"},
				&azurebrigade.Build{ID: "05"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mocks.
			ms := &mstore.Store{}
			ms.On("GetProject", mock.Anything).Return(nil, nil)
			ms.On("GetProjectBuilds", mock.Anything).Return(test.builds, nil)

			// Create service and run.
			bsvc := brigade.NewService(ms)
			gotBs, err := bsvc.GetProjectBuilds(&brigademodel.Project{ID: "test"}, test.desc)

			if assert.NoError(err) {
				assert.Equal(test.expBuilds, gotBs)
			}
		})
	}
}

func TestGetBuildJobs(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		desc    bool
		jobs    []*azurebrigade.Job
		expJobs []*brigademodel.Job
	}{
		{
			name: "A list of jobs in ascending order should be ordered correctly by time and maintain the equals in time",
			desc: false,
			jobs: []*azurebrigade.Job{
				&azurebrigade.Job{ID: "01", StartTime: now.Add(-10 * time.Hour)},
				&azurebrigade.Job{ID: "02", StartTime: now.Add(-11 * time.Hour)},
				&azurebrigade.Job{ID: "03", StartTime: now.Add(-20 * time.Hour)},
				&azurebrigade.Job{ID: "04", StartTime: now.Add(-1 * time.Hour)},
				&azurebrigade.Job{ID: "05", StartTime: now.Add(-10 * time.Hour)},
				&azurebrigade.Job{ID: "06", StartTime: now.Add(-3 * time.Hour)},
				&azurebrigade.Job{ID: "07", StartTime: now.Add(-4 * time.Hour)},
				&azurebrigade.Job{ID: "08", StartTime: now.Add(-3 * time.Hour)},
				&azurebrigade.Job{ID: "09", StartTime: now.Add(-5 * time.Hour)},
			},
			expJobs: []*brigademodel.Job{
				&azurebrigade.Job{ID: "03", StartTime: now.Add(-20 * time.Hour)},
				&azurebrigade.Job{ID: "02", StartTime: now.Add(-11 * time.Hour)},
				&azurebrigade.Job{ID: "01", StartTime: now.Add(-10 * time.Hour)},
				&azurebrigade.Job{ID: "05", StartTime: now.Add(-10 * time.Hour)},
				&azurebrigade.Job{ID: "09", StartTime: now.Add(-5 * time.Hour)},
				&azurebrigade.Job{ID: "07", StartTime: now.Add(-4 * time.Hour)},
				&azurebrigade.Job{ID: "06", StartTime: now.Add(-3 * time.Hour)},
				&azurebrigade.Job{ID: "08", StartTime: now.Add(-3 * time.Hour)},
				&azurebrigade.Job{ID: "04", StartTime: now.Add(-1 * time.Hour)},
			},
		},
		{
			name: "A list of jobs in descenging order should be ordered correctly by time and maintain the equals in time",
			desc: true,
			jobs: []*azurebrigade.Job{
				&azurebrigade.Job{ID: "01", StartTime: now.Add(-10 * time.Hour)},
				&azurebrigade.Job{ID: "02", StartTime: now.Add(-11 * time.Hour)},
				&azurebrigade.Job{ID: "03", StartTime: now.Add(-20 * time.Hour)},
				&azurebrigade.Job{ID: "04", StartTime: now.Add(-1 * time.Hour)},
				&azurebrigade.Job{ID: "05", StartTime: now.Add(-10 * time.Hour)},
				&azurebrigade.Job{ID: "06", StartTime: now.Add(-3 * time.Hour)},
				&azurebrigade.Job{ID: "07", StartTime: now.Add(-4 * time.Hour)},
				&azurebrigade.Job{ID: "08", StartTime: now.Add(-3 * time.Hour)},
				&azurebrigade.Job{ID: "09", StartTime: now.Add(-5 * time.Hour)},
			},
			expJobs: []*brigademodel.Job{
				&azurebrigade.Job{ID: "04", StartTime: now.Add(-1 * time.Hour)},
				&azurebrigade.Job{ID: "06", StartTime: now.Add(-3 * time.Hour)},
				&azurebrigade.Job{ID: "08", StartTime: now.Add(-3 * time.Hour)},
				&azurebrigade.Job{ID: "07", StartTime: now.Add(-4 * time.Hour)},
				&azurebrigade.Job{ID: "09", StartTime: now.Add(-5 * time.Hour)},
				&azurebrigade.Job{ID: "01", StartTime: now.Add(-10 * time.Hour)},
				&azurebrigade.Job{ID: "05", StartTime: now.Add(-10 * time.Hour)},
				&azurebrigade.Job{ID: "02", StartTime: now.Add(-11 * time.Hour)},
				&azurebrigade.Job{ID: "03", StartTime: now.Add(-20 * time.Hour)},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mocks.
			ms := &mstore.Store{}
			ms.On("GetBuild", mock.Anything).Return(nil, nil)
			ms.On("GetBuildJobs", mock.Anything).Return(test.jobs, nil)

			// Create service and run.
			bsvc := brigade.NewService(ms)
			gotJs, err := bsvc.GetBuildJobs("test", test.desc)

			if assert.NoError(err) {
				assert.Equal(test.expJobs, gotJs)
			}
		})
	}
}

func TestRerunBuild(t *testing.T) {

	tests := []struct {
		name     string
		build    *azurebrigade.Build
		expBuild *azurebrigade.Build
	}{
		{
			name: "Rerunning an existing build should rerun the build",
			build: &azurebrigade.Build{
				ID:      "01",
				Script:  []byte("myScipt"),
				Payload: []byte("myPayload"),

				Worker: &azurebrigade.Worker{
					StartTime: time.Now(),
				},
			},

			expBuild: &azurebrigade.Build{
				Script:  []byte("myScipt"),
				Payload: []byte("myPayload"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mocks.
			ms := &mstore.Store{}
			ms.On("GetBuild", mock.Anything).Once().Return(test.build, nil)
			ms.On("CreateBuild", test.expBuild).Once().Return(nil)

			// Create service and run.
			bsvc := brigade.NewService(ms)
			err := bsvc.RerunBuild("test")

			if assert.NoError(err) {
				ms.AssertExpectations(t)
			}
		})
	}
}
