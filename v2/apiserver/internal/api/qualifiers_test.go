package api

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestQualifiersMarshalBSONValue(t *testing.T) {
	// These keys are NOT in lexical order
	qualifiers := Qualifiers{
		"foo": "bar",
		"bat": "baz",
		"abc": "xyz",
	}
	_, bsonBytes, err := bson.MarshalValue(qualifiers)
	require.NoError(t, err)
	// Unmarshal into a generic bson.D and verify that the fields are now in
	// lexical order.
	d := bson.D{}
	err = bson.Unmarshal(bsonBytes, &d)
	require.NoError(t, err)
	require.Equal(
		t,
		// These keys ARE in lexical order
		bson.D{
			bson.E{
				Key:   "abc",
				Value: "xyz",
			},
			bson.E{
				Key:   "bat",
				Value: "baz",
			},
			bson.E{
				Key:   "foo",
				Value: "bar",
			},
		},
		d,
	)
}
