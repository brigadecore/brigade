package commands

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/Azure/brigade/pkg/storage/kube"
)

const projectGetUsage = `Get details for a project.

Print the attributes of a project.

The PROJECT may either be the ID or the name of a project.
`

func init() {
	project.AddCommand(projectGet)
}

var projectGet = &cobra.Command{
	Use:   "get PROJECT",
	Short: "get a project",
	Long:  projectGetUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("project name is a required argument")
		}
		return getProject(cmd.OutOrStdout(), args[0])
	},
}

func getProject(out io.Writer, name string) error {
	c, err := kubeClient()
	if err != nil {
		return err
	}

	store := kube.New(c, globalNamespace)
	p, err := store.GetProject(name)
	if err != nil {
		return err
	}

	bytes, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, string(bytes))
	return nil
}
