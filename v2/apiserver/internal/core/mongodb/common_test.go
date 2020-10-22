package mongodb

import (
	"context"
	"reflect"
	"unsafe"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/description"
)

type mockCollection struct {
	CountDocumentsFn func(
		ctx context.Context,
		filter interface{},
		opts ...*options.CountOptions,
	) (int64, error)

	DeleteManyFn func(
		ctx context.Context,
		filter interface{},
		opts ...*options.DeleteOptions,
	) (*mongo.DeleteResult, error)

	DeleteOneFn func(
		ctx context.Context,
		filter interface{},
		opts ...*options.DeleteOptions,
	) (*mongo.DeleteResult, error)

	FindFn func(
		ctx context.Context,
		filter interface{},
		opts ...*options.FindOptions,
	) (*mongo.Cursor, error)

	FindOneFn func(
		ctx context.Context,
		filter interface{},
		opts ...*options.FindOneOptions,
	) *mongo.SingleResult

	InsertOneFn func(
		ctx context.Context,
		document interface{},
		opts ...*options.InsertOneOptions,
	) (*mongo.InsertOneResult, error)

	UpdateManyFn func(
		ctx context.Context,
		filter interface{},
		update interface{},
		opts ...*options.UpdateOptions,
	) (*mongo.UpdateResult, error)

	UpdateOneFn func(
		ctx context.Context,
		filter interface{},
		update interface{},
		opts ...*options.UpdateOptions,
	) (*mongo.UpdateResult, error)
}

func (m *mockCollection) CountDocuments(
	ctx context.Context,
	filter interface{},
	opts ...*options.CountOptions,
) (int64, error) {
	return m.CountDocumentsFn(ctx, filter, opts...)
}

func (m *mockCollection) DeleteMany(
	ctx context.Context,
	filter interface{},
	opts ...*options.DeleteOptions,
) (*mongo.DeleteResult, error) {
	return m.DeleteManyFn(ctx, filter, opts...)
}

func (m *mockCollection) DeleteOne(
	ctx context.Context,
	filter interface{},
	opts ...*options.DeleteOptions,
) (*mongo.DeleteResult, error) {
	return m.DeleteOneFn(ctx, filter, opts...)
}

func (m *mockCollection) Find(
	ctx context.Context,
	filter interface{},
	opts ...*options.FindOptions,
) (*mongo.Cursor, error) {
	return m.FindFn(ctx, filter, opts...)
}

func (m *mockCollection) FindOne(
	ctx context.Context,
	filter interface{},
	opts ...*options.FindOneOptions,
) *mongo.SingleResult {
	return m.FindOneFn(ctx, filter, opts...)
}

func (m *mockCollection) InsertOne(
	ctx context.Context,
	document interface{},
	opts ...*options.InsertOneOptions,
) (*mongo.InsertOneResult, error) {
	return m.InsertOneFn(ctx, document, opts...)
}

func (m *mockCollection) UpdateMany(
	ctx context.Context,
	filter interface{},
	update interface{},
	opts ...*options.UpdateOptions,
) (*mongo.UpdateResult, error) {
	return m.UpdateManyFn(ctx, filter, update, opts...)
}

func (m *mockCollection) UpdateOne(
	ctx context.Context,
	filter interface{},
	update interface{},
	opts ...*options.UpdateOptions,
) (*mongo.UpdateResult, error) {
	return m.UpdateOneFn(ctx, filter, update, opts...)
}

var mockWriteException = mongo.WriteException{
	WriteErrors: mongo.WriteErrors{
		mongo.WriteError{
			Code: 11000,
		},
	},
}

func mockSingleResult(obj interface{}) (*mongo.SingleResult, error) {
	if err, ok := obj.(error); ok {
		res := &mongo.SingleResult{}
		setUnexportedField(res, "err", err)
		return res, nil
	}

	cursor, err := mockCursor(obj)
	if err != nil {
		return nil, err
	}

	// Build the codec registry so that the data can actually be decoded
	registryBuilder := bsoncodec.NewRegistryBuilder()
	defaultValueDecoders := bsoncodec.DefaultValueDecoders{}
	defaultValueDecoders.RegisterDefaultDecoders(registryBuilder)
	registry := registryBuilder.Build()

	res := &mongo.SingleResult{}
	setUnexportedField(res, "reg", registry)
	setUnexportedField(res, "cur", cursor)

	return res, nil
}

func mockCursor(objs ...interface{}) (*mongo.Cursor, error) {
	if objs == nil {
		objs = []interface{}{}
	}
	docBytes, err := bson.Marshal(
		bson.M{
			"ok": 1,
			"cursor": bson.M{
				// id 0 indicates this is the last chunk of data and prevents going back
				// to the server for more, which we defintely don't want since there is
				// no server. :)
				"id":         int64(0),
				"firstBatch": objs,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	cursorResponse, err :=
		driver.NewCursorResponse(docBytes, nil, description.Server{})
	if err != nil {
		return nil, err
	}
	batchCursor, err :=
		driver.NewBatchCursor(cursorResponse, nil, nil, driver.CursorOptions{})
	if err != nil {
		return nil, err
	}

	// Build the codec registry so that the data can actually be decoded
	registryBuilder := bsoncodec.NewRegistryBuilder()
	defaultValueDecoders := bsoncodec.DefaultValueDecoders{}
	defaultValueDecoders.RegisterDefaultDecoders(registryBuilder)
	registry := registryBuilder.Build()

	cursor := &mongo.Cursor{}
	setUnexportedField(cursor, "bc", batchCursor)
	setUnexportedField(cursor, "registry", registry)

	return cursor, nil
}

func setUnexportedField(
	objPtr interface{},
	fieldName string,
	fieldValue interface{},
) {
	field := reflect.ValueOf(objPtr).Elem().FieldByName(fieldName)
	reflect.NewAt(
		field.Type(),
		unsafe.Pointer(field.UnsafeAddr()),
	).Elem().Set(reflect.ValueOf(fieldValue))
}
