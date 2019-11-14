package commands

import (
	"testing"

	"github.com/brigadecore/brigade/pkg/decolorizer"

	"k8s.io/client-go/kubernetes/fake"

	"github.com/brigadecore/brigade/pkg/script"
)

func TestRerun_updateRunner(t *testing.T) {
	type updateRunnerTest struct {
		Name       string
		NoColor    bool
		NoProgress bool
		Background bool
		Verbose    bool
	}

	testcases := []updateRunnerTest{
		{
			Name:       "Defaults",
			NoColor:    false,
			NoProgress: false,
			Background: false,
			Verbose:    false,
		},
		{
			Name:       "Overrides",
			NoColor:    true,
			NoProgress: true,
			Background: true,
			Verbose:    true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			r := &script.Runner{}

			if tc.Name == "Overrides" {
				rerunBackground = tc.Background
				rerunNoColor = tc.NoColor
				rerunNoProgress = tc.NoProgress
				globalVerbose = tc.Verbose
			}

			updateRunner(r)

			if r.NoProgress != tc.NoProgress {
				t.Errorf("expected NoProgress to be %t, was %t", tc.NoProgress, r.NoProgress)
			}

			if r.Background != tc.Background {
				t.Errorf("expected Background to be %t, was %t", tc.Background, r.Background)
			}

			if r.Verbose != tc.Verbose {
				t.Errorf("expected Verbose to be %t, was %t", tc.Verbose, r.Verbose)
			}

			if _, ok := r.ScriptLogDestination.(*decolorizer.Writer); ok != tc.NoColor {
				if tc.NoColor {
					t.Error("expected ScriptLogDestination to be of type *decolorizer.Writer")
				} else {
					t.Error("expected ScriptLogDestination to be of type io.Writer")
				}
			}
		})
	}
}

func TestRerun_getUpdatedBuild(t *testing.T) {
	type updatedBuildTest struct {
		Name     string
		LogLevel string
		Type     string
		Config   string
		Commit   string
		Ref      string
	}

	testcases := []updatedBuildTest{
		{
			Name:     "Defaults",
			LogLevel: defaultStubBuildData.LogLevel,
			Type:     defaultStubBuildData.Event,
			Commit:   defaultStubBuildData.Commit,
			Config:   defaultStubBuildData.Config,
			Ref:      defaultStubBuildData.Ref,
		},
		{
			Name:     "Overrides",
			LogLevel: "warn",
			Type:     "different-event",
			Commit:   "new-commit",
			Config:   "new-config",
			Ref:      "new-ref",
		},
	}

	client := fake.NewSimpleClientset()

	// Thank you, build_list_test.go!
	createFakeBuilds(t, client)

	r, err := script.NewDelegatedRunner(client, "default")
	if err != nil {
		t.Fatalf("expected err to be nil, was: %s", err.Error())
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Name == "Overrides" {
				rerunLogLevel = tc.LogLevel
				rerunCommitish = tc.Commit
				rerunRef = tc.Ref
				rerunEvent = tc.Type
			}

			build, err := getUpdatedBuild(r, stubBuild1ID)
			if err != nil {
				t.Errorf("expected err to be nil, was: %s", err.Error())
			}

			if build.ID != "" {
				t.Errorf("expected build.ID to be empty, was: %s", build.ID)
			}

			if build.Worker != nil {
				t.Errorf("expected build.Worker to be %v, was: %v", nil, build.Worker)
			}

			if build.LogLevel != tc.LogLevel {
				t.Errorf("expected build.LogLevel to be %s, was: %s", tc.LogLevel, build.LogLevel)
			}

			if build.Revision.Commit != tc.Commit {
				t.Errorf("expected build.Revision.Commit to be %s, was: %s", tc.Commit, build.Revision.Commit)
			}

			if build.Revision.Ref != tc.Ref {
				t.Errorf("expected build.Revision.Ref to be %s, was: %s", tc.Ref, build.Revision.Ref)
			}

			if build.Type != tc.Type {
				t.Errorf("expected build.Type to be %s, was: %s", tc.Type, build.Type)
			}
		})
	}
}
