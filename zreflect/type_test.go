package zreflect_test

import (
	"reflect"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zreflect"
)

func TestType(t *testing.T) {
	tt := zlsgo.NewTest(t)

	typ := zreflect.TypeOf(zreflect.Demo)
	ztyp := zreflect.NewType(zreflect.Demo)
	atyp := zreflect.NewValue(zreflect.Demo).Type()

	tt.Equal(reflect.Struct, typ.Kind())
	tt.Equal(reflect.Struct, ztyp.Native().Kind())
	tt.Equal(reflect.Struct, atyp.Native().Kind())

	tt.Equal(typ.NumMethod(), ztyp.NumMethod())

}
