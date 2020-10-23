package mongodb

import (
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

func TestIsDuplicateKeyError(t *testing.T) {
	testCases := []struct {
		name                string
		err                 error
		isDuplicateKeyError bool
	}{
		{
			name:                "is not write exception",
			err:                 errors.New("some random error"),
			isDuplicateKeyError: false,
		},
		{
			name: "has wrong error code",
			err: mongo.WriteErrors{
				mongo.WriteError{
					Code: 42,
				},
			},
			isDuplicateKeyError: false,
		},
		{
			name: "is a legitimate duplicate key error",
			err: mongo.WriteErrors{
				mongo.WriteError{
					Code: 11000,
				},
			},
			isDuplicateKeyError: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

		})
	}
}
