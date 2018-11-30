package commands

import (
	"fmt"
	"io"
	"time"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage/kube"
)

const buildListUsage = `List all installed builds.

Print a list of the current builds, from latest to oldest. By default it will fetch all the builds, use --count to get a subset of them.
`

var buildListCount int

func init() {
	build.AddCommand(buildList)
	buildList.Flags().IntVar(&buildListCount, "count", 0, "The maximum number of builds to return. 0 for all")
}

var buildList = &cobra.Command{
	Use:   "list [project]",
	Short: "list builds",
	Long:  buildListUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		proj := ""
		if len(args) > 0 {
			proj = args[0]
		}
		return listBuilds(cmd.OutOrStdout(), proj)
	},
}

func listBuilds(out io.Writer, project string) error {
	c, err := kubeClient()
	if err != nil {
		return err
	}

	store := kube.New(c, globalNamespace)

	var bs []*brigade.Build
	if project == "" {
		bs, err = store.GetBuilds()
		if err != nil {
			return err
		}
	} else {
		proj, err := store.GetProject(project)
		if err != nil {
			return err
		}

		bs, err = store.GetProjectBuilds(proj)
		if err != nil {
			return err
		}
	}

	table := uitable.New()
	table.AddRow("ID", "TYPE", "PROVIDER", "PROJECT", "STATUS", "AGE")

	for i := len(bs) - 1; i >= 0; i-- {

		if buildListCount > 0 && len(bs)-i-1 >= buildListCount {
			break
		}

		b := bs[i]

		status := "???"
		since := "???"
		if b.Worker != nil {
			status = b.Worker.Status.String()
			if b.Worker.Status == brigade.JobSucceeded || b.Worker.Status == brigade.JobFailed {
				since = duration.ShortHumanDuration(time.Since(b.Worker.EndTime))
			}
		}
		table.AddRow(b.ID, b.Type, b.Provider, b.ProjectID, status, since)
	}
	fmt.Fprintln(out, table)
	return nil
}
