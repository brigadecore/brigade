package core

import (
	"sort"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

// Labels is a map of key/value pairs utilized mutually by Events in describing
// themselves and by EventSubscriptions in describing Events of interest to a
// Project.
type Labels map[string]string

// MarshalBSONValue implements custom BSON marshaling for the Labels type.
// In MongoDB, when matching documents based on the value of a subdocument, the
// order of the fields in the subdocument matters. Therefore, we need to always
// normalize labels before storing them OR using them as a query parameter.
func (l Labels) MarshalBSONValue() (bsontype.Type, []byte, error) {
	keys := make([]string, len(l))
	var i int
	for k := range l {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	d := make(bson.D, len(l))
	for i, k := range keys {
		d[i] = bson.E{Key: k, Value: l[k]}
	}
	return bson.MarshalValue(d)
}
