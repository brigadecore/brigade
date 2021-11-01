package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
)

var nonInteractiveFlag = &cli.BoolFlag{
	Name:    flagNonInteractive,
	Aliases: []string{"n"},
	Usage:   "Disable all interactive prompts",
}

// confirmed prompts the user to confirm an irreversible action and returns a
// bool indicating assent (true) or dissent (false).
func confirmed(c *cli.Context) (bool, error) {
	confirmed := c.Bool(flagYes)
	if confirmed {
		return true, nil
	}

	if c.Bool(flagNonInteractive) || !terminal.IsTerminal(int(os.Stdout.Fd())) {
		return false,
			errors.New(
				"In non-interactive mode, this action must be confirmed with the " +
					"--yes flag",
			)
	}

	if err := survey.AskOne(
		&survey.Confirm{
			Message: "This action cannot be undone. Are you sure?",
		},
		&confirmed,
	); err != nil {
		return false, errors.Wrap(err, "error confirming action")
	}
	fmt.Println()
	return confirmed, nil
}

// shouldContinue prompts the user to indicate whether additional query results
// should be fetched using another API request and returns a bool indicating yes
// (true) or no (false).
func shouldContinue(
	c *cli.Context,
	remainingItemCount int64,
	continueVal string,
) (bool, error) {
	outputFormat := strings.ToLower(c.String(flagOutput))

	if remainingItemCount < 1 || continueVal == "" || outputFormat != "table" {
		return false, nil
	}

	// If running with --non-interactive or this isn't a terminal, tell the
	// user how to get the next page of results.
	if c.Bool(flagNonInteractive) || !terminal.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Printf(
			"\n%d results remain. Fetch the next page using --continue %s\n",
			remainingItemCount,
			continueVal,
		)
		return false, nil
	}

	var shouldContinue bool
	fmt.Println()
	if err := survey.AskOne(
		&survey.Confirm{
			Message: fmt.Sprintf(
				"%d results remain. Fetch more?",
				remainingItemCount,
			),
			Default: false,
		},
		&shouldContinue,
	); err != nil {
		return false, errors.Wrap(
			err,
			"error confirming if user wishes to continue",
		)
	}
	return shouldContinue, nil
}
