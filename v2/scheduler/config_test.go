package main

import (
	"os"
	"testing"

	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
	"github.com/brigadecore/brigade/v2/scheduler/internal/lib/queue/amqp"
	"github.com/stretchr/testify/require"
)

// Note that unit testing in Go does NOT clear environment variables between
// tests, which can sometimes be a pain, but it's fine here-- so each of these
// test functions uses a series of test cases that cumulatively build upon one
// another.

func TestAPIClientConfig(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func()
		assertions func(
			address string,
			token string,
			opts restmachinery.APIClientOptions,
			err error,
		)
	}{
		{
			name:  "API_ADDRESS not set",
			setup: func() {},
			assertions: func(
				_ string,
				_ string,
				_ restmachinery.APIClientOptions,
				err error,
			) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "API_ADDRESS")
			},
		},
		{
			name: "API_TOKEN not set",
			setup: func() {
				os.Setenv("API_ADDRESS", "foo")
			},
			assertions: func(
				_ string,
				_ string,
				_ restmachinery.APIClientOptions,
				err error,
			) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "API_TOKEN")
			},
		},
		{
			name: "SUCCESS not set",
			setup: func() {
				os.Setenv("API_TOKEN", "bar")
				os.Setenv("API_IGNORE_CERT_WARNINGS", "true")
			},
			assertions: func(
				address string,
				token string,
				opts restmachinery.APIClientOptions,
				err error,
			) {
				require.NoError(t, err)
				require.Equal(t, "foo", address)
				require.Equal(t, "bar", token)
				require.True(t, opts.AllowInsecureConnections)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup()
			address, token, opts, err := apiClientConfig()
			testCase.assertions(address, token, opts, err)
		})
	}
}

func TestReaderFactoryConfig(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func()
		assertions func(amqp.ReaderFactoryConfig, error)
	}{
		{
			name:  "AMQP_ADDRESS not set",
			setup: func() {},
			assertions: func(_ amqp.ReaderFactoryConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "AMQP_ADDRESS")
			},
		},
		{
			name: "AMQP_USERNAME not set",
			setup: func() {
				os.Setenv("AMQP_ADDRESS", "foo")
			},
			assertions: func(_ amqp.ReaderFactoryConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "AMQP_USERNAME")
			},
		},
		{
			name: "AMQP_PASSWORD not set",
			setup: func() {
				os.Setenv("AMQP_USERNAME", "bar")
			},
			assertions: func(_ amqp.ReaderFactoryConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "AMQP_PASSWORD")
			},
		},
		{
			name: "success",
			setup: func() {
				os.Setenv("AMQP_PASSWORD", "bat")
			},
			assertions: func(config amqp.ReaderFactoryConfig, err error) {
				require.NoError(t, err)
				require.Equal(
					t,
					amqp.ReaderFactoryConfig{
						Address:  "foo",
						Username: "bar",
						Password: "bat",
					},
					config,
				)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup()
			config, err := readerFactoryConfig()
			testCase.assertions(config, err)
		})
	}
}
