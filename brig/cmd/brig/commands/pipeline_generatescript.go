package commands

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/Azure/brigade/pkg/pipeline/api"
)

const pipelineGenerateScriptUsage = `Generate script for pipeline.

Generates the javascript for the pipeline.
`

func init() {
	pipeline.AddCommand(pipelineGenerateScript)
}

var pipelineGenerateScript = &cobra.Command{
	Use:   "generatescript",
	Short: "generates pipeline script",
	Long:  pipelineGenerateScriptUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("pipeline name is a required argument")
		}
		return generateScript(cmd.OutOrStdout(), args[0])
	},
}

func generateScript(out io.Writer, name string) error {
	c, err := getKubeConfig()
	if err != nil {
		return err
	}

	client, err := api.New(c)
	if err != nil {
		return err
	}

	script, err := client.GenerateScript(name, globalNamespace)
	if err != nil {
		return fmt.Errorf("Failed to generate script: %v", err)
	}

	fmt.Print(script)
	return nil
}
