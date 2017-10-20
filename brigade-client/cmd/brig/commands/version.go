package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/deis/brigade/pkg/version"
)

func init() {
	Root.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print brig version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
	},
}
