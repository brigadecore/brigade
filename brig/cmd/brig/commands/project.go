package commands

import (
	"github.com/spf13/cobra"
)

const projectUsage = `manage projects

Work with Brigade projects.
`

func init() {
	Root.AddCommand(project)
}

var project = &cobra.Command{
	Use:   "project",
	Short: "manage projects",
	Long:  projectUsage,
}
