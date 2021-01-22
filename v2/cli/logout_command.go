package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var logoutCommand = &cli.Command{
	Name:   "logout",
	Usage:  "Log out of Brigade",
	Action: logout,
}

func logout(c *cli.Context) error {
	client, err := getClient(c)
	if err != nil {
		return err
	}

	// We're ignoring any error here because even if the session wasn't found
	// and deleted server-side, we still want to move on to destroying the local
	// token.
	client.Authn().Sessions().Delete(c.Context) // nolint: errcheck

	if err := deleteConfig(); err != nil {
		return errors.Wrap(err, "error deleting configuration")
	}

	fmt.Println("Logout was successful.")

	return nil
}
