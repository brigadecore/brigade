package commands

import (
	"errors"
	"io"

	"github.com/Azure/brigade/pkg/storage"
	"github.com/Azure/brigade/pkg/storage/kube"

	"github.com/spf13/cobra"
)

const buildDeleteAllUsage = `Deletes all builds for a project.`

var forceDeleteRunningAll bool

func init() {
	build.AddCommand(buildDeleteAll)
	buildDeleteAll.Flags().BoolVar(&forceDeleteRunningAll, "force", false, "If set, will also delete running builds. Default: false")
}

var buildDeleteAll = &cobra.Command{
	Use:   "delete-all PROJECT_ID",
	Short: "deletes all builds for a project",
	Long:  buildDeleteAllUsage,
	Args:  cobra.ExactArgs(1)
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("project ID is a required argument")
		}
		return deleteAllBuilds(cmd.OutOrStdout(), args[0])
	},
}

func deleteAllBuilds(out io.Writer, projectID string) error {
	c, err := kubeClient()
	if err != nil {
		return err
	}

	store := kube.New(c, globalNamespace)

	proj, err := store.GetProject(projectID)
	if err != nil {
		return err
	}

	builds, err := store.GetProjectBuilds(proj)
	if err != nil {
		return err
	}

	var errDeletes error
	for _, b := range builds {
		err = store.DeleteBuild(b.ID, storage.DeleteBuildOptions{
			SkipRunningBuilds: !forceDeleteRunningAll,
		})
		// loop will continue even if an error is encountered
		// maybe add a check => "if errorCount > threshold then return"?
		if err != nil {
			if errDeletes == nil {
				errDeletes = errors.New("")
			}
			errDeletes = errors.New(errDeletes.Error() + err.Error())
		}
	}

	return errDeletes
}
