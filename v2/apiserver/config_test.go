package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	"github.com/brigadecore/brigade/v2/apiserver/internal/api/kubernetes"
	"github.com/brigadecore/brigade/v2/apiserver/internal/api/rest"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue/amqp"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

// Note that unit testing in Go does NOT clear environment variables between
// tests, which can sometimes be a pain, but it's fine here-- so each of these
// test functions uses a series of test cases that cumulatively build upon one
// another.

func TestDatabaseConnection(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func()
		assertions func(*mongo.Database, error)
	}{
		{
			name:  "DATABASE_HOSTS not set",
			setup: func() {},
			assertions: func(_ *mongo.Database, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "DATABASE_HOSTS")
			},
		},
		{
			name: "DATABASE_USERNAME not set",
			setup: func() {
				t.Setenv("DATABASE_HOSTS", "localhost")
			},
			assertions: func(_ *mongo.Database, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "DATABASE_USERNAME")
			},
		},
		{
			name: "DATABASE_PASSWORD not set",
			setup: func() {
				t.Setenv("DATABASE_USERNAME", "jarvis")
			},
			assertions: func(_ *mongo.Database, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "DATABASE_PASSWORD")
			},
		},
		{
			name: "DATABASE_NAME not set",
			setup: func() {
				t.Setenv("DATABASE_PASSWORD", "yourenotironmaniam")
			},
			assertions: func(_ *mongo.Database, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "DATABASE_NAME")
			},
		},
		{
			name: "success",
			setup: func() {
				t.Setenv("DATABASE_NAME", "brigade")
				t.Setenv("DATABASE_REPLICA_SET", "rs0")
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
			database, err := databaseConnection(context.Background())
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
				t.Setenv("AMQP_ADDRESS", "foo")
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
				t.Setenv("AMQP_USERNAME", "bar")
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
				t.Setenv("AMQP_PASSWORD", "bat")
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
	const testBrigadeID = "4077th"
	const testAPIAddress = "http://localhost"
	const testGitInitializerImage = "brigadecore/brigade2-git-initializer:2.0.0"
	const testGitInitializerImagePullPolicy = api.ImagePullPolicy("IfNotPresent")
	const testDefaultWorkerImage = "brigadecore/brigade2-worker:2.0.0"
	const testDefaultWorkerImagePullPolicy = api.ImagePullPolicy("IfNotPresent")
	const testWorkspaceStorageClass = "nfs"
	testCases := []struct {
		name       string
		setup      func()
		assertions func(kubernetes.SubstrateConfig, error)
	}{
		{
			name:  "BRIGADE_ID not set",
			setup: func() {},
			assertions: func(_ kubernetes.SubstrateConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "BRIGADE_ID")
			},
		},
		{
			name: "API_ADDRESS not set",
			setup: func() {
				t.Setenv("BRIGADE_ID", testBrigadeID)
			},
			assertions: func(_ kubernetes.SubstrateConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "API_ADDRESS")
			},
		},
		{
			name: "GIT_INITIALIZER_IMAGE not set",
			setup: func() {
				t.Setenv("API_ADDRESS", testAPIAddress)
			},
			assertions: func(_ kubernetes.SubstrateConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "GIT_INITIALIZER_IMAGE")
			},
		},
		{
			name: "GIT_INITIALIZER_IMAGE_PULL_POLICY not set",
			setup: func() {
				t.Setenv("GIT_INITIALIZER_IMAGE", testGitInitializerImage)
			},
			assertions: func(_ kubernetes.SubstrateConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "GIT_INITIALIZER_IMAGE_PULL_POLICY")
			},
		},
		{
			name: "DEFAULT_WORKER_IMAGE not set",
			setup: func() {
				t.Setenv(
					"GIT_INITIALIZER_IMAGE_PULL_POLICY",
					string(testGitInitializerImagePullPolicy),
				)
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
				t.Setenv("DEFAULT_WORKER_IMAGE", testDefaultWorkerImage)
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
				t.Setenv(
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
				t.Setenv("WORKSPACE_STORAGE_CLASS", testWorkspaceStorageClass)
			},
			assertions: func(config kubernetes.SubstrateConfig, err error) {
				require.NoError(t, err)
				require.Equal(t, testBrigadeID, config.BrigadeID)
				require.Equal(t, testAPIAddress, config.APIAddress)
				require.Equal(t, testGitInitializerImage, config.GitInitializerImage)
				require.Equal(
					t,
					testGitInitializerImagePullPolicy,
					config.GitInitializerImagePullPolicy,
				)
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

func TestThirdPartyAuthHelper(t *testing.T) {
	// Set up test OIDC auth server
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/.well-known/openid-configuration", r.URL.Path)
				require.Equal(t, 0, len(r.URL.Query()))
				bodyBytes, err := json.Marshal(
					struct {
						Issuer string `json:"issuer"`
					}{
						Issuer: fmt.Sprintf("http://%s", r.Host),
					},
				)
				require.NoError(t, err)
				fmt.Fprintln(w, string(bodyBytes))
				w.(http.Flusher).Flush()
			},
		),
	)
	defer server.Close()

	testCases := []struct {
		name       string
		setup      func()
		assertions func(api.ThirdPartyAuthHelper, error)
	}{
		{
			name: "THIRD_PARTY_AUTH_STRATEGY has invalid value",
			setup: func() {
				t.Setenv("THIRD_PARTY_AUTH_STRATEGY", "bogus")
			},
			assertions: func(_ api.ThirdPartyAuthHelper, err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"unrecognized THIRD_PARTY_AUTH_STRATEGY",
				)
			},
		},
		{
			name: "OIDC_PROVIDER_URL required but not set",
			setup: func() {
				t.Setenv("THIRD_PARTY_AUTH_STRATEGY", thirdPartyAuthStrategyOIDC)
			},
			assertions: func(_ api.ThirdPartyAuthHelper, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "OIDC_PROVIDER_URL")
			},
		},
		{
			name: "OIDC_CLIENT_ID required but not set",
			setup: func() {
				t.Setenv("OIDC_PROVIDER_URL", server.URL)
			},
			assertions: func(_ api.ThirdPartyAuthHelper, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "OIDC_CLIENT_ID")
			},
		},
		{
			name: "OIDC_CLIENT_SECRET required but not set",
			setup: func() {
				t.Setenv("OIDC_CLIENT_ID", "hal9000")
			},
			assertions: func(_ api.ThirdPartyAuthHelper, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "OIDC_CLIENT_SECRET")
			},
		},
		{
			name: "OIDC_REDIRECT_URL_BASE required but not set",
			setup: func() {
				t.Setenv("OIDC_CLIENT_SECRET", "hello, dave")
			},
			assertions: func(_ api.ThirdPartyAuthHelper, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "OIDC_REDIRECT_URL_BASE")
			},
		},
		{
			name: "success getting OIDC-based ThirdPartyAuthHelper",
			setup: func() {
				t.Setenv("OIDC_REDIRECT_URL_BASE", "https://brigade.example.com")
			},
			assertions: func(helper api.ThirdPartyAuthHelper, err error) {
				require.NoError(t, err)
				require.Equal(
					t,
					"*oidc.thirdPartyAuthHelper",
					reflect.TypeOf(helper).String(),
				)
			},
		},
		{
			name: "GITHUB_CLIENT_ID required but not set",
			setup: func() {
				t.Setenv("THIRD_PARTY_AUTH_STRATEGY", thirdPartyAuthStrategyGitHub)
				t.Setenv("GITHUB_AUTH_ENABLED", "true")
			},
			assertions: func(_ api.ThirdPartyAuthHelper, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "GITHUB_CLIENT_ID")
			},
		},
		{
			name: "GITHUB_CLIENT_SECRET required but not set",
			setup: func() {
				t.Setenv("GITHUB_CLIENT_ID", "foo")
			},
			assertions: func(_ api.ThirdPartyAuthHelper, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "GITHUB_CLIENT_SECRET")
			},
		},
		{
			name: "success getting github-based ThirdPartyAuthHelper",
			setup: func() {
				t.Setenv("GITHUB_CLIENT_SECRET", "bar")
			},
			assertions: func(helper api.ThirdPartyAuthHelper, err error) {
				require.NoError(t, err)
				require.Equal(
					t,
					"*github.thirdPartyAuthHelper",
					reflect.TypeOf(helper).String(),
				)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup()
			thirdPartyAuthHelper, err := thirdPartyAuthHelper(context.Background())
			testCase.assertions(thirdPartyAuthHelper, err)
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
				t.Setenv("ROOT_USER_ENABLED", "aw hell no")
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as a bool")
				require.Contains(t, err.Error(), "ROOT_USER_ENABLED")
			},
		},
		{
			name: "ROOT_USER_SESSION_TTL not parsable as duration",
			setup: func() {
				t.Setenv("ROOT_USER_ENABLED", "true")
				t.Setenv("ROOT_USER_SESSION_TTL", "in like an hour")
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as a duration")
				require.Contains(t, err.Error(), "ROOT_USER_SESSION_TTL")
			},
		},
		{
			name: "ROOT_USER_PASSWORD required but not set",
			setup: func() {
				t.Setenv("ROOT_USER_SESSION_TTL", "1h")
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "ROOT_USER_PASSWORD")
			},
		},
		{
			name: "USER_SESSION_TTL not parsable as duration",
			setup: func() {
				t.Setenv("ROOT_USER_PASSWORD", "12345")
				t.Setenv("USER_SESSION_TTL", "in like a day")
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as a duration")
				require.Contains(t, err.Error(), "USER_SESSION_TTL")
			},
		},
		{
			name: "success",
			setup: func() {
				t.Setenv("USER_SESSION_TTL", "1h")
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup()
			_, err := sessionsServiceConfig()
			testCase.assertions(err)
		})
	}
}

func TestUsersServiceConfig(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func()
		assertions func(api.UsersServiceConfig)
	}{
		{
			name: "third party auth disabled",
			setup: func() {
				t.Setenv("THIRD_PARTY_AUTH_STRATEGY", thirdPartyAuthStrategyDisabled)
			},
			assertions: func(config api.UsersServiceConfig) {
				require.False(t, config.ThirdPartyAuthEnabled)
			},
		},
		{
			name: "third party auth enabled",
			setup: func() {
				t.Setenv("THIRD_PARTY_AUTH_STRATEGY", thirdPartyAuthStrategyGitHub)
			},
			assertions: func(config api.UsersServiceConfig) {
				require.True(t, config.ThirdPartyAuthEnabled)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup()
			testCase.assertions(usersServiceConfig())
		})
	}
}

func TestTokenAuthFilterConfig(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func()
		assertions func(rest.TokenAuthFilterConfig, error)
	}{
		{
			name: "ROOT_USER_ENABLED not parsable as bool",
			setup: func() {
				t.Setenv("ROOT_USER_ENABLED", "yuppers")
			},
			assertions: func(_ rest.TokenAuthFilterConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as a bool")
				require.Contains(t, err.Error(), "ROOT_USER_ENABLED")
			},
		},
		{
			name: "SCHEDULER_TOKEN not set",
			setup: func() {
				t.Setenv("ROOT_USER_ENABLED", "true")
			},
			assertions: func(_ rest.TokenAuthFilterConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "SCHEDULER_TOKEN")
			},
		},
		{
			name: "OBSERVER_TOKEN not set",
			setup: func() {
				t.Setenv("SCHEDULER_TOKEN", "foo")
			},
			assertions: func(_ rest.TokenAuthFilterConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "OBSERVER_TOKEN")
			},
		},
		{
			name: "success",
			setup: func() {
				t.Setenv("OBSERVER_TOKEN", "bar")
				t.Setenv("THIRD_PARTY_AUTH_STRATEGY", thirdPartyAuthStrategyGitHub)
			},
			assertions: func(config rest.TokenAuthFilterConfig, err error) {
				require.NoError(t, err)
				require.NotNil(t, config.FindUserFn)
				require.True(t, config.RootUserEnabled)
				require.True(t, config.ThirdPartyAuthEnabled)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup()
			config, err := tokenAuthFilterConfig(
				func(context.Context, string) (api.User, error) {
					return api.User{}, nil
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
				t.Setenv("API_SERVER_PORT", "foo")
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
				t.Setenv("API_SERVER_PORT", "8080")
				t.Setenv("TLS_ENABLED", "nope")
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
				t.Setenv("TLS_ENABLED", "true")
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
				t.Setenv("TLS_CERT_PATH", "/var/ssl/cert")
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
				t.Setenv("TLS_KEY_PATH", "/var/ssl/key")
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
