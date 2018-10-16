package commands

import (
	"fmt"
	"io"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"github.com/Azure/brigade/pkg/pipeline/api"
)

const pipelineDefinitionListUsage = `List all pipelines.

Print a list of all of the pipelines in all namespaces.
`

func init() {
	pipelineDefinitions.AddCommand(pipelineDefinitionsList)
}

var pipelineDefinitionsList = &cobra.Command{
	Use:   "list",
	Short: "list pipeline definitions",
	Long:  pipelineDefinitionListUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listPipelineDefinitions(cmd.OutOrStdout())
	},
}

func listPipelineDefinitions(out io.Writer) error {
	c, err := getKubeConfig()
	if err != nil {
		return err
	}

	client, err := api.New(c)
	if err != nil {
		return err
	}
	definitions, err := client.GetPipelineDefinitions("")
	table := uitable.New()
	table.AddRow("NAME", "DESCRIPTION")

	for _, d := range definitions {
		table.AddRow(d.Name, d.Spec.Description)
	}

	fmt.Fprintln(out, table)
	return nil
}
