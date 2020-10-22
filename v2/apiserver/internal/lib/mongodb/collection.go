package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection is an interface for the subset of *mongo.Collection functions that
// we actually use. Using this interface in our datastores, instead of using the
// *mongo.Collection type directly, allows for the possibility of utilizing a
// mock implementation for testing purposes. Adding only the subset of functions
// that we actually use limits the effort involved in creating such mocks.
type Collection interface {
	// CountDocuments returns the number of documents in the collection.
	CountDocuments(
		ctx context.Context,
		filter interface{},
		opts ...*options.CountOptions,
	) (int64, error)
	// DeleteMany executes a delete command to delete documents from the
	// collection.
	DeleteMany(
		ctx context.Context,
		filter interface{},
		opts ...*options.DeleteOptions,
	) (*mongo.DeleteResult, error)
	// DeleteOne executes a delete command to delete at most one document from the
	// collection.
	DeleteOne(
		ctx context.Context,
		filter interface{},
		opts ...*options.DeleteOptions,
	) (*mongo.DeleteResult, error)
	// Find executes a find command and returns a Cursor over the matching
	// documents in the collection.
	Find(
		ctx context.Context,
		filter interface{},
		opts ...*options.FindOptions,
	) (*mongo.Cursor, error)
	// FindOne executes a find command and returns a SingleResult for one document
	// in the collection.
	FindOne(
		ctx context.Context,
		filter interface{},
		opts ...*options.FindOneOptions,
	) *mongo.SingleResult
	// InsertOne executes an insert command to insert a single document into the
	// collection.
	InsertOne(
		ctx context.Context,
		document interface{},
		opts ...*options.InsertOneOptions,
	) (*mongo.InsertOneResult, error)
	// UpdateMany executes an update command to update documents in the
	// collection.
	UpdateMany(
		ctx context.Context,
		filter interface{},
		update interface{},
		opts ...*options.UpdateOptions,
	) (*mongo.UpdateResult, error)
	// UpdateOne executes an update command to update at most one document in the
	// collection.
	UpdateOne(
		ctx context.Context,
		filter interface{},
		update interface{},
		opts ...*options.UpdateOptions,
	) (*mongo.UpdateResult, error)
}
