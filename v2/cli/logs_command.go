package main

import (
	"fmt"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/urfave/cli/v2"
)

var logsCommand = &cli.Command{
	Name:  "logs",
	Usage: "View worker or job logs",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    flagContainer,
			Aliases: []string{"c"},
			Usage: "View logs from the specified container; if not set, displays " +
				"logs from the worker or job's \"primary\" container",
		},
		&cli.StringFlag{
			Name:     flagID,
			Aliases:  []string{"i", flagEvent, "e"},
			Usage:    "View logs from the specified event",
			Required: true,
		},
		&cli.BoolFlag{
			Name:    flagFollow,
			Aliases: []string{"f"},
			Usage:   "If set, will stream logs until interrupted",
		},
		&cli.StringFlag{
			Name:    flagJob,
			Aliases: []string{"j"},
			Usage: "View logs from the specified job; if not set, displays " +
				"worker logs",
		},
	},
	Action: logs,
}

func logs(c *cli.Context) error {
	eventID := c.String(flagID)
	follow := c.Bool(flagFollow)

	selector := core.LogsSelector{
		Job:       c.String(flagJob),
		Container: c.String(flagContainer),
	}
	opts := core.LogStreamOptions{
		Follow: follow,
	}

	client, err := getClient(c)
	if err != nil {
		return err
	}

	logEntryCh, errCh, err :=
		client.Core().Events().Logs().Stream(c.Context, eventID, &selector, &opts)
	if err != nil {
		return err
	}
	for {
		select {
		case logEntry, ok := <-logEntryCh:
			if ok {
				fmt.Println(logEntry.Message)
			} else {
				// logEntryCh was closed, but want to keep looping through this select
				// in case there are pending errors on the errCh still. nil channels are
				// never readable, so we'll just nil out logEntryCh and move on.
				logEntryCh = nil
			}
		case err, ok := <-errCh:
			if ok {
				return err
			}
			// errCh was closed, but want to keep looping through this select in case
			// there are pending messages on the logEntryCh still. nil channels are
			// never readable, so we'll just nil out errCh and move on.
			errCh = nil
		case <-c.Context.Done():
			return nil
		}
		// If BOTH logEntryCh and errCh were closed, we're done.
		if logEntryCh == nil && errCh == nil {
			return nil
		}
	}
}
