package commands

import (
	"fmt"
	"io"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

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
	c, err := kube.GetClient("", kubeConfigPath())
	if err != nil {
		return err
	}

	store := kube.New(c, globalNamespace)
	bs, err := store.GetBuilds()
	if err != nil {
		return err
	}

	table := uitable.New()
	table.AddRow("ID", "TYPE", "PROVIDER", "PROJECT")

	for _, b := range bs {
		table.AddRow(b.ID, b.Type, b.Provider, b.ProjectID)
	}
	fmt.Fprintln(out, table)
	return nil
}
