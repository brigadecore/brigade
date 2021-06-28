package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/authn"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var serviceAccountCommand = &cli.Command{
	Name:    "service-account",
	Aliases: []string{"sa", "service-accounts"},
	Usage:   "Manage service accounts",
	Subcommands: []*cli.Command{
		{
			Name:  "create",
			Usage: "Create a new service account",
			Flags: []cli.Flag{
				// Using custom flagOutput here, to support plaintext output
				// as opposed to table.  Perhaps we can consider adding plaintext
				// support to the common flag definition at a later date.
				&cli.StringFlag{
					Name:    flagOutput,
					Aliases: []string{"o"},
					Usage: "Return output in the specified format; supported formats: " +
						"plaintext, yaml, json",
					Value: flagOutputPlaintext,
				},
				&cli.StringFlag{
					Name:    flagID,
					Aliases: []string{"i"},
					Usage: "Create a service account with the specified ID " +
						"(required)",
					Required: true,
				},
				&cli.StringFlag{
					Name:    flagDescription,
					Aliases: []string{"d"},
					Usage: "Create a service account with the specified " +
						"description (required)",
					Required: true,
				},
			},
			Action: serviceAccountCreate,
		},
		{
			Name:  "get",
			Usage: "Retrieve a service account",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Retrieve the specified service account (required)",
					Required: true,
				},
				cliFlagOutput,
			},
			Action: serviceAccountGet,
		},
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "List service accounts",
			Flags: []cli.Flag{
				cliFlagOutput,
				&cli.StringFlag{
					Name: flagContinue,
					Usage: "Advanced-- passes an opaque value obtained from a " +
						"previous command back to the server to access the next page " +
						"of results",
				},
				nonInteractiveFlag,
			},
			Action: serviceAccountList,
		},
		{
			Name:  "lock",
			Usage: "Lock a service account out of Brigade",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Lock the specified service account (required)",
					Required: true,
				},
			},
			Action: serviceAccountLock,
		},
		{
			Name:  "unlock",
			Usage: "Restore a service account's access to Brigade",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Unlock the specified service account (required)",
					Required: true,
				},
			},
			Action: serviceAccountUnlock,
		},
	},
}

func serviceAccountCreate(c *cli.Context) error {
	output := c.String(flagOutput)
	description := c.String(flagDescription)
	id := c.String(flagID)

	// Validate output format
	// Note: currently not using validateOutputFormat as this command supports
	// plaintext as opposed to table output.
	switch strings.ToLower(output) {
	case flagOutputPlaintext:
	case flagOutputYAML:
	case flagOutputJSON:
	default:
		return errors.Errorf("unknown output format %q", output)
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	token, err := client.Authn().ServiceAccounts().Create(
		c.Context,
		authn.ServiceAccount{
			ObjectMeta: meta.ObjectMeta{
				ID: id,
			},
			Description: description,
		},
	)
	if err != nil {
		return err
	}

	switch strings.ToLower(output) {
	case flagOutputPlaintext:
		fmt.Printf("\nService account %q created with token:\n", id)
		fmt.Printf("\n\t%s\n", token.Value)
		fmt.Println(
			"\nStore this token someplace secure NOW. It cannot be retrieved " +
				"later through any other means.",
		)

	case flagOutputYAML:
		yamlBytes, err := yaml.Marshal(token)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from create service account operation",
			)
		}
		fmt.Println(string(yamlBytes))

	case flagOutputJSON:
		prettyJSON, err := json.MarshalIndent(token, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from create service account operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}

func serviceAccountList(c *cli.Context) error {
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	opts := meta.ListOptions{
		Continue: c.String(flagContinue),
	}

	for {
		serviceAccounts, err := client.Authn().ServiceAccounts().List(
			c.Context,
			nil,
			&opts,
		)
		if err != nil {
			return err
		}

		if len(serviceAccounts.Items) == 0 {
			fmt.Println("No service accounts found.")
			return nil
		}

		switch strings.ToLower(output) {
		case flagOutputTable:
			table := uitable.New()
			table.AddRow("ID", "DESCRIPTION", "AGE", "LOCKED?")
			for _, serviceAccounts := range serviceAccounts.Items {
				table.AddRow(
					serviceAccounts.ID,
					serviceAccounts.Description,
					duration.ShortHumanDuration(time.Since(*serviceAccounts.Created)),
					serviceAccounts.Locked != nil,
				)
			}
			fmt.Println(table)

		case flagOutputYAML:
			yamlBytes, err := yaml.Marshal(serviceAccounts)
			if err != nil {
				return errors.Wrap(
					err,
					"error formatting output from get service accounts operation",
				)
			}
			fmt.Println(string(yamlBytes))

		case flagOutputJSON:
			prettyJSON, err := json.MarshalIndent(serviceAccounts, "", "  ")
			if err != nil {
				return errors.Wrap(
					err,
					"error formatting output from get service accounts operation",
				)
			}
			fmt.Println(string(prettyJSON))
		}

		if shouldContinue, err :=
			shouldContinue(
				c,
				serviceAccounts.RemainingItemCount,
				serviceAccounts.Continue,
			); err != nil {
			return err
		} else if !shouldContinue {
			break
		}

		opts.Continue = serviceAccounts.Continue
	}

	return nil
}

func serviceAccountGet(c *cli.Context) error {
	id := c.String(flagID)
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient()
	if err != nil {
		return err
	}

	serviceAccount, err := client.Authn().ServiceAccounts().Get(c.Context, id)
	if err != nil {
		return err
	}

	switch strings.ToLower(output) {
	case flagOutputTable:
		table := uitable.New()
		table.AddRow("ID", "DESCRIPTION", "AGE", "LOCKED?")
		var age string
		if serviceAccount.Created != nil {
			age = duration.ShortHumanDuration(time.Since(*serviceAccount.Created))
		}
		table.AddRow(
			serviceAccount.ID,
			serviceAccount.Description,
			age,
			serviceAccount.Locked != nil,
		)
		fmt.Println(table)

	case flagOutputYAML:
		yamlBytes, err := yaml.Marshal(serviceAccount)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get service account operation",
			)
		}
		fmt.Println(string(yamlBytes))

	case flagOutputJSON:
		prettyJSON, err := json.MarshalIndent(serviceAccount, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get service account operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}

func serviceAccountLock(c *cli.Context) error {
	id := c.String(flagID)

	client, err := getClient()
	if err != nil {
		return err
	}

	if err := client.Authn().ServiceAccounts().Lock(c.Context, id); err != nil {
		return err
	}

	fmt.Printf("Service account %q locked.\n", id)

	return nil
}

func serviceAccountUnlock(c *cli.Context) error {
	id := c.String(flagID)

	client, err := getClient()
	if err != nil {
		return err
	}

	token, err := client.Authn().ServiceAccounts().Unlock(c.Context, id)
	if err != nil {
		return err
	}

	fmt.Printf(
		"\nService account %q unlocked and a new token has been issued:\n",
		id,
	)
	fmt.Printf("\n\t%s\n", token.Value)
	fmt.Println(
		"\nStore this token someplace secure NOW. It cannot be retrieved " +
			"later through any other means.",
	)

	return nil
}
