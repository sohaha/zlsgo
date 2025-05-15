package zreflect

import (
	"reflect"
	"unsafe"
	_ "unsafe"
)

// TypeOf returns the reflection Type of the value v.
// This is similar to reflect.TypeOf but with additional handling for zreflect.Type and zreflect.Value types.
func TypeOf(v interface{}) reflect.Type {
	return toRType(NewType(v))
}

//go:linkname toRType reflect.toType
//go:noescape
func toRType(Type) reflect.Type

// NewType creates a new Type from the given value.
// It handles various input types including Type, Value, reflect.Type, reflect.Value,
// or any other value, converting them to the internal Type representation.
func NewType(v interface{}) Type {
	switch t := v.(type) {
	case Type:
		return t
	case Value:
		return t.typ
	case reflect.Type:
		return rtypeToType(t)
	case reflect.Value:
		return (*Value)(unsafe.Pointer(&t)).typ
	default:
		return (*Value)(unsafe.Pointer(&v)).typ
	}
}

// Native converts the internal Type representation to the standard reflect.Type.
// This allows interoperability with the standard reflect package.
func (t *rtype) Native() reflect.Type {
	return toRType(t)
}

// rtypeToType converts a reflect.Type to the internal Type representation.
// This is an internal helper function used for type conversions.
func rtypeToType(t reflect.Type) Type {
	return (Type)(((*Value)(unsafe.Pointer(&t))).ptr)
}

//go:linkname typeNumMethod reflect.(*rtype).NumMethod
//go:noescape
func typeNumMethod(Type) int

// NumMethod returns the number of exported methods in the type's method set.
// This is equivalent to reflect.Type.NumMethod() but works on the internal Type representation.
func (t *rtype) NumMethod() int {
	return typeNumMethod(t)
}
