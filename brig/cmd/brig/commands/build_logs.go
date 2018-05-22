package commands

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
	"github.com/Azure/brigade/pkg/storage/kube"
)

const (
	logHeader      = "\n==========[  %s  ]==========\n"
	buildLogsUsage = `Show log(s) for a build

Print the logs for a build, and (optionally) the jobs executed as part of the
build.
`
)

var (
	logsJobs bool
	logsLast bool
)

func init() {
	buildLogs.Flags().BoolVarP(&logsJobs, "jobs", "j", false, "Show job logs as well as the worker log")
	buildLogs.Flags().BoolVarP(&logsLast, "last", "l", false, "Show last build's log (ignores BUILD_ID)")
	build.AddCommand(buildLogs)
}

var buildLogs = &cobra.Command{
	Use:   "logs BUILD_ID",
	Short: "show build logs",
	Long:  buildLogsUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		if logsLast {
			return showBuildLogs(cmd.OutOrStdout(), "")
		}
		if len(args) == 0 {
			return errors.New("either BUILD_ID or --last is required")
		}
		return showBuildLogs(cmd.OutOrStdout(), args[0])
	},
}

func showBuildLogs(out io.Writer, buildID string) error {
	c, err := kubeClient()
	if err != nil {
		return err
	}

	store := kube.New(c, globalNamespace)
	if buildID == "" {
		buildID, err = lastBuildID(store)
		if err != nil {
			return err
		}
	}

	bs, err := store.GetBuild(buildID)
	if err != nil {
		return err
	}

	if bs.Worker == nil {
		return fmt.Errorf("No logs for build %q", buildID)
	}

	workerLog, err := store.GetWorkerLog(bs.Worker)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, logHeader, bs.Worker.ID)
	fmt.Fprint(out, workerLog)
	if logsJobs {
		return showJobLogs(out, bs, store)
	}
	return nil
}

func lastBuildID(store storage.Store) (string, error) {
	builds, err := store.GetBuilds()
	if err != nil {
		return "", err
	}

	if len(builds) == 0 {
		return "", errors.New("no builds")
	}

	return builds[len(builds)-1].ID, nil
}

func showJobLogs(out io.Writer, build *brigade.Build, store storage.Store) error {
	// This indicates that there were no jobs because the worker never
	// spawned.
	if build.Worker == nil {
		return fmt.Errorf("Build %s has no worker", build.ID)
	}

	jobs, err := store.GetBuildJobs(build)
	if err != nil {
		return err
	}

	for _, j := range jobs {
		fmt.Fprintf(out, logHeader, j.ID)
		log, err := store.GetJobLog(j)
		if err != nil {
			fmt.Fprintf(out, "log not found: %s", err)
			continue
		}
		fmt.Fprint(out, log)
	}
	return nil
}
