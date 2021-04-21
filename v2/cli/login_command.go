package main

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/bacongobbler/browser"
	"github.com/brigadecore/brigade/sdk/v2/authn"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
)

var loginCommand = &cli.Command{
	Name:  "login",
	Usage: "Log in to Brigade",
	Description: "By default, initiates authentication using OpenID " +
		"Connect. This may not be supported by all Brigade API servers.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    flagServer,
			Aliases: []string{"s"},
			Usage: "Log into the API server at the specified address " +
				"(required)",
			Required: true,
		},
		&cli.BoolFlag{
			Name:    flagBrowse,
			Aliases: []string{"b"},
			Usage: "Use the system's default web browser to complete " +
				"authentication; not applicable when --root is used",
		},
		&cli.StringFlag{
			Name:    flagPassword,
			Aliases: []string{"p"},
			Usage: "Specify the password for non-interactive root user login; " +
				"only applicable when --root is used",
		},
		&cli.BoolFlag{
			Name:    flagRoot,
			Aliases: []string{"r"},
			Usage:   "Log in as the root user; does not use OpenID Connect",
		},
	},
	Action: login,
}

func login(c *cli.Context) error {
	address := c.String(flagServer)
	browseToAuthURL := c.Bool(flagBrowse)
	password := c.String(flagPassword)
	rootLogin := c.Bool(flagRoot)

	client := authn.NewSessionsClient(
		address,
		"",
		&restmachinery.APIClientOptions{
			AllowInsecureConnections: c.Bool(flagInsecure),
		},
	)

	var tokenStr, authURL string

	if rootLogin {

		if password == "" &&
			(c.Bool(flagNonInteractive) ||
				!terminal.IsTerminal(int(os.Stdout.Fd()))) {
			return errors.New(
				"--root and --non-interactive flags together require the --password " +
					"flag to also be used",
			)
		}

		for {
			if password != "" {
				break
			}
			prompt := &survey.Password{
				Message: "Root user password",
			}
			if err := survey.AskOne(prompt, &password); err != nil {
				return err
			}
		}

		token, err := client.CreateRootSession(c.Context, password)
		if err != nil {
			return err
		}
		tokenStr = token.Value
	} else {
		oidcAuthDetails, err :=
			client.CreateUserSession(c.Context, nil)
		if err != nil {
			return err
		}
		authURL = oidcAuthDetails.AuthURL
		tokenStr = oidcAuthDetails.Token
	}

	if err := saveConfig(
		config{
			APIAddress: address,
			APIToken:   tokenStr,
		},
	); err != nil {
		return errors.Wrap(err, "error persisting configuration")
	}

	if rootLogin {
		fmt.Println("\nYou are logged in as the root user.")
		return nil
	}

	if browseToAuthURL {
		if err := browser.Open(authURL); err != nil {
			return errors.Wrapf(
				err,
				"Error opening authentication URL using the system's default web "+
					"browser.\n\nPlease visit  %s  to complete authentication.\n",
				authURL,
			)
		}
		return nil
	}

	fmt.Printf("Please visit  %s  to complete authentication.\n", authURL)

	return nil
}
