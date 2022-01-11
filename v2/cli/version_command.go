package main

import (
	"fmt"

	"github.com/brigadecore/brigade-foundations/file"
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
	fmt.Printf(
		"Brigade client: version %s -- commit %s\n",
		version.Version(),
		version.Commit(),
	)
	// Server version
	if !c.Bool(flagClient) {
		// Skip checking server version if not logged in to any server
		if configFile, err := getConfigPath(); err != nil {
			return err
		} else if exists, err := file.Exists(configFile); err != nil || !exists {
			return err
		}
		client, err := getClient(false)
		if err != nil {
			return err
		}
		serverVersionBytes, err := client.System().UnversionedPing(c.Context)
		if err != nil {
			return err
		}
		fmt.Printf("Brigade API server: %s\n", string(serverVersionBytes))
	}
	return nil
}
