package commands

import (
	"fmt"
	"io"
	"time"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/Azure/brigade/pkg/storage/kube"
)

const buildListUsage = `List all installed builds.

Print a list of all of the current builds.
`

func init() {
	build.AddCommand(buildList)
}

var buildList = &cobra.Command{
	Use:   "list",
	Short: "list builds",
	Long:  buildListUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listBuilds(cmd.OutOrStdout())
	},
}

func listBuilds(out io.Writer) error {
	c, err := kubeClient()
	if err != nil {
		return err
	}

	store := kube.New(c, globalNamespace)
	bs, err := store.GetBuilds()
	if err != nil {
		return err
	}

	table := uitable.New()
	table.AddRow("ID", "TYPE", "PROVIDER", "PROJECT", "STATUS", "AGE")

	for _, b := range bs {
		status := "???"
		if b.Worker != nil {
			status = b.Worker.Status.String()
		}
		since := duration.ShortHumanDuration(time.Since(b.Worker.EndTime))
		table.AddRow(b.ID, b.Type, b.Provider, b.ProjectID, status, since)
	}
	fmt.Fprintln(out, table)
	return nil
}
