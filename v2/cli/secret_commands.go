package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var secretValFromEnvVarRegex = regexp.MustCompile(`\$\{(\w+)\}`)

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
				nonInteractiveFlag,
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
				&cli.StringFlag{
					Name:    flagFile,
					Aliases: []string{"f"},
					Usage: `A "flat" JSON or YAML file containing secrets as ` +
						`key/value pairs`,
					Required:  false,
					TakesFile: true,
				},
				&cli.StringSliceFlag{
					Name:    flagSet,
					Aliases: []string{"s"},
					Usage: "Set a secret using the specified key=value pair. Secrets " +
						"specified using this flag take precedence over any specified " +
						"using the --file flag",
					Required: false,
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

	client, err := getClient(false)
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
	filename := c.String(flagFile)
	projectID := c.String(flagID)
	kvPairStrs := c.StringSlice(flagSet)

	if filename == "" && len(kvPairStrs) == 0 {
		return errors.New(
			"either a secrets file must be provided using the --file flag or " +
				"key/value pairs must be specified with one or more occurrences of " +
				"the --set flag",
		)
	}

	// We'll make two passes-- we'll parse all the input, both from the file (if
	// applicable) and command line input, into a map first, verifying as we go
	// that the input looks good. Only after we know it's good will we iterate
	// over the k/v pairs in the map to set secrets via the API.

	secrets := map[string]string{}

	if filename != "" {
		secretsFile, err := os.Open(filename)
		if err != nil {
			return errors.Wrapf(err, "error opening secrets file %s", filename)
		}
		defer secretsFile.Close()
		secretBytes, err := ioutil.ReadAll(secretsFile)
		if err != nil {
			return errors.Wrapf(err, "error reading secrets file %s", filename)
		}
		if strings.HasSuffix(filename, ".yaml") ||
			strings.HasSuffix(filename, ".yml") {
			if secretBytes, err = yaml.YAMLToJSON(secretBytes); err != nil {
				return errors.Wrapf(
					err,
					"error converting secrets file %s to JSON",
					filename,
				)
			}
		}
		if err := json.Unmarshal(secretBytes, &secrets); err != nil {
			return errors.Wrapf(err, "error parsing secrets file %s", filename)
		}
		for k, v := range secrets {
			secrets[k] = resolveEnvVars(v)
		}
	}

	for _, kvPairStr := range kvPairStrs {
		kvTokens := strings.SplitN(kvPairStr, "=", 2)
		if len(kvTokens) != 2 || kvTokens[0] == "" || kvTokens[1] == "" {
			return errors.Errorf(
				"secrets set argument %q is formatted incorrectly",
				kvPairStr,
			)
		}
		secrets[kvTokens[0]] = kvTokens[1]
	}

	client, err := getClient(false)
	if err != nil {
		return err
	}

	for k, v := range secrets {
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
		fmt.Printf("Set secret %q for project %q.\n", k, projectID)
	}

	return nil
}

func secretsUnset(c *cli.Context) error {
	projectID := c.String(flagID)
	keys := c.StringSlice(flagUnset)

	client, err := getClient(false)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := client.Core().Projects().Secrets().Unset(
			c.Context,
			projectID,
			key,
		); err != nil {
			return err
		}
		fmt.Printf("Unset secret %q for project %q.\n", key, projectID)
	}

	return nil
}

func resolveEnvVars(val string) string {
	for {
		matches := secretValFromEnvVarRegex.FindStringSubmatch(val)
		if len(matches) != 2 {
			break
		}
		val = strings.Replace(
			val,
			matches[0],
			os.Getenv(matches[1]),
			1,
		)
	}
	return val
}
