package commands

import (
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/brigadecore/brigade/pkg/decolorizer"
	"github.com/brigadecore/brigade/pkg/script"
)

var (
	runFile          string
	runConfigFile    string
	runEvent         string
	runPayloadFile   string
	runInlinePayload string
	runCommitish     string
	runRef           string
	runLogLevel      string
	runNoProgress    bool
	runNoColor       bool
	runBackground    bool
)

const (
	defaultRef = "master"
)

const runUsage = `Send a Brigade JS file to the server.

This sends a file into the cluster and waits for it to complete. It accepts
a project name or project ID.

	$ brig run brigadecore/empty-testbed

When no JS file is supplied, the project will be checked for a brigade.js file
in the associated repository.

To send a local JS file to the server, use the '-f' flag:

	$ brig run -f my.js brigadecore/empty-testbed

While specifying an event is possible, use caution. Many events expect a
particular payload.

A payload can be either specified inline or in a payload file.

To specify a payload file, use the '-p' flag:

	$ brig run -p payload.json

Alternatively to specify it inline, use the '-i' flag:

	$ brig run -i {"key": "value"}

To run the job in the background, use -b/--background. Note, though, that in this
case the exit code indicates only whether the event was submitted, not whether
the worker successfully ran to completion.
`

func init() {
	run.Flags().StringVarP(&runFile, "file", "f", "", "The JavaScript file to execute")
	run.Flags().StringVarP(&runEvent, "event", "e", "exec", "The name of the event to fire")
	run.Flags().StringVarP(&runPayloadFile, "payload", "p", "", "The path to a payload file")
	run.Flags().StringVarP(&runInlinePayload, "inline-payload", "i", "", "The payload specified inline")
	run.Flags().StringVar(&runConfigFile, "config", "", "The brigade.json config file")
	run.Flags().StringVarP(&runCommitish, "commit", "c", "", "A VCS (git) commit")
	run.Flags().StringVarP(&runRef, "ref", "r", defaultRef, "A VCS (git) version, tag, or branch")
	run.Flags().BoolVar(&runNoProgress, "no-progress", false, "Disable progress meter")
	run.Flags().BoolVar(&runNoColor, "no-color", false, "Remove color codes from log output")
	run.Flags().BoolVarP(&runBackground, "background", "b", false, "Trigger the event and exit. Let the job run in the background.")
	run.Flags().StringVarP(&runLogLevel, "level", "l", "log", "Specified log level: log, info, warn, error")
	Root.AddCommand(run)
}

var run = &cobra.Command{
	Use:     "run PROJECT",
	Aliases: []string{"exec"},
	Short:   "Run a brigade.js file",
	Long:    runUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("project name required")
		}
		proj := args[0]

		scr, err := readFileParam(runFile)
		if err != nil {
			return err
		}

		config, err := readFileParam(runConfigFile)
		if err != nil {
			return err
		}

		if runPayloadFile != "" && runInlinePayload != "" {
			return errors.New("Both payload and inline-payload should not be specified")
		}

		var payload []byte
		if runInlinePayload != "" {
			payload = []byte(runInlinePayload)
		} else {
			payload, err = readFileParam(runPayloadFile)
			if err != nil {
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

		runner, err := script.NewDelegatedRunner(kc, globalNamespace)
		if err != nil {
			return err
		}
		runner.ScriptLogDestination = destination
		runner.NoProgress = runNoProgress
		runner.Background = runBackground
		runner.Verbose = globalVerbose

		err = runner.SendScript(proj, scr, config, runEvent, runCommitish, runRef, payload, runLogLevel)
		if err == nil {
			return nil
		}

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

func readFileParam(path string) ([]byte, error) {
	if len(path) > 0 {
		return ioutil.ReadFile(path)
	}
	return []byte{}, nil
}
