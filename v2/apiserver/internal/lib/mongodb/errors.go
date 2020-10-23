package mongodb

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// IsDuplicateKeyError returns a bool indicating whether the provided error is a
// MongoDB duplicate key error.
func IsDuplicateKeyError(err error) bool {
	if writeException, ok := err.(mongo.WriteException); ok {
		return len(writeException.WriteErrors) == 1 &&
			writeException.WriteErrors[0].Code == 11000
	}
	return false
}
