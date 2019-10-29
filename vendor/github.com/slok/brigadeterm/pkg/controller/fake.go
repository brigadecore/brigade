package controller

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"time"
)

var (
	staticNow = time.Now()
)

type fake struct {
	// faked builds
	builds []*Build

	// Track when job log asked for the first time.
	jobLogAskedFirstTime time.Time
	jobFinished          bool
}

// NewFakeController returns a new faked controller.
func NewFakeController() Controller {
	return &fake{
		builds: []*Build{
			&Build{
				ID:        "lkjdsbfdflkdsnflkjdsbflkjadbflkjaful",
				Version:   "3140d400028b44f9f21f597b0c4d61f537fc51fc",
				State:     SuccessedState,
				EventType: "github:push",
				Started:   staticNow.Add(-9999 * time.Hour),
				Ended:     time.Now().Add(-9998 * time.Hour),
			},
			&Build{
				ID:        "flkjdsbfuldflkdsnflkjdsbflkjadbflkja",
				Version:   "1f537fc5c1f028b44f9f274d3140d40061f59b0c",
				State:     FailedState,
				EventType: "deploy",
				Started:   staticNow.Add(-1 * time.Hour),
				Ended:     time.Now().Add(-50 * time.Minute),
			},
			&Build{
				ID:        "24351321ldflkds32kjdsbflkj323dbflkja",
				Version:   "b44f9f21f59b7161f537fc5cf0280c4d3140d400",
				State:     PendingState,
				EventType: "github:push",
				Started:   staticNow.Add(-120 * time.Hour).Add(-19 * time.Minute),
				Ended:     time.Now().Add(-120 * time.Hour).Add(-18 * time.Minute),
			},
			&Build{
				ID:        "2oijohpobna123213eewfeflkj323dbflkja",
				Version:   "061f537fc5c71f0d3140d4028b44f9f21f59b0c4",
				State:     UnknownState,
				EventType: "github:push",
				Started:   staticNow.Add(-5 * time.Minute),
				Ended:     time.Now().Add(-128 * time.Second),
			},
			&Build{
				ID:        "oijohpobna123213eewfeflkj323dbfl2kja",
				Version:   "244f9f0d40537fc5c1f59061fb0c4d31471f028b",
				State:     RunningState,
				EventType: "github:pull_reqest",
				Started:   staticNow.Add(-30 * time.Second),
			},
			&Build{},
			nil,
		},
	}
}

func (f *fake) ProjectListPageContext() *ProjectListPageContext {
	return &ProjectListPageContext{
		Projects: []*Project{
			&Project{
				ID:   "1",
				Name: "company1/AAAAAA",
				LastBuilds: []*Build{
					&Build{
						ID:        "lkjdsbfdflkdsnflkjdsbflkjadbflkjaful",
						Version:   "3140d400028b44f9f21f597b0c4d61f537fc51fc",
						State:     SuccessedState,
						EventType: "github:push",
						Started:   staticNow.Add(-9999 * time.Hour),
						Ended:     time.Now().Add(-9998 * time.Hour),
					},
					&Build{State: FailedState},
					&Build{State: SuccessedState},
					&Build{State: SuccessedState},
					&Build{State: FailedState},
				},
			},
			&Project{
				ID:   "2",
				Name: "company1/123456",
				LastBuilds: []*Build{
					&Build{
						ID:        "flkjdsbfuldflkdsnflkjdsbflkjadbflkja",
						Version:   "1f537fc5c1f028b44f9f274d3140d40061f59b0c",
						State:     SuccessedState,
						EventType: "deploy",
						Started:   staticNow.Add(-1 * time.Hour),
						Ended:     time.Now().Add(-50 * time.Minute),
					},
					&Build{State: SuccessedState},
					&Build{State: SuccessedState},
				},
			},
			&Project{
				ID:   "3",
				Name: "company2/987652",
				LastBuilds: []*Build{
					&Build{
						ID:        "24351321ldflkds32kjdsbflkj323dbflkja",
						Version:   "b44f9f21f59b7161f537fc5cf0280c4d3140d400",
						State:     FailedState,
						EventType: "github:push",
						Started:   staticNow.Add(-120 * time.Hour).Add(-19 * time.Minute),
						Ended:     time.Now().Add(-120 * time.Hour).Add(-18 * time.Minute),
					},
					&Build{State: FailedState},
					&Build{State: UnknownState},
					&Build{State: SuccessedState},
					&Build{State: UnknownState},
				},
			},
			&Project{
				ID:   "4",
				Name: "company3/123sads",
				LastBuilds: []*Build{
					&Build{
						ID:        "2oijohpobna123213eewfeflkj323dbflkja",
						Version:   "061f537fc5c71f0d3140d4028b44f9f21f59b0c4",
						State:     PendingState,
						EventType: "github:push",
						Started:   staticNow.Add(-5 * time.Minute),
						Ended:     time.Now().Add(-128 * time.Second),
					},
					&Build{State: SuccessedState},
					&Build{State: SuccessedState},
					&Build{State: FailedState},
					&Build{State: SuccessedState},
				},
			},
			&Project{
				ID:   "5",
				Name: "company4/3fg1",
				LastBuilds: []*Build{
					&Build{
						ID:        "oijohpobna123213eewfeflkj323dbfl2kja",
						Version:   "244f9f0d40537fc5c1f59061fb0c4d31471f028b",
						State:     RunningState,
						EventType: "github:pull_reqest",
						Started:   time.Now().Add(-30 * time.Second),
					},
					&Build{State: FailedState},
					&Build{State: FailedState},
					&Build{State: FailedState},
					&Build{State: FailedState},
				},
			},
			&Project{
				ID:   "6",
				Name: "company4/1234567566345",
				LastBuilds: []*Build{
					&Build{
						ID:        "flkjdsbfuldflkdsnflkjdsbflkjadbflkja",
						Version:   "1f537fc5c1f028b44f9f274d3140d40061f59b0c",
						State:     SuccessedState,
						EventType: "deploy",
						Started:   staticNow.Add(-1 * time.Hour),
						Ended:     time.Now().Add(-50 * time.Minute),
					},
					&Build{State: FailedState},
					&Build{State: FailedState},
					&Build{State: FailedState},
					&Build{State: UnknownState},
				},
			},
			&Project{
				ID:   "7",
				Name: "company5/423df",
				LastBuilds: []*Build{
					&Build{
						ID:        "24351321ldflkds32kjdsbflkj323dbflkja",
						Version:   "b44f9f21f59b7161f537fc5cf0280c4d3140d400",
						State:     UnknownState,
						EventType: "github:push",
						Started:   staticNow.Add(-120 * time.Hour).Add(-19 * time.Minute),
						Ended:     time.Now().Add(-120 * time.Hour).Add(-18 * time.Minute),
					},
					&Build{State: SuccessedState},
					&Build{State: SuccessedState},
					&Build{State: SuccessedState},
					&Build{State: SuccessedState},
				},
			},
			&Project{
				ID:   "8",
				Name: "company1/ggasdasft",
				LastBuilds: []*Build{
					&Build{
						ID:        "2oijohpobna123213eewfeflkj323dbflkja",
						Version:   "061f537fc5c71f0d3140d4028b44f9f21f59b0c4",
						State:     SuccessedState,
						EventType: "github:push",
						Started:   staticNow.Add(-5 * time.Minute),
						Ended:     time.Now().Add(-128 * time.Second),
					},
					&Build{State: FailedState},
					&Build{State: SuccessedState},
					&Build{State: SuccessedState},
					&Build{State: SuccessedState},
				},
			},
			&Project{
				ID:   "8",
				Name: "company1/0184848q1danfubu<s",
			},
			nil,
		},
	}
}

func (f *fake) ProjectBuildListPageContext(projectID string) *ProjectBuildListPageContext {
	return &ProjectBuildListPageContext{
		ProjectName: "company1/AAAAAA",
		ProjectNS:   "ci",
		ProjectURL:  "git@github.com:slok/brigadeterm",
		Builds:      f.builds,
	}
}

func (f *fake) BuildJobListPageContext(buildID string) *BuildJobListPageContext {
	return &BuildJobListPageContext{
		BuildInfo: &Build{
			ID:        "2oijohpobna123213eewfeflkj323dbflkja",
			Version:   "061f537fc5c71f0d3140d4028b44f9f21f59b0c4",
			State:     SuccessedState,
			EventType: "github:push",
			Started:   time.Now().Add(-5 * time.Minute),
			Ended:     time.Now().Add(-128 * time.Second),
		},
		Jobs: []*Job{
			&Job{
				ID:      "unit-test-01c8zehre13ht12776hdkms8gf",
				Name:    "unit-test",
				Image:   "golang:1.9",
				State:   FailedState,
				Started: staticNow.Add(-11 * time.Minute),
				Ended:   time.Now().Add(-9 * time.Minute),
			},
			&Job{
				ID:      "build-binary-1-01c8zehre13ht12776hdkms8gf",
				Name:    "build-binary-1",
				Image:   "docker:stable-dind",
				State:   RunningState,
				Started: staticNow.Add(-9 * time.Minute),
				Ended:   time.Now().Add(-5 * time.Minute),
			},
			&Job{
				ID:      "build-binary-1-01c8zehre13ht12776hdkms8gf",
				Name:    "build-binary-2",
				Image:   "docker:stable-dind",
				State:   PendingState,
				Started: staticNow.Add(-9 * time.Minute),
				Ended:   time.Now().Add(-5 * time.Minute),
			},
			&Job{
				ID:      "build-binary-3-01c8zehre13ht12776hdkms8gf",
				Name:    "build-binary-3",
				Image:   "docker:stable-dind",
				State:   UnknownState,
				Started: staticNow.Add(-9 * time.Minute),
				Ended:   time.Now().Add(-3 * time.Minute),
			},
			&Job{
				ID:      "set-github-build-status-01c8zehre13ht12776hdkms8gf",
				Name:    "set-github-build-status",
				Image:   "technosophos/github-notify:latest",
				State:   RunningState,
				Started: staticNow.Add(-3 * time.Minute),
				Ended:   time.Now().Add(-1 * time.Minute),
			},
			nil,
		},
	}
}

func (f *fake) JobLogPageContext(jobID string) *JobLogPageContext {
	var logrc io.ReadCloser
	state := UnknownState

	// If not finished fake a live log, when the job is marked as finished.
	// return a regular finished log.
	if !f.jobFinished {
		state = RunningState                // Fake the state.
		f.jobLogAskedFirstTime = time.Now() // Track ask for first time.

		// Create a pipe for our fake log.
		r, w := io.Pipe()

		// Stream the faked log.
		go func() {
			defer w.Close()
			linesCnt := 0
			for {
				color := time.Now().Nanosecond() % 7
				_, err := fmt.Fprintf(w, "---> \x1b[01;3%dmlogline %d --------- %d\n", color, linesCnt, time.Now().Nanosecond())
				if err != nil {
					return // Something happenned, we don't mind if it's ended or not, stop.
				}
				sleepMS := time.Duration(time.Now().Nanosecond() % 500)
				time.Sleep(sleepMS * time.Millisecond)
				linesCnt++
			}
		}()
		logrc = r
	} else {
		state = SuccessedState // Fake the state.

		// Create our fake log.
		log := ""
		for i := 0; i < 1000; i++ {
			color := time.Now().Nanosecond() % 7
			log = fmt.Sprintf("%s---> \x1b[01;3%dmlogline %d --------- %d\n", log, color, i, time.Now().Nanosecond())
		}

		b := bytes.NewBufferString(log)
		logrc = ioutil.NopCloser(b)
	}

	return &JobLogPageContext{
		Job: &Job{
			ID:      "build-binary-3-01c8zehre13ht12776hdkms8gf",
			Name:    "build-binary-3",
			Image:   "docker:stable-dind",
			State:   state,
			Started: staticNow.Add(-9 * time.Minute),
			Ended:   time.Now().Add(-3 * time.Minute),
		},
		Log: logrc,
	}
}

func (f *fake) JobRunning(jobID string) bool {
	// If we have been getting the logs for more than 20s then mark as ended.
	if time.Now().Sub(f.jobLogAskedFirstTime) > (20 * time.Second) {
		f.jobFinished = true
		return false
	}

	return true
}

func (f *fake) RerunBuild(buildID string) error {
	f.builds = append(f.builds, &Build{
		ID:        fmt.Sprintf("rerun-%s", buildID),
		Version:   "11111111111111111111111111111",
		State:     SuccessedState,
		EventType: "someEvent",
		Started:   staticNow.Add(-9999 * time.Hour),
		Ended:     time.Now().Add(-9998 * time.Hour),
	})
	return nil
}
