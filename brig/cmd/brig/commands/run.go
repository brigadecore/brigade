package commands

import (
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/Azure/brigade/pkg/decolorizer"
	"github.com/Azure/brigade/pkg/script"
)

var (
	runFile       string
	runEvent      string
	runPayload    string
	runCommitish  string
	runRef        string
	runLogLevel   string
	runNoProgress bool
	runNoColor    bool
	runBackground bool
)

const (
	defaultRef = "master"
)

const runUsage = `Send a Brigade JS file to the server.

This sends a file into the cluster and waits for it to complete. It accepts
a project name or project ID.

	$ brig run deis/empty-testbed

When no JS file is supplied, the project will be checked for a brigade.js file
in the associated repository.

To send a local JS file to the server, use the '-f' flag:

	$ brig run -f my.js deis/empty-testbed

While specifying an event is possible, use caution. Many events expect a
particular payload.

To run the job in the background, use -b/--background. Note, though, that in this
case the exit code indicates only whether the event was submitted, not whether
the worker successfully ran to completion.
`

func init() {
	run.Flags().StringVarP(&runFile, "file", "f", "", "The JavaScript file to execute")
	run.Flags().StringVarP(&runEvent, "event", "e", "exec", "The name of the event to fire")
	run.Flags().StringVarP(&runPayload, "payload", "p", "", "The path to a payload file")
	run.Flags().StringVarP(&runCommitish, "commit", "c", "", "A VCS (git) commit")
	run.Flags().StringVarP(&runRef, "ref", "r", defaultRef, "A VCS (git) version, tag, or branch")
	run.Flags().BoolVar(&runNoProgress, "no-progress", false, "Disable progress meter")
	run.Flags().BoolVar(&runNoColor, "no-color", false, "Remove color codes from log output")
	run.Flags().BoolVarP(&runBackground, "background", "b", false, "Trigger the event and exit. Let the job run in the background.")
	run.Flags().StringVarP(&runLogLevel, "level", "l", "log", "Specified log level: log, info, warn, error")
	Root.AddCommand(run)
}

var run = &cobra.Command{
	Use:   "run PROJECT",
	Short: "Run a brigade.js file",
	Long:  runUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("project name required")
		}
		proj := args[0]

		var scr []byte
		if len(runFile) > 0 {
			var err error
			if scr, err = ioutil.ReadFile(runFile); err != nil {
				return err
			}
		}

		var payload []byte
		if len(runPayload) > 0 {
			var err error
			if payload, err = ioutil.ReadFile(runPayload); err != nil {
				return err
			}
		}

		var destination io.Writer = os.Stdout
		if runNoColor {
			// Pipe the data through a Writer that strips the color codes and then
			// sends the resulting data to the underlying writer.
			destination = decolorizer.New(destination)
		}

		kc, err := kubeClient()
		if err != nil {
			return err
		}

		runner, err := script.NewDelegatedRunner(kc, destination, globalNamespace)
		if err != nil {
			return err
		}
		runner.NoProgress = runNoProgress
		runner.Background = runBackground
		runner.Verbose = globalVerbose

		return runner.SendScript(proj, scr, runEvent, runCommitish, runRef, payload, runLogLevel)
	},
}
