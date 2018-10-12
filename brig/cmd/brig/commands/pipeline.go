package commands

import (
	"github.com/spf13/cobra"
)

const pipelineUsage = `Manage pipelines

Work with Brigade declarative pipelines.
`

func init() {
	Root.AddCommand(pipeline)
}

var pipeline = &cobra.Command{
	Use:   "pipeline",
	Short: "Manage pipelines",
	Long:  pipelineUsage,
}
