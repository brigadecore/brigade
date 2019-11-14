package commands

import (
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/brigadecore/brigade/pkg/decolorizer"
	"github.com/brigadecore/brigade/pkg/script"
)

const rerunUsage = `Request that Brigade re-run a build.

This will clone an old event, assign it a new build ID, and resubmit it. The build
must still be accessible. Once a build is vacuumed, it can no longer be re-run.

Using the '-f' flag will cause brig to resend the old payload, but override the old
script with the provided one.

As this command features all of the same flags as run, additional overrides are available,
such as re-running with a different event, payload body, commit, etc.  See usage via -h/--help.
`

var (
	rerunFile       string
	rerunEvent      string
	rerunPayload    string
	rerunConfigFile string
	rerunCommitish  string
	rerunRef        string
	rerunLogLevel   string
	rerunNoProgress bool
	rerunNoColor    bool
	rerunBackground bool
)

func init() {
	rerun.Flags().StringVarP(&rerunFile, "file", "f", "", "The JavaScript file to execute")
	rerun.Flags().StringVarP(&rerunEvent, "event", "e", "", "The name of the event to fire")
	rerun.Flags().StringVarP(&rerunPayload, "payload", "p", "", "The path to a payload file")
	rerun.Flags().StringVar(&rerunConfigFile, "config", "", "The brigade.json config file")
	rerun.Flags().StringVarP(&rerunCommitish, "commit", "c", "", "A VCS (git) commit")
	rerun.Flags().StringVarP(&rerunRef, "ref", "r", "", "A VCS (git) version, tag, or branch")
	rerun.Flags().BoolVar(&rerunNoProgress, "no-progress", runNoProgress, "Disable progress meter")
	rerun.Flags().BoolVar(&rerunNoColor, "no-color", runNoColor, "Remove color codes from log output")
	rerun.Flags().BoolVarP(&rerunBackground, "background", "b", runBackground, "Trigger the event and exit. Let the job rerun in the background.")
	rerun.Flags().StringVarP(&rerunLogLevel, "level", "l", "log", "Specified log level: log, info, warn, error")
	Root.AddCommand(rerun)
}

var rerun = &cobra.Command{
	Use:   "rerun BUILD_ID",
	Short: "Re-run the build associated with the provided ID.",
	Long:  rerunUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("build ID required")
		}
		bid := args[0]

		kc, err := kubeClient()
		if err != nil {
			return err
		}

		a, err := script.NewDelegatedRunner(kc, globalNamespace)
		if err != nil {
			return err
		}

		updateRunner(a)

		build, err := getUpdatedBuild(a, bid)
		if err != nil {
			return err
		}

		err = a.SendBuild(build)

		// If err is a BuildFailure, then we don't want Cobra to print the Usage
		// instructions on failure, since it's a pipeline issue and not a CLI issue.
		_, ok := err.(script.BuildFailure)
		if ok {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			return BrigError{Code: 2, cause: err}
		}

		return err
	},
}

func updateRunner(r *script.Runner) {
	var destination io.Writer = os.Stdout
	if rerunNoColor {
		// Pipe the data through a Writer that strips the color codes and then
		// sends the resulting data to the underlying writer.
		destination = decolorizer.New(destination)
	}

	r.ScriptLogDestination = destination
	r.NoProgress = rerunNoProgress
	r.Background = rerunBackground
	r.Verbose = globalVerbose
}

func getUpdatedBuild(r *script.Runner, bid string) (*brigade.Build, error) {
	build, err := r.GetBuild(bid)
	if err != nil {
		return build, err
	}

	if rerunFile != "" {
		data, err := ioutil.ReadFile(rerunFile)
		if err != nil {
			return build, err
		}
		build.Script = data
	}

	if rerunConfigFile != "" {
		config, err := ioutil.ReadFile(rerunConfigFile)
		if err != nil {
			return build, err
		}
		build.Config = config
	}

	if len(rerunPayload) > 0 {
		var err error
		if build.Payload, err = ioutil.ReadFile(rerunPayload); err != nil {
			return build, err
		}
	}

	if rerunEvent != "" {
		build.Type = rerunEvent
	}
	if rerunCommitish != "" {
		build.Revision.Commit = rerunCommitish
	}
	if rerunRef != "" {
		build.Revision.Ref = rerunRef
	}

	// Override a few things
	build.ID = ""
	build.LogLevel = rerunLogLevel
	build.Worker = nil

	return build, nil
}
