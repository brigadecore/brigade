package commands

import (
	"github.com/spf13/cobra"
)

const buildUsage = `manage builds

Work with Brigade builds.
`

func init() {
	Root.AddCommand(build)
}

var build = &cobra.Command{
	Use:   "build",
	Short: "manage builds",
	Long:  buildUsage,
}
