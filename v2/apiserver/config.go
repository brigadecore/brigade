package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/brigadecore/brigade-foundations/crypto"
	"github.com/brigadecore/brigade-foundations/os"
	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	"github.com/brigadecore/brigade/v2/apiserver/internal/api/github"
	"github.com/brigadecore/brigade/v2/apiserver/internal/api/kubernetes"
	myOIDC "github.com/brigadecore/brigade/v2/apiserver/internal/api/oidc"
	"github.com/brigadecore/brigade/v2/apiserver/internal/api/rest"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue/amqp"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"golang.org/x/oauth2"
)

const (
	thirdPartyAuthStrategyDisabled = "disabled"
	thirdPartyAuthStrategyOIDC     = "oidc"
	thirdPartyAuthStrategyGitHub   = "github"
)

// databaseConnection returns a *mongo.Database connection based on
// configuration obtained from environment variables.
func databaseConnection(ctx context.Context) (*mongo.Database, error) {
	hosts, err := os.GetRequiredEnvVar("DATABASE_HOSTS")
	if err != nil {
		return nil, err
	}
	username, err := os.GetRequiredEnvVar("DATABASE_USERNAME")
	if err != nil {
		return nil, err
	}
	password, err := os.GetRequiredEnvVar("DATABASE_PASSWORD")
	if err != nil {
		return nil, err
	}
	replicaSetName := os.GetEnvVar("DATABASE_REPLICA_SET", "")
	name, err := os.GetRequiredEnvVar("DATABASE_NAME")
	if err != nil {
		return nil, err
	}

	opts := &options.ClientOptions{
		Hosts: strings.Split(hosts, ","),
		Auth: &options.Credential{
			AuthSource:  name,
			Username:    username,
			Password:    password,
			PasswordSet: true,
		},
	}
	if replicaSetName != "" {
		opts.ReplicaSet = &replicaSetName
		opts.WriteConcern = writeconcern.New(writeconcern.WMajority())
		opts.ReadConcern = readconcern.Linearizable()
	}

	connectCtx, connectCancel := context.WithTimeout(ctx, 10*time.Second)
	defer connectCancel()
	// This client's settings favor consistency over speed
	var mongoClient *mongo.Client
	mongoClient, err = mongo.Connect(connectCtx, opts)
	if err != nil {
		return nil, err
	}
	return mongoClient.Database(name), nil
}

// writerFactoryConfig returns an amqp.WriterFactoryConfig based on
// configuration obtained from environment variables.
func writerFactoryConfig() (amqp.WriterFactoryConfig, error) {
	config := amqp.WriterFactoryConfig{}
	var err error
	config.Address, err = os.GetRequiredEnvVar("AMQP_ADDRESS")
	if err != nil {
		return config, err
	}
	config.Username, err = os.GetRequiredEnvVar("AMQP_USERNAME")
	if err != nil {
		return config, err
	}
	config.Password, err = os.GetRequiredEnvVar("AMQP_PASSWORD")
	return config, err
}

// substrateConfig returns a kubernetes.SubstrateConfig based on configuration
// obtained from environment variables.
func substrateConfig() (kubernetes.SubstrateConfig, error) {
	config := kubernetes.SubstrateConfig{}
	var err error
	config.APIAddress, err = os.GetRequiredEnvVar("API_ADDRESS")
	if err != nil {
		return config, err
	}
	config.GitInitializerImage, err =
		os.GetRequiredEnvVar("GIT_INITIALIZER_IMAGE")
	if err != nil {
		return config, err
	}
	gitInitializerImagePullPolicyStr, err :=
		os.GetRequiredEnvVar("GIT_INITIALIZER_IMAGE_PULL_POLICY")
	if err != nil {
		return config, err
	}
	config.GitInitializerImagePullPolicy =
		api.ImagePullPolicy(gitInitializerImagePullPolicyStr)
	config.DefaultWorkerImage, err = os.GetRequiredEnvVar("DEFAULT_WORKER_IMAGE")
	if err != nil {
		return config, err
	}
	defaultWorkerImagePullPolicyStr, err :=
		os.GetRequiredEnvVar("DEFAULT_WORKER_IMAGE_PULL_POLICY")
	if err != nil {
		return config, err
	}
	config.DefaultWorkerImagePullPolicy =
		api.ImagePullPolicy(defaultWorkerImagePullPolicyStr)
	config.WorkspaceStorageClass, err =
		os.GetRequiredEnvVar("WORKSPACE_STORAGE_CLASS")
	return config, err
}

// thirdPartyAuthHelper returns an appropriate instance of
// api.ThirdPartyAuthHelper based on configuration obtained from environment
// variables.
func thirdPartyAuthHelper(
	ctx context.Context,
) (api.ThirdPartyAuthHelper, error) {
	thirdPartyAuthStrategy :=
		os.GetEnvVar("THIRD_PARTY_AUTH_STRATEGY", thirdPartyAuthStrategyDisabled)
	switch api.ThirdPartyAuthStrategy(thirdPartyAuthStrategy) {
	case thirdPartyAuthStrategyOIDC:
		providerURL, err := os.GetRequiredEnvVar("OIDC_PROVIDER_URL")
		if err != nil {
			return nil, err
		}
		provider, err := oidc.NewProvider(ctx, providerURL)
		if err != nil {
			return nil, err
		}
		clientID, err := os.GetRequiredEnvVar("OIDC_CLIENT_ID")
		if err != nil {
			return nil, err
		}
		clientSecret, err := os.GetRequiredEnvVar("OIDC_CLIENT_SECRET")
		if err != nil {
			return nil, err
		}
		redirectURLBase, err := os.GetRequiredEnvVar("OIDC_REDIRECT_URL_BASE")
		if err != nil {
			return nil, err
		}
		return myOIDC.NewThirdPartyAuthHelper(
			&oauth2.Config{
				Endpoint:     provider.Endpoint(),
				ClientID:     clientID,
				ClientSecret: clientSecret,
				RedirectURL:  fmt.Sprintf("%s/%s", redirectURLBase, "v2/session/auth"),
				Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
			},
			provider.Verifier(
				&oidc.Config{
					ClientID: clientID,
				},
			),
		), nil
	case thirdPartyAuthStrategyGitHub:
		config := github.ThirdPartyAuthHelperConfig{}
		var err error
		if config.ClientID, err =
			os.GetRequiredEnvVar("GITHUB_CLIENT_ID"); err != nil {
			return nil, err
		}
		if config.ClientSecret, err =
			os.GetRequiredEnvVar("GITHUB_CLIENT_SECRET"); err != nil {
			return nil, err
		}
		config.AllowedOrganizations =
			os.GetStringSliceFromEnvVar("GITHUB_ALLOWED_ORGANIZATIONS", []string{})
		return github.NewThirdPartyAuthHelper(config), nil
	case thirdPartyAuthStrategyDisabled:
		return nil, nil
	default:
		return nil, errors.Errorf(
			"unrecognized THIRD_PARTY_AUTH_STRATEGY %q",
			thirdPartyAuthStrategy,
		)
	}
}

// sessionsServiceConfig returns an api.SessionsServiceConfig based on
// configuration obtained from environment variables.
// nolint: gocyclo
func sessionsServiceConfig() (api.SessionsServiceConfig, error) {
	config := api.SessionsServiceConfig{}
	var err error
	if config.RootUserEnabled, err =
		os.GetBoolFromEnvVar("ROOT_USER_ENABLED", false); err != nil {
		return config, err
	}
	if config.RootUserSessionTTL, err = os.GetDurationFromEnvVar(
		"ROOT_USER_SESSION_TTL",
		time.Hour,
	); err != nil {
		return config, err
	}
	if config.RootUserEnabled {
		if config.RootUserPassword, err =
			os.GetRequiredEnvVar("ROOT_USER_PASSWORD"); err != nil {
			return config, err
		}
	}
	thirdPartyAuthStrategy :=
		os.GetEnvVar("THIRD_PARTY_AUTH_STRATEGY", thirdPartyAuthStrategyDisabled)
	config.ThirdPartyAuthEnabled = thirdPartyAuthStrategy != "disabled"
	config.UserSessionTTL, err = os.GetDurationFromEnvVar(
		"USER_SESSION_TTL",
		time.Hour,
	)
	config.AdminUserIDs =
		os.GetStringSliceFromEnvVar("ADMIN_USER_IDS", []string{})
	return config, err
}

// tokenAuthFilterConfig returns an api.TokenAuthFilterConfig based on
// configuration obtained from environment variables.
func tokenAuthFilterConfig(
	findUserFn func(ctx context.Context, id string) (api.User, error),
) (rest.TokenAuthFilterConfig, error) {
	config := rest.TokenAuthFilterConfig{
		FindUserFn: findUserFn,
	}
	var err error
	if config.RootUserEnabled, err =
		os.GetBoolFromEnvVar("ROOT_USER_ENABLED", false); err != nil {
		return config, err
	}
	thirdPartyAuthStrategy :=
		os.GetEnvVar("THIRD_PARTY_AUTH_STRATEGY", thirdPartyAuthStrategyDisabled)
	config.ThirdPartyAuthEnabled = thirdPartyAuthStrategy != "disabled"
	schedulerToken, err := os.GetRequiredEnvVar("SCHEDULER_TOKEN")
	if err != nil {
		return config, err
	}
	config.HashedSchedulerToken = crypto.Hash("", schedulerToken)
	observerToken, err := os.GetRequiredEnvVar("OBSERVER_TOKEN")
	if err != nil {
		return config, err
	}
	config.HashedObserverToken = crypto.Hash("", observerToken)
	return config, nil
}

// serverConfig returns a restmachinery.ServerConfig based on configuration
// obtained from environment variables.
func serverConfig() (restmachinery.ServerConfig, error) {
	config := restmachinery.ServerConfig{}
	var err error
	config.Port, err = os.GetIntFromEnvVar("API_SERVER_PORT", 8080)
	if err != nil {
		return config, err
	}
	config.TLSEnabled, err = os.GetBoolFromEnvVar("TLS_ENABLED", false)
	if err != nil {
		return config, err
	}
	if config.TLSEnabled {
		config.TLSCertPath, err = os.GetRequiredEnvVar("TLS_CERT_PATH")
		if err != nil {
			return config, err
		}
		config.TLSKeyPath, err = os.GetRequiredEnvVar("TLS_KEY_PATH")
		if err != nil {
			return config, err
		}
	}
	return config, nil
}
