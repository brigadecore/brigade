package main

import (
	"context"

	"github.com/pkg/errors"

	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

func getClient(testConn bool) (sdk.APIClient, error) {
	cfg, err := getConfig()
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error getting brigade client: error retrieving configuration",
		)
	}
	client := sdk.NewAPIClient(
		cfg.APIAddress,
		cfg.APIToken,
		&restmachinery.APIClientOptions{
			AllowInsecureConnections: cfg.IgnoreCertErrors,
		},
	)
	if testConn {
		_, err = client.System().UnversionedPing(context.Background())
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"error getting brigade client: error pinging API server",
			)
		}
	}
	return client, nil
}
