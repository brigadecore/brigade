package commands

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/Azure/brigade/pkg/storage/kube"
)

const buildGetUsage = `Get details for a build.

Print the attributes of a build.
`

func init() {
	build.AddCommand(buildGet)
}

var buildGet = &cobra.Command{
	Use:   "get BUILD_ID",
	Short: "get a build",
	Long:  buildGetUsage,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("build ID is a required argument")
		}
		return getBuild(cmd.OutOrStdout(), args[0])
	},
}

func getBuild(out io.Writer, bid string) error {
	c, err := kubeClient()
	if err != nil {
		return err
	}

	store := kube.New(c, globalNamespace)
	b, err := store.GetBuild(bid)
	if err != nil {
		return err
	}

	script := string(b.Script)
	payload := string(b.Payload)

	b.Script = nil
	b.Payload = nil

	data, err := yaml.Marshal(b)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, string(data))
	if globalVerbose {
		fmt.Fprintf(out, "script: |-\n%s\npayload: |- %s\n", script, payload)
	}

	return nil
}
