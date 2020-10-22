package main

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// confirmed prompts the user to confirm an irreversible action and returns a
// bool indicating assent (true) or dissent (false).
func confirmed(c *cli.Context) (bool, error) {
	confirmed := c.Bool(flagYes)
	if confirmed {
		return true, nil
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
func shouldContinue(remainingItemCount int64) (bool, error) {
	var shouldContinue bool
	fmt.Println()
	if err := survey.AskOne(
		&survey.Confirm{
			Message: fmt.Sprintf(
				"%d results remain. Fetch more?",
				remainingItemCount,
			),
			Default: true,
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
