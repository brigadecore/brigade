package main

import (
	"context"
	"fmt"
	"log"
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
	dbConnectionStr := os.GetEnvVar("DATABASE_CONNECTION_STRING", "")
	dbName, err := os.GetRequiredEnvVar("DATABASE_NAME")
	if err != nil {
		return nil, err
	}
	var dbClientOpts *options.ClientOptions
	if dbConnectionStr != "" {
		dbClientOpts = options.Client().ApplyURI(dbConnectionStr)
	} else {
		var hosts string
		if hosts, err = os.GetRequiredEnvVar("DATABASE_HOSTS"); err != nil {
			return nil, err
		}
		var username string
		if username, err = os.GetRequiredEnvVar("DATABASE_USERNAME"); err != nil {
			return nil, err
		}
		var password string
		if password, err = os.GetRequiredEnvVar("DATABASE_PASSWORD"); err != nil {
			return nil, err
		}
		replicaSetName := os.GetEnvVar("DATABASE_REPLICA_SET", "")
		dbClientOpts = &options.ClientOptions{
			Hosts: strings.Split(hosts, ","),
			Auth: &options.Credential{
				AuthSource:  dbName,
				Username:    username,
				Password:    password,
				PasswordSet: true,
			},
		}
		if replicaSetName != "" {
			dbClientOpts.ReplicaSet = &replicaSetName
			dbClientOpts.WriteConcern = writeconcern.New(writeconcern.WMajority())
			dbClientOpts.ReadConcern = readconcern.Linearizable()
		}
	}
	connectCtx, connectCancel := context.WithTimeout(ctx, 10*time.Second)
	defer connectCancel()
	// This client's settings favor consistency over speed
	mongoClient, err := mongo.Connect(connectCtx, dbClientOpts)
	if err != nil {
		return nil, err
	}
	return mongoClient.Database(dbName), nil
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
	log.Println("AMQP_ADDRESS: ", config.Address)
	config.Username, err = os.GetRequiredEnvVar("AMQP_USERNAME")
	if err != nil {
		return config, err
	}
	log.Println("AMQP_USERNAME: ", config.Username)
	config.Password, err = os.GetRequiredEnvVar("AMQP_PASSWORD")
	return config, err
}

// substrateConfig returns a kubernetes.SubstrateConfig based on configuration
// obtained from environment variables.
func substrateConfig() (kubernetes.SubstrateConfig, error) {
	config := kubernetes.SubstrateConfig{}
	var err error
	config.BrigadeID, err = os.GetRequiredEnvVar("BRIGADE_ID")
	if err != nil {
		return config, err
	}
	log.Println("BRIGADE_ID: ", config.BrigadeID)
	config.APIAddress, err = os.GetRequiredEnvVar("API_ADDRESS")
	if err != nil {
		return config, err
	}
	log.Println("API_ADDRESS: ", config.APIAddress)
	config.GitInitializerImage, err =
		os.GetRequiredEnvVar("GIT_INITIALIZER_IMAGE")
	if err != nil {
		return config, err
	}
	log.Println("GIT_INITIALIZER_IMAGE: ", config.GitInitializerImage)
	gitInitializerImagePullPolicyStr, err :=
		os.GetRequiredEnvVar("GIT_INITIALIZER_IMAGE_PULL_POLICY")
	if err != nil {
		return config, err
	}
	log.Println("GIT_INITIALIZER_IMAGE_PULL_POLICY: ",
		config.GitInitializerImagePullPolicy)
	config.GitInitializerImagePullPolicy =
		api.ImagePullPolicy(gitInitializerImagePullPolicyStr)
	config.GitInitializerWindowsImage, err =
		os.GetRequiredEnvVar("GIT_INITIALIZER_WINDOWS_IMAGE")
	if err != nil {
		return config, err
	}
	log.Println("GIT_INITIALIZER_WINDOWS_IMAGE: ",
		config.GitInitializerWindowsImage)
	gitInitializerWindowsImagePullPolicyStr, err :=
		os.GetRequiredEnvVar("GIT_INITIALIZER_WINDOWS_IMAGE_PULL_POLICY")
	if err != nil {
		return config, err
	}
	log.Println("GIT_INITIALIZER_WINDOWS_IMAGE_PULL_POLICY: ",
		config.GitInitializerWindowsImagePullPolicy)
	config.GitInitializerWindowsImagePullPolicy =
		api.ImagePullPolicy(gitInitializerWindowsImagePullPolicyStr)
	config.DefaultWorkerImage, err = os.GetRequiredEnvVar("DEFAULT_WORKER_IMAGE")
	if err != nil {
		return config, err
	}
	log.Println("DEFAULT_WORKER_IMAGE: ", config.DefaultWorkerImage)
	defaultWorkerImagePullPolicyStr, err :=
		os.GetRequiredEnvVar("DEFAULT_WORKER_IMAGE_PULL_POLICY")
	if err != nil {
		return config, err
	}
	log.Println("DEFAULT_WORKER_IMAGE_PULL_POLICY: ",
		config.DefaultWorkerImagePullPolicy)
	config.DefaultWorkerImagePullPolicy =
		api.ImagePullPolicy(defaultWorkerImagePullPolicyStr)
	config.WorkspaceStorageClass, err =
		os.GetRequiredEnvVar("WORKSPACE_STORAGE_CLASS")
	if err != nil {
		return config, err
	}
	log.Println("WORKSPACE_STORAGE_CLASS: ", config.WorkspaceStorageClass)
	config.NodeSelectorKey = os.GetEnvVar("NODE_SELECTOR_KEY", "")
	config.NodeSelectorValue = os.GetEnvVar("NODE_SELECTOR_VALUE", "")
	config.TolerationKey = os.GetEnvVar("TOLERATION_KEY", "")
	config.TolerationValue = os.GetEnvVar("TOLERATION_VALUE", "")
	return config, nil
}

// thirdPartyAuthHelper returns an appropriate instance of
// api.ThirdPartyAuthHelper based on configuration obtained from environment
// variables.
func thirdPartyAuthHelper(
	ctx context.Context,
) (api.ThirdPartyAuthHelper, error) {
	thirdPartyAuthStrategy :=
		os.GetEnvVar("THIRD_PARTY_AUTH_STRATEGY", thirdPartyAuthStrategyDisabled)
	log.Println("THIRD_PARTY_AUTH_STRATEGY: ", thirdPartyAuthStrategy)
	switch api.ThirdPartyAuthStrategy(thirdPartyAuthStrategy) {
	case thirdPartyAuthStrategyOIDC:
		providerURL, err := os.GetRequiredEnvVar("OIDC_PROVIDER_URL")
		if err != nil {
			return nil, err
		}
		log.Println("OIDC_PROVIDER_URL: ", providerURL)
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
		log.Println("OIDC_REDIRECT_URL_BASE: ", redirectURLBase)
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
		log.Println("GITHUB_ALLOWED_ORGANIZATIONS: ", config.AllowedOrganizations)
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
	log.Println("ROOT_USER_ENABLED: ", config.RootUserEnabled)
	if config.RootUserSessionTTL, err =
		os.GetDurationFromEnvVar("ROOT_USER_SESSION_TTL", time.Hour); err != nil {
		return config, err
	}
	log.Println("ROOT_USER_SESSION_TTL: ", config.RootUserSessionTTL)
	if config.RootUserEnabled {
		if config.RootUserPassword, err =
			os.GetRequiredEnvVar("ROOT_USER_PASSWORD"); err != nil {
			return config, err
		}
	}
	thirdPartyAuthStrategy :=
		os.GetEnvVar("THIRD_PARTY_AUTH_STRATEGY", thirdPartyAuthStrategyDisabled)
	log.Println("THIRD_PARTY_AUTH_STRATEGY: ", thirdPartyAuthStrategy)
	config.ThirdPartyAuthEnabled =
		thirdPartyAuthStrategy != thirdPartyAuthStrategyDisabled
	if config.UserSessionTTL, err =
		os.GetDurationFromEnvVar("USER_SESSION_TTL", time.Hour); err != nil {
		return config, err
	}
	log.Println("USER_SESSION_TTL: ", config.UserSessionTTL)
	config.AdminUserIDs =
		os.GetStringSliceFromEnvVar("ADMIN_USER_IDS", []string{})
	log.Println("ADMIN_USER_IDS: ", config.AdminUserIDs)
	config.GrantReadOnInitialLogin, err =
		os.GetBoolFromEnvVar("GRANT_READ_ON_INITIAL_LOGIN", false)
	if err != nil {
		return config, err
	}
	log.Println("GRANT_READ_ON_INITIAL_LOGIN: ", config.GrantReadOnInitialLogin)
	return config, nil
}

// usersServiceConfig returns an api.UsersServiceConfig based on configuration
// obtained from environment variables. nolint: gocyclo
func usersServiceConfig() api.UsersServiceConfig {
	return api.UsersServiceConfig{
		ThirdPartyAuthEnabled: os.GetEnvVar(
			"THIRD_PARTY_AUTH_STRATEGY",
			thirdPartyAuthStrategyDisabled,
		) != thirdPartyAuthStrategyDisabled,
	}
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
	log.Println("ROOT_USER_ENABLED: ", config.RootUserEnabled)
	thirdPartyAuthStrategy :=
		os.GetEnvVar("THIRD_PARTY_AUTH_STRATEGY", thirdPartyAuthStrategyDisabled)
	log.Println("THIRD_PARTY_AUTH_STRATEGY", thirdPartyAuthStrategy)
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
	log.Println("API_SERVER_PORT: ", config.Port)
	config.TLSEnabled, err = os.GetBoolFromEnvVar("TLS_ENABLED", false)
	if err != nil {
		return config, err
	}
	log.Println("TLS_ENABLED: ", config.TLSEnabled)
	if config.TLSEnabled {
		config.TLSCertPath, err = os.GetRequiredEnvVar("TLS_CERT_PATH")
		if err != nil {
			return config, err
		}
		log.Println("TLS_CERT_PATH: ", config.TLSCertPath)
		config.TLSKeyPath, err = os.GetRequiredEnvVar("TLS_KEY_PATH")
		if err != nil {
			return config, err
		}
		log.Println("TLS_KEY_PATH: ", config.TLSKeyPath)
	}
	return config, nil
}
