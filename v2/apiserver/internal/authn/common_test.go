package authn

import (
	"reflect"
	"unsafe"
)

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
