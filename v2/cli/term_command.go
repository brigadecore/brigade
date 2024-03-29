package main

import (
	terminal "github.com/brigadecore/brigade/v2/cli/term"
	"github.com/rivo/tview"
	"github.com/urfave/cli/v2"
)

var termCommand = &cli.Command{
	Name:        "term",
	Usage:       "Run the interactive terminal",
	Description: "Run an EXPERIMENTAL interactive terminal.",
	Action:      term,
}

func term(c *cli.Context) error {
	client, err := getClient(true)
	if err != nil {
		return err
	}
	app := tview.NewApplication()
	app.SetRoot(
		tview.NewFlex().SetDirection(tview.FlexRow).AddItem(
			terminal.NewPageRouter(client, app),
			0,
			1, // Proportionate width-- 100%
			true,
		),
		true,
	)
	return app.Run()
}
