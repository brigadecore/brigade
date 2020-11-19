package main

import (
	"context"
	"os"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/core/kubernetes"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue/amqp"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery/authn"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

// Note that unit testing in Go does NOT clear environment variables between
// tests, which can sometimes be a pain, but it's fine here-- so each of these
// test functions uses a series of test cases that cumulatively build upon one
// another.

func TestDatabase(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func()
		assertions func(*mongo.Database, error)
	}{
		{
			name:  "DATABASE_USERNAME not set",
			setup: func() {},
			assertions: func(_ *mongo.Database, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "DATABASE_USERNAME")
			},
		},
		{
			name: "DATABASE_PASSWORD not set",
			setup: func() {
				os.Setenv("DATABASE_USERNAME", "jarvis")
			},
			assertions: func(_ *mongo.Database, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "DATABASE_PASSWORD")
			},
		},
		{
			name: "DATABASE_HOST not set",
			setup: func() {
				os.Setenv("DATABASE_PASSWORD", "yourenotironmaniam")
			},
			assertions: func(_ *mongo.Database, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "DATABASE_HOST")
			},
		},
		{
			name: "DATABASE_PORT not an int",
			setup: func() {
				os.Setenv("DATABASE_HOST", "http://localhost")
				os.Setenv("DATABASE_PORT", "foo")
			},
			assertions: func(_ *mongo.Database, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as an int")
				require.Contains(t, err.Error(), "DATABASE_PORT")
			},
		},
		{
			name: "DATABASE_NAME not set",
			setup: func() {
				os.Setenv("DATABASE_PORT", "27017")
			},
			assertions: func(_ *mongo.Database, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "DATABASE_NAME")
			},
		},
		{
			name: "DATABASE_REPLICA_SET not set",
			setup: func() {
				os.Setenv("DATABASE_NAME", "brigade")
			},
			assertions: func(_ *mongo.Database, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "DATABASE_REPLICA_SET")
			},
		},
		{
			name: "success",
			setup: func() {
				os.Setenv("DATABASE_REPLICA_SET", "rs0")
			},
			assertions: func(database *mongo.Database, err error) {
				require.NoError(t, err)
				require.NotNil(t, database)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup()
			database, err := database(context.Background())
			testCase.assertions(database, err)
		})
	}
}

func TestWriterFactoryConfig(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func()
		assertions func(amqp.WriterFactoryConfig, error)
	}{
		{
			name:  "AMQP_ADDRESS not set",
			setup: func() {},
			assertions: func(_ amqp.WriterFactoryConfig, err error) {
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
			assertions: func(_ amqp.WriterFactoryConfig, err error) {
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
			assertions: func(_ amqp.WriterFactoryConfig, err error) {
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
			assertions: func(config amqp.WriterFactoryConfig, err error) {
				require.NoError(t, err)
				require.Equal(
					t,
					amqp.WriterFactoryConfig{
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
			config, err := writerFactoryConfig()
			testCase.assertions(config, err)
		})
	}
}

func TestSubstrateConfig(t *testing.T) {
	const testAPIAddress = "http://localhost"
	const testDefaultWorkerImage = "brigadecore/brigade-worker:2.0.0"
	const testDefaultWorkerImagePullPolicy = core.ImagePullPolicy("IfNotPresent")
	const testWorkspaceStorageClass = "nfs"
	testCases := []struct {
		name       string
		setup      func()
		assertions func(kubernetes.SubstrateConfig, error)
	}{
		{
			name:  "API_ADDRESS not set",
			setup: func() {},
			assertions: func(_ kubernetes.SubstrateConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "API_ADDRESS")
			},
		},
		{
			name: "DEFAULT_WORKER_IMAGE not set",
			setup: func() {
				os.Setenv("API_ADDRESS", testAPIAddress)
			},
			assertions: func(_ kubernetes.SubstrateConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "DEFAULT_WORKER_IMAGE")
			},
		},
		{
			name: "DEFAULT_WORKER_IMAGE_PULL_POLICY not set",
			setup: func() {
				os.Setenv("DEFAULT_WORKER_IMAGE", testDefaultWorkerImage)
			},
			assertions: func(_ kubernetes.SubstrateConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "DEFAULT_WORKER_IMAGE_PULL_POLICY")
			},
		},
		{
			name: "WORKSPACE_STORAGE_CLASS not set",
			setup: func() {
				os.Setenv(
					"DEFAULT_WORKER_IMAGE_PULL_POLICY",
					string(testDefaultWorkerImagePullPolicy),
				)
			},
			assertions: func(_ kubernetes.SubstrateConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "WORKSPACE_STORAGE_CLASS")
			},
		},
		{
			name: "success",
			setup: func() {
				os.Setenv("WORKSPACE_STORAGE_CLASS", testWorkspaceStorageClass)
			},
			assertions: func(config kubernetes.SubstrateConfig, err error) {
				require.NoError(t, err)
				require.Equal(t, testAPIAddress, config.APIAddress)
				require.Equal(t, testDefaultWorkerImage, config.DefaultWorkerImage)
				require.Equal(
					t,
					testDefaultWorkerImagePullPolicy,
					config.DefaultWorkerImagePullPolicy,
				)
				require.Equal(
					t,
					testWorkspaceStorageClass,
					config.WorkspaceStorageClass,
				)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup()
			config, err := substrateConfig()
			testCase.assertions(config, err)
		})
	}
}

func TestSessionsServiceConfig(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func()
		assertions func(error)
	}{
		{
			name: "ROOT_USER_ENABLED not parsable as bool",
			setup: func() {
				os.Setenv("ROOT_USER_ENABLED", "aw hell no")
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as a bool")
				require.Contains(t, err.Error(), "ROOT_USER_ENABLED")
			},
		},
		{
			name: "OIDC_ENABLED not parsable as bool",
			setup: func() {
				os.Setenv("ROOT_USER_ENABLED", "true")
				os.Setenv("OIDC_ENABLED", "aw hell no")
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as a bool")
				require.Contains(t, err.Error(), "OIDC_ENABLED")
			},
		},
		{
			name: "ROOT_USER_PASSWORD required but not set",
			setup: func() {
				os.Setenv("OIDC_ENABLED", "true")
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "ROOT_USER_PASSWORD")
			},
		},
		{
			name: "OIDC_PROVIDER_URL required but not set",
			setup: func() {
				os.Setenv("ROOT_USER_PASSWORD", "12345")
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "OIDC_PROVIDER_URL")
			},
		},
		{
			name: "OIDC_CLIENT_ID required but not set",
			setup: func() {
				// This needs to be a legit URL unless we want to mock out all kinds of
				// OIDC stuff. There's no real harm in this. This is an AAD tenant owned
				// by krancour.
				os.Setenv(
					"OIDC_PROVIDER_URL",
					"https://login.microsoftonline.com/cc18ecf3-7acb-4d14-ba43-fc25dc310191/v2.0", // nolint: lll
				)
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "OIDC_CLIENT_ID")
			},
		},
		{
			name: "OIDC_CLIENT_SECRET required but not set",
			setup: func() {
				// Even thought we used a real provider URL, this client ID is made up
				os.Setenv("OIDC_CLIENT_ID", "hal9000")
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "OIDC_CLIENT_SECRET")
			},
		},
		{
			name: "OIDC_REDIRECT_URL_BASE required but not set",
			setup: func() {
				// Even thought we used a real provider URL, this client secret is made
				// up
				os.Setenv("OIDC_CLIENT_SECRET", "hello, dave")
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "OIDC_REDIRECT_URL_BASE")
			},
		},
		{
			name: "success",
			setup: func() {
				os.Setenv("OIDC_REDIRECT_URL_BASE", "https://localhost")
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup()
			_, err := sessionsServiceConfig(context.Background())
			testCase.assertions(err)
		})
	}
}

func TestTokenAuthFilterConfig(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func()
		assertions func(authn.TokenAuthFilterConfig, error)
	}{
		{
			name: "ROOT_USER_ENABLED not parsable as bool",
			setup: func() {
				os.Setenv("ROOT_USER_ENABLED", "yuppers")
			},
			assertions: func(_ authn.TokenAuthFilterConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as a bool")
				require.Contains(t, err.Error(), "ROOT_USER_ENABLED")
			},
		},
		{
			name: "OIDC_ENABLED not parsable as bool",
			setup: func() {
				os.Setenv("ROOT_USER_ENABLED", "true")
				os.Setenv("OIDC_ENABLED", "yuppers")
			},
			assertions: func(_ authn.TokenAuthFilterConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as a bool")
				require.Contains(t, err.Error(), "OIDC_ENABLED")
			},
		},
		{
			name: "SCHEDULER_TOKEN not set",
			setup: func() {
				os.Setenv("OIDC_ENABLED", "true")
			},
			assertions: func(_ authn.TokenAuthFilterConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "SCHEDULER_TOKEN")
			},
		},
		{
			name: "OBSERVER_TOKEN not set",
			setup: func() {
				os.Setenv("SCHEDULER_TOKEN", "foo")
			},
			assertions: func(_ authn.TokenAuthFilterConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "OBSERVER_TOKEN")
			},
		},
		{
			name: "success",
			setup: func() {
				os.Setenv("OBSERVER_TOKEN", "bar")
			},
			assertions: func(config authn.TokenAuthFilterConfig, err error) {
				require.NoError(t, err)
				require.NotNil(t, config.FindUserFn)
				require.True(t, config.RootUserEnabled)
				require.True(t, config.OpenIDConnectEnabled)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup()
			config, err := tokenAuthFilterConfig(
				func(context.Context, string) (authx.User, error) {
					return authx.User{}, nil
				},
			)
			testCase.assertions(config, err)
		})
	}
}

func TestServerConfig(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func()
		assertions func(restmachinery.ServerConfig, error)
	}{
		{
			name: "API_SERVER_PORT not parsable as int",
			setup: func() {
				os.Setenv("API_SERVER_PORT", "foo")
			},
			assertions: func(_ restmachinery.ServerConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as an int")
				require.Contains(t, err.Error(), "API_SERVER_PORT")
			},
		},
		{
			name: "TLS_ENABLED not parsable as bool",
			setup: func() {
				os.Setenv("API_SERVER_PORT", "8080")
				os.Setenv("TLS_ENABLED", "nope")
			},
			assertions: func(_ restmachinery.ServerConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as a bool")
				require.Contains(t, err.Error(), "TLS_ENABLED")
			},
		},
		{
			name: "TLS_CERT_PATH required but not set",
			setup: func() {
				os.Setenv("TLS_ENABLED", "true")
			},
			assertions: func(_ restmachinery.ServerConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "TLS_CERT_PATH")
			},
		},
		{
			name: "TLS_KEY_PATH required but not set",
			setup: func() {
				os.Setenv("TLS_CERT_PATH", "/var/ssl/cert")
			},
			assertions: func(_ restmachinery.ServerConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "TLS_KEY_PATH")
			},
		},
		{
			name: "success",
			setup: func() {
				os.Setenv("TLS_KEY_PATH", "/var/ssl/key")
			},
			assertions: func(config restmachinery.ServerConfig, err error) {
				require.NoError(t, err)
				require.Equal(
					t,
					restmachinery.ServerConfig{
						Port:        8080,
						TLSEnabled:  true,
						TLSCertPath: "/var/ssl/cert",
						TLSKeyPath:  "/var/ssl/key",
					},
					config,
				)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup()
			config, err := serverConfig()
			testCase.assertions(config, err)
		})
	}
}
