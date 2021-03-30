package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var secretsCommand = &cli.Command{
	Name:    "secret",
	Aliases: []string{"secrets"},
	Usage:   "Manage project secrets",
	Subcommands: []*cli.Command{
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "List project secrets; values are always redacted",
			Flags: []cli.Flag{
				cliFlagOutput,
				&cli.StringFlag{
					Name: flagContinue,
					Usage: "Advanced-- passes an opaque value obtained from a " +
						"previous command back to the server to access the next page " +
						"of results",
				},
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i", flagProject, "p"},
					Usage:    "Retrieve secrets for the specified project (required)",
					Required: true,
				},
			},
			Action: secretsList,
		},
		{
			Name:  "set",
			Usage: "Define or redefine the value of one or more secrets",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i", flagProject, "p"},
					Usage:    "Set secrets for the specified project (required)",
					Required: true,
				},
				&cli.StringSliceFlag{
					Name:    flagSet,
					Aliases: []string{"s"},
					Usage: "Set a secret using the specified key=value pair " +
						"(required)",
					Required: true,
				},
			},
			Action: secretsSet,
		},
		{
			Name:  "unset",
			Usage: "Clear the value of one or more secrets",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i", flagProject, "p"},
					Usage:    "Clear secrets for the specified project",
					Required: true,
				},
				&cli.StringSliceFlag{
					Name:     flagUnset,
					Aliases:  []string{"u"},
					Usage:    "Clear a secret having the specified key (required)",
					Required: true,
				},
			},
			Action: secretsUnset,
		},
	},
}

func secretsList(c *cli.Context) error {
	output := c.String(flagOutput)
	projectID := c.String(flagID)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return err
	}

	opts := meta.ListOptions{
		Continue: c.String(flagContinue),
	}

	for {
		secrets, err :=
			client.Core().Projects().Secrets().List(c.Context, projectID, &opts)
		if err != nil {
			return err
		}

		switch strings.ToLower(output) {
		case flagOutputTable:
			table := uitable.New()
			table.AddRow("KEY", "VALUE")
			for _, secret := range secrets.Items {
				table.AddRow(secret.Key, "*** REDACTED ***")
			}
			fmt.Println(table)

		case flagOutputYAML:
			yamlBytes, err := yaml.Marshal(secrets)
			if err != nil {
				return errors.Wrap(
					err,
					"error formatting output from get secrets operation",
				)
			}
			fmt.Println(string(yamlBytes))

		case flagOutputJSON:
			prettyJSON, err := json.MarshalIndent(secrets, "", "  ")
			if err != nil {
				return errors.Wrap(
					err,
					"error formatting output from get secrets operation",
				)
			}
			fmt.Println(string(prettyJSON))
		}

		if shouldContinue, err :=
			shouldContinue(
				c,
				secrets.RemainingItemCount,
				secrets.Continue,
			); err != nil {
			return err
		} else if !shouldContinue {
			break
		}

		opts.Continue = secrets.Continue
	}

	return nil
}

func secretsSet(c *cli.Context) error {
	projectID := c.String(flagID)
	kvPairsStr := c.StringSlice(flagSet)

	// We'll make two passes-- we'll parse all the input into a map first,
	// verifying as we go that the input looks good. Only after we know it's good
	// will we iterate over the k/v pairs in the map to set secrets via the API.

	kvPairs := map[string]string{}
	for _, kvPairStr := range kvPairsStr {
		kvTokens := strings.SplitN(kvPairStr, "=", 2)
		if len(kvTokens) != 2 {
			return errors.New("secrets set argument %q is formatted incorrectly")
		}
		kvPairs[kvTokens[0]] = kvTokens[1]
	}

	client, err := getClient(c)
	if err != nil {
		return err
	}

	// Note: The pattern for setting multiple secrets RESTfully in one shot isn't
	// clear, so for now we settle for iterating over the secrets and making an
	// API call for each one. This can be revisited in the future if someone is
	// aware of or discovers the right pattern for this.
	for k, v := range kvPairs {
		secret := core.Secret{
			Key:   k,
			Value: v,
		}
		if err := client.Core().Projects().Secrets().Set(
			c.Context,
			projectID,
			secret,
		); err != nil {
			return err
		}
	}

	fmt.Printf("Set secrets for project %q.\n", projectID)

	return nil
}

func secretsUnset(c *cli.Context) error {
	projectID := c.String(flagID)
	keys := c.StringSlice(flagUnset)

	client, err := getClient(c)
	if err != nil {
		return err
	}

	// Note: The pattern for deleting multiple secrets RESTfully in one shot isn't
	// clear, so for now we settle for iterating over the secrets and making an
	// API call for each one. This can be revisited in the future if someone is
	// aware of or discovers the right pattern for this.
	for _, key := range keys {
		if err := client.Core().Projects().Secrets().Unset(
			c.Context,
			projectID,
			key,
		); err != nil {
			return err
		}
	}

	fmt.Printf("Unset secrets for project %q.\n", projectID)

	return nil
}
