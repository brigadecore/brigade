package commands

import (
	"fmt"
	"io"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"github.com/Azure/brigade/pkg/pipeline/api"
)

const pipelineListUsage = `List all pipelines.

Print a list of all of the pipelines in all namespaces.
`

func init() {
	pipeline.AddCommand(pipelineList)
}

var pipelineList = &cobra.Command{
	Use:   "list",
	Short: "list pipelines",
	Long:  pipelineListUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listPipelines(cmd.OutOrStdout())
	},
}

func listPipelines(out io.Writer) error {
	c, err := getKubeConfig()
	if err != nil {
		return err
	}

	client, err := api.New(c)
	if err != nil {
		return err
	}
	pipelines, err := client.GetPipelines(globalNamespace)
	if err != nil {
		return err
	}

	table := uitable.New()
	table.AddRow("NAME", "DESCRIPTION")

	for _, d := range pipelines {
		table.AddRow(d.Name)
	}
	fmt.Fprintln(out, table)
	return nil
}
