package zreflect

import (
	"reflect"
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestForEachMethod(t *testing.T) {
	tt := zlsgo.NewTest(t)
	v := ValueOf(Demo)

	err := ForEachMethod(v, func(index int, method reflect.Method, value reflect.Value) error {
		tt.Log(index, method.Name, value.Kind())
		return nil
	})
	tt.NoError(err)
}

func TestForEach(t *testing.T) {
	tt := zlsgo.NewTest(t)
	typ := TypeOf(Demo)

	err := ForEach(typ, func(parent []string, index int, tag string, field reflect.StructField) error {
		return nil
	})
	tt.NoError(err)
	err = ForEach(typ, func(parent []string, index int, tag string, field reflect.StructField) error {
		tt.Log(parent, index, tag, field.Name)
		return SkipChild
	})
	tt.NoError(err)
}

func TestForEachValue(t *testing.T) {
	tt := zlsgo.NewTest(t)
	v := ValueOf(Demo)

	err := ForEachValue(v, func(parent []string, index int, tag string, field reflect.StructField, value reflect.Value) error {
		tt.Log(parent, index, tag, field.Name, value.Interface())
		return nil
	})
	tt.NoError(err)
}
