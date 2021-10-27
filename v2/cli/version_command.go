package main

import (
	"context"
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
	if c.Bool(flagClient) {
		printClientVersion()
	} else {
		printClientVersion()
		if err := printServerVersion(); err != nil {
			return err
		}
	}
	return nil
}

func printClientVersion() {
	fmt.Printf("Brigade version %s -- commit %s\n",
		version.Version(), version.Commit())
}

func printServerVersion() error {
	client, err := getClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	serverVersionRaw, err := client.System().UnversionedPing(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Brigade API Server version %s", string(serverVersionRaw))
	return nil
}
