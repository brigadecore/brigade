package core

import (
	"sort"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type Qualifiers map[string]string

func (q Qualifiers) MarshalBSONValue() (bsontype.Type, []byte, error) {
	keys := make([]string, len(q))
	var i int
	for k := range q {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	d := make(bson.D, len(q))
	for i, k := range keys {
		d[i] = bson.E{Key: k, Value: q[k]}
	}
	return bson.MarshalValue(d)
}
