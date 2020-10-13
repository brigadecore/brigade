package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/brigadecore/brigade/v2/internal/os"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// Database returns a *mongo.Database connection based on configuration obtained
// from environment variables.
func Database(ctx context.Context) (*mongo.Database, error) {
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
