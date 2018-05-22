package commands

import (
	"fmt"
	"io"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"github.com/Azure/brigade/pkg/storage/kube"
)

const projectListUsage = `List all installed projects.

Print a list of all of the projects in the given namespace.
`

func init() {
	project.AddCommand(projectList)
}

var projectList = &cobra.Command{
	Use:   "list",
	Short: "list projects",
	Long:  projectListUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listProjects(cmd.OutOrStdout())
	},
}

func listProjects(out io.Writer) error {
	c, err := kubeClient()
	if err != nil {
		return err
	}

	store := kube.New(c, globalNamespace)
	ps, err := store.GetProjects()
	if err != nil {
		return err
	}

	table := uitable.New()
	table.AddRow("NAME", "ID", "REPO")

	for _, p := range ps {
		table.AddRow(p.Name, p.ID, p.Repo.Name)
	}
	fmt.Fprintln(out, table)
	return nil
}
