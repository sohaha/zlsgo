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

// GetUnexportedField Get unexported field, hazardous operation, please use with caution
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

// SetUnexportedField Set unexported field, hazardous operation, please use with caution
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
