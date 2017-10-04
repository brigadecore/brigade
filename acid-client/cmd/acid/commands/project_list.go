package commands

import (
	"fmt"
	"io"

	"github.com/deis/acid/pkg/storage/kube"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
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
		return listProjects(cmd.OutOrStdout(), globalNamespace)
	},
}

func listProjects(out io.Writer, ns string) error {
	c, err := kube.GetClient("", kubeConfigPath())
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
