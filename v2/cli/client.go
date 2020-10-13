package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

func getClient(c *cli.Context) (sdk.APIClient, error) {
	cfg, err := getConfig()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error getting brigade client: error retrieving configuration",
		)
	}
	return sdk.NewAPIClient(
		cfg.APIAddress,
		cfg.APIToken,
		&restmachinery.APIClientOptions{
			AllowInsecureConnections: c.Bool(flagInsecure),
		},
	), nil
}
