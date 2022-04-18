package zreflect

import (
	"reflect"
	"strings"
	"time"
)

const (
	ignoreTagValue = "-"
	nameConnector  = "::"
)

func GetStructTag(field reflect.StructField) (tagValue string) {
	tagValue = field.Tag.Get(Tag)
	if checkTagValidity(tagValue) {
		return tagValue
	}

	tagValue = field.Tag.Get("json")
	if checkTagValidity(tagValue) {
		return strings.Split(tagValue, ",")[0]
	}

	return field.Name
}

func checkTagValidity(tagValue string) bool {
	if tagValue != "" && tagValue != ignoreTagValue {
		return true
	}
	return false
}

func isTimeType(value reflect.Value) bool {
	if _, ok := value.Interface().(time.Time); ok {
		return true
	}
	return false
}

func Nonzero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() != complex(0, 0)
	case reflect.String:
		return v.String() != ""
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if Nonzero(GetInterfaceField(v, i)) {
				return true
			}
		}
		return false
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if Nonzero(v.Index(i)) {
				return true
			}
		}
		return false
	case reflect.Map, reflect.Interface, reflect.Slice, reflect.Ptr, reflect.Chan, reflect.Func:
		return !v.IsNil()
	case reflect.UnsafePointer:
		return v.Pointer() != 0
	}
	return true
}

func GetInterfaceField(v reflect.Value, i int) reflect.Value {
	val := v.Field(i)
	if val.Kind() == reflect.Interface && !val.IsNil() {
		val = val.Elem()
	}
	return val
}

func CanExpand(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map, reflect.Struct,
		reflect.Interface, reflect.Array, reflect.Slice,
		reflect.Ptr:
		return true
	}
	return false
}

func CanInline(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map:
		return !CanExpand(t.Elem())
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if CanExpand(t.Field(i).Type) {
				return false
			}
		}
		return true
	case reflect.Interface:
		return false
	case reflect.Array, reflect.Slice:
		return !CanExpand(t.Elem())
	case reflect.Ptr:
		return false
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return false
	}
	return true
}

func IsLabelType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Interface, reflect.Struct:
		return true
	}
	return false
}
