package zreflect

import (
	"reflect"
	"unsafe"
)

// ValueOf returns a reflect.Value for the specified interface{}.
// This is similar to reflect.ValueOf but with additional handling for nil values.
func ValueOf(v interface{}) reflect.Value {
	if v == nil {
		return reflect.Value{}
	}
	valueLayout := (*Value)(unsafe.Pointer(&v))
	value := Value{}
	value.typ = valueLayout.typ
	value.ptr = valueLayout.ptr
	f := flag(toRType(value.typ).Kind())
	if ifaceIndir(value.typ) {
		f |= 1 << 7
	}
	value.flag = f
	return value.Native()
}

// NewValue creates a new Value from the given interface{}.
// It handles various input types including Value, reflect.Value, or any other value,
// converting them to the internal Value representation.
func NewValue(v interface{}) Value {
	switch vv := v.(type) {
	case Value:
		return vv
	case reflect.Value:
		return *(*Value)(unsafe.Pointer(&vv))
	default:
		value := Value{}
		if v == nil {
			return value
		}
		valueLayout := (*Value)(unsafe.Pointer(&v))
		value.typ = valueLayout.typ
		value.ptr = valueLayout.ptr
		f := flag(toRType(value.typ).Kind())
		if ifaceIndir(value.typ) {
			f |= 1 << 7
		}
		value.flag = f
		return value
	}
}

// Native converts the internal Value representation to the standard reflect.Value.
// This allows interoperability with the standard reflect package.
func (v Value) Native() reflect.Value {
	return *(*reflect.Value)(unsafe.Pointer(&v))
}

// Type returns the Type of the value.
// This is equivalent to reflect.Value.Type() but returns the internal Type representation.
func (v Value) Type() Type {
	return v.typ
}
