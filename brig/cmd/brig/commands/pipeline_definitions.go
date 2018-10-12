package commands

import (
	"github.com/spf13/cobra"
)

const pipelineDefinitionsUsage = `Manage pipeline definitions

Work with Brigade pipeline definitions.
`

func init() {
	pipeline.AddCommand(pipelineDefinitions)
}

var pipelineDefinitions = &cobra.Command{
	Use:   "definitions",
	Short: "Manage pipeline definitions",
	Long:  pipelineDefinitionsUsage,
}
