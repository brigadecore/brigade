package main

import (
	"fmt"

	"github.com/brigadecore/brigade-foundations/version"
	"github.com/urfave/cli/v2"
)

var versionCommand = &cli.Command{
	Name:   "version",
	Usage:  "Print Brigade version",
	Action: printVersion,
}

func printVersion(c *cli.Context) error {
	fmt.Printf("Brigade version %s", version.Version())
	return nil
}
