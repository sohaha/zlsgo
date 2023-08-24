package zreflect

import (
	"reflect"
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestValue(t *testing.T) {
	tt := zlsgo.NewTest(t)

	val := ValueOf(Demo)
	zval := NewValue(Demo)
	gval := NewValue(val)
	zzval := NewValue(zval)

	tt.Equal(reflect.Struct, val.Kind())
	tt.Equal(reflect.Struct, zval.Native().Kind())
	tt.Equal(reflect.Struct, gval.Native().Kind())
	tt.Equal(reflect.Struct, zzval.Native().Kind())
	tt.Equal(reflect.Struct, zval.Type().Native().Kind())

	tt.Log(val.Interface())
	tt.Log(zval.Native().Interface())
}
