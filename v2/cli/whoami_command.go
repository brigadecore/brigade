package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var whoAmICommand = &cli.Command{
	Name:        "whoami",
	Usage:       "Return the current user's identity",
	Description: "Return the current user's identity.",
	Flags:       []cli.Flag{cliFlagOutput},
	Action:      whoAmI,
}

func whoAmI(c *cli.Context) error {
	output := c.String(flagOutput)

	client, err := getClient(false)
	if err != nil {
		return err
	}

	ref, err := client.Authn().WhoAmI(c.Context)
	if err != nil {
		return err
	}

	switch strings.ToLower(output) {
	case flagOutputTable:
		table := uitable.New()
		table.AddRow("PRINCIPAL TYPE", "ID")
		table.AddRow(ref.Type, ref.ID)
		fmt.Println(table)
	case flagOutputYAML:
		yamlBytes, err := yaml.Marshal(ref)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from whoami operation",
			)
		}
		fmt.Println(string(yamlBytes))
	case flagOutputJSON:
		prettyJSON, err := json.MarshalIndent(ref, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from whoami operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}
