package commands

import (
	"errors"
	"io"

	"github.com/brigadecore/brigade/pkg/storage"
	"github.com/brigadecore/brigade/pkg/storage/kube"

	"github.com/spf13/cobra"
)

const buildDeleteUsage = `Deletes a build and its corresponding jobs.`

var forceDeleteRunning bool

func init() {
	build.AddCommand(buildDelete)
	buildDelete.Flags().BoolVar(&forceDeleteRunning, "force", false, "If set, will also delete running builds. Default: false")
}

var buildDelete = &cobra.Command{
	Use:   "delete BUILD_ID",
	Short: "deletes build",
	Long:  buildDeleteUsage,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("build ID is a required argument")
		}
		return deleteBuild(cmd.OutOrStdout(), args[0])
	},
}

func deleteBuild(out io.Writer, bid string) error {
	c, err := kubeClient()
	if err != nil {
		return err
	}

	store := kube.New(c, globalNamespace)
	return store.DeleteBuild(bid, storage.DeleteBuildOptions{
		SkipRunningBuilds: !forceDeleteRunning,
	})
}
