package main

import (
	"fmt"

	"github.com/brigadecore/brigade-foundations/version"
	"github.com/urfave/cli/v2"
)

var versionCommand = &cli.Command{
	Name:  "version",
	Usage: "Print Brigade version",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    flagClient,
			Aliases: []string{"c"},
			Usage:   "Prints only the Brigade client version",
		},
	},
	Action: printVersion,
}

func printVersion(c *cli.Context) error {
	// Client version
	fmt.Printf("Brigade version %s -- commit %s\n",
		version.Version(), version.Commit())
	// Server version
	if !c.Bool(flagClient) {
		client, err := getClient()
		if err != nil {
			return err
		}
		serverVersionRaw, err := client.System().UnversionedPing(c.Context)
		if err != nil {
			return err
		}
		fmt.Printf("Brigade API Server version %s", string(serverVersionRaw))
	}
	return nil
}
