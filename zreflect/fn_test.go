package zreflect_test

import (
	"reflect"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zreflect"
)

func TestTyper_ForEachMethod(t *testing.T) {
	demo := DemoSt{Name: "ForEachMethod"}
	tt := zlsgo.NewTest(t)
	tp, _ := zreflect.ValueOf(&demo)

	typ, err := zreflect.NewVal(tp)
	tt.ErrorNil(err)

	err = typ.ForEachMethod(func(index int, method reflect.Method, value reflect.Value) error {
		fn, ok := value.Interface().(func())
		if ok {
			fn()
		}
		return nil
	})
	tt.ErrorNil(err)
}
