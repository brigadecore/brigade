package commands

import (
	"github.com/spf13/cobra"
)

const buildUsage = `Manage builds

Work with Brigade builds.
`

func init() {
	Root.AddCommand(build)
}

var build = &cobra.Command{
	Use:   "build",
	Short: "Manage builds",
	Long:  buildUsage,
}
