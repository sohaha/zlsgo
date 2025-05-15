package zreflect

import (
	"errors"
	"reflect"
	"unsafe"

	"github.com/sohaha/zlsgo/zstring"
)

type (
	Type  = *rtype
	rtype struct{}
	flag  uintptr
	Value struct {
		typ Type
		ptr unsafe.Pointer
		flag
	}
)

// GetUnexportedField retrieves the value of an unexported field from a struct.
// This is a hazardous operation that bypasses Go's type safety and should be used with extreme caution.
//
// v is the reflect.Value of the struct containing the unexported field.
// field is the name of the unexported field to access.
//
// It returns the value of the unexported field, or an error if the field
// doesn't exist or cannot be accessed.
func GetUnexportedField(v reflect.Value, field string) (interface{}, error) {
	f, b, err := getField(v, field)
	if err != nil {
		return nil, err
	}

	if b {
		return f.Interface(), nil
	}

	return reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Interface(), nil
}

// SetUnexportedField sets the value of an unexported field in a struct.
// This is a hazardous operation that bypasses Go's type safety and should be used with extreme caution.
//
// v is the reflect.Value of the struct containing the unexported field.
// field is the name of the unexported field to modify.
// value is the new value to set for the field.
//
// It returns an error if the field doesn't exist, cannot be modified, or if the value type doesn't match.
func SetUnexportedField(v reflect.Value, field string, value interface{}) error {
	f, b, err := getField(v, field)
	if err != nil {
		return err
	}

	nv := reflect.ValueOf(value)
	kind := f.Kind()
	if kind != reflect.Interface && kind != nv.Kind() {
		return errors.New("value type not match")
	}

	if b {
		if !f.CanSet() {
			return errors.New("field can not set")
		}
		f.Set(nv)
		return nil
	}

	reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).
		Elem().
		Set(nv)

	return nil
}

// getField is an internal helper function that retrieves a field from a struct by name.
// It returns the field's reflect.Value, a boolean indicating if the field is exported,
// and an error if the field doesn't exist or cannot be accessed.
func getField(v reflect.Value, field string) (reflect.Value, bool, error) {
	ve := reflect.Indirect(v)
	if ve.Kind() != reflect.Struct {
		return reflect.Value{}, false, errors.New("value must be struct")
	}

	f := ve.FieldByName(field)
	if !f.IsValid() {
		return reflect.Value{}, false, errors.New("field not exists")
	}

	if zstring.IsUcfirst(field) {
		return f, true, nil
	}

	if v.Kind() != reflect.Ptr {
		return reflect.Value{}, false, errors.New("value must be ptr")
	}
	return f, false, nil
}
