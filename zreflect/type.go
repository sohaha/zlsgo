package zreflect

import (
	"reflect"
	"unsafe"
	_ "unsafe"
)

func TypeOf(v interface{}) reflect.Type {
	return toRType(NewType(v))
}

//go:linkname toRType reflect.toType
//go:noescape
func toRType(Type) reflect.Type

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

func (t *rtype) Native() reflect.Type {
	return toRType(t)
}

func rtypeToType(t reflect.Type) Type {
	return (Type)(((*Value)(unsafe.Pointer(&t))).ptr)
}

//go:linkname typeNumMethod reflect.(*rtype).NumMethod
//go:noescape
func typeNumMethod(Type) int

// NumMethod returns the number of exported methods in the type's method set.
func (t *rtype) NumMethod() int {
	return typeNumMethod(t)
}
