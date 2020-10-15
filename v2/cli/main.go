package main

import (
	"fmt"
	"os"

	"github.com/brigadecore/brigade/v2/internal/signals"
	"github.com/brigadecore/brigade/v2/internal/version"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "Brigade"
	app.Usage = "Event Driven Scripting for Kubernetes"
	app.Version = fmt.Sprintf(
		"%s -- commit %s",
		version.Version(),
		version.Commit(),
	)
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    flagInsecure,
			Aliases: []string{"k"},
			Usage:   "Allow insecure API server connections when using HTTPS",
		},
	}
	app.Commands = []*cli.Command{
		loginCommand,
		logoutCommand,
		projectCommand,
	}
	fmt.Println()
	if err := app.RunContext(signals.Context(), os.Args); err != nil {
		fmt.Printf("\n%s\n\n", err)
		os.Exit(1)
	}
	fmt.Println()
}
