package commands

import (
	"errors"

	"github.com/spf13/cobra"
)

const rerunUsage = `Request that Brigade re-run a build.

This will clone an old event, assign it a new build ID, and resubmit it. The build
must still be accessible. Once a build is vacuumed, it can no longer be re-run.

`

var rerunLogLevel string

func init() {
	rerun.Flags().StringVarP(&rerunLogLevel, "level", "l", "log", "Specified log level: log, info, warn, error")
	Root.AddCommand(rerun)
}

var rerun = &cobra.Command{
	Use:   "rerun BUILD_ID",
	Short: "Given an existing build ID, re-run the same event.",
	Long:  rerunUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("build ID required")
		}
		bid := args[0]

		a, err := newScriptRunner()
		if err != nil {
			return err
		}

		build, err := a.getBuild(bid)
		if err != nil {
			return err
		}

		// Override a few things
		build.ID = ""
		build.LogLevel = rerunLogLevel
		build.Worker = nil

		return a.sendBuild(build)
	},
}
