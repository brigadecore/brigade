package main

import (
	"context"
	"fmt"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/core/kubernetes"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/crypto"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue/amqp"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery/authn"
	"github.com/brigadecore/brigade/v2/internal/os"
	"github.com/coreos/go-oidc"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"golang.org/x/oauth2"
)

// database returns a *mongo.Database connection based on configuration obtained
// from environment variables.
func database(ctx context.Context) (*mongo.Database, error) {
	username, err := os.GetRequiredEnvVar("DATABASE_USERNAME")
	if err != nil {
		return nil, err
	}
	password, err := os.GetRequiredEnvVar("DATABASE_PASSWORD")
	if err != nil {
		return nil, err
	}
	host, err := os.GetRequiredEnvVar("DATABASE_HOST")
	if err != nil {
		return nil, err
	}
	port, err := os.GetIntFromEnvVar("DATABASE_PORT", 27017)
	if err != nil {
		return nil, err
	}
	name, err := os.GetRequiredEnvVar("DATABASE_NAME")
	if err != nil {
		return nil, err
	}
	replicaSetName, err := os.GetRequiredEnvVar("DATABASE_REPLICA_SET")
	if err != nil {
		return nil, err
	}
	connectionString := fmt.Sprintf(
		"mongodb://%s:%s@%s:%d/%s?replicaSet=%s",
		username,
		password,
		host,
		port,
		name,
		replicaSetName,
	)
	connectCtx, connectCancel := context.WithTimeout(ctx, 10*time.Second)
	defer connectCancel()
	// This client's settings favor consistency over speed
	var mongoClient *mongo.Client
	mongoClient, err = mongo.Connect(
		connectCtx,
		options.Client().ApplyURI(connectionString).SetWriteConcern(
			writeconcern.New(writeconcern.WMajority()),
		).SetReadConcern(readconcern.Linearizable()),
	)
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
	return config, err
}

// sessionsServiceConfig returns an authx.SessionsServiceConfig based on
// configuration obtained from environment variables.
func sessionsServiceConfig(
	ctx context.Context,
) (authx.SessionsServiceConfig, error) {
	config := authx.SessionsServiceConfig{}
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
		config.RootUserPassword, err = os.GetRequiredEnvVar("ROOT_USER_PASSWORD")
		if err != nil {
			return config, err
		}
	}
	if config.OpenIDConnectEnabled {
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

// tokenAuthFilterConfig returns an authn.TokenAuthFilterConfig based on
// configuration obtained from environment variables.
func tokenAuthFilterConfig(
	findUserFn func(ctx context.Context, id string) (authx.User, error),
) (authn.TokenAuthFilterConfig, error) {
	config := authn.TokenAuthFilterConfig{
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
