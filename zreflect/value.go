package zreflect

import (
	"reflect"
	"unsafe"
)

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

func (v Value) Native() reflect.Value {
	return *(*reflect.Value)(unsafe.Pointer(&v))
}

func (v Value) Type() Type {
	return v.typ
}
