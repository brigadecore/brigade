package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/core/kubernetes"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/crypto"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue/amqp"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	sysAuthn "github.com/brigadecore/brigade/v2/apiserver/internal/system/authn"
	"github.com/brigadecore/brigade/v2/internal/os"
	"github.com/coreos/go-oidc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"golang.org/x/oauth2"
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
		core.ImagePullPolicy(gitInitializerImagePullPolicyStr)
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
		core.ImagePullPolicy(defaultWorkerImagePullPolicyStr)
	config.WorkspaceStorageClass, err =
		os.GetRequiredEnvVar("WORKSPACE_STORAGE_CLASS")
	return config, err
}

// sessionsServiceConfig returns an authn.SessionsServiceConfig based on
// configuration obtained from environment variables.
func sessionsServiceConfig(
	ctx context.Context,
) (authn.SessionsServiceConfig, error) {
	config := authn.SessionsServiceConfig{}
	var err error
	config.RootUserEnabled, err = os.GetBoolFromEnvVar("ROOT_USER_ENABLED", false)
	if err != nil {
		return config, err
	}
	config.OpenIDConnectEnabled, err =
		os.GetBoolFromEnvVar("OIDC_ENABLED", false)
	if err != nil {
		return config, err
	}
	if config.RootUserEnabled {
		if config.RootUserSessionTTL, err = os.GetDurationFromEnvVar(
			"ROOT_USER_SESSION_TTL",
			time.Hour,
		); err != nil {
			return config, err
		}
		config.RootUserPassword, err = os.GetRequiredEnvVar("ROOT_USER_PASSWORD")
		if err != nil {
			return config, err
		}
	}
	if config.OpenIDConnectEnabled {
		if config.UserSessionTTL, err = os.GetDurationFromEnvVar(
			"OIDC_USER_SESSION_TTL",
			time.Hour,
		); err != nil {
			return config, err
		}
		providerURL, err := os.GetRequiredEnvVar("OIDC_PROVIDER_URL")
		if err != nil {
			return config, err
		}
		provider, err := oidc.NewProvider(ctx, providerURL)
		if err != nil {
			return config, err
		}
		clientID, err := os.GetRequiredEnvVar("OIDC_CLIENT_ID")
		if err != nil {
			return config, err
		}
		clientSecret, err := os.GetRequiredEnvVar("OIDC_CLIENT_SECRET")
		if err != nil {
			return config, err
		}
		redirectURLBase, err := os.GetRequiredEnvVar("OIDC_REDIRECT_URL_BASE")
		if err != nil {
			return config, err
		}
		config.OAuth2Helper = &oauth2.Config{
			Endpoint:     provider.Endpoint(),
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  fmt.Sprintf("%s/%s", redirectURLBase, "v2/session/auth"),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		}
		config.OpenIDConnectTokenVerifier = provider.Verifier(
			&oidc.Config{
				ClientID: clientID,
			},
		)
	}
	return config, nil
}

// tokenAuthFilterConfig returns an sysAuthn.TokenAuthFilterConfig based on
// configuration obtained from environment variables.
func tokenAuthFilterConfig(
	findUserFn func(ctx context.Context, id string) (authn.User, error),
) (sysAuthn.TokenAuthFilterConfig, error) {
	config := sysAuthn.TokenAuthFilterConfig{
		FindUserFn: findUserFn,
	}
	var err error
	config.RootUserEnabled, err =
		os.GetBoolFromEnvVar("ROOT_USER_ENABLED", false)
	if err != nil {
		return config, err
	}
	config.OpenIDConnectEnabled, err =
		os.GetBoolFromEnvVar("OIDC_ENABLED", false)
	if err != nil {
		return config, err
	}
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
