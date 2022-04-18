package zreflect_test

import (
	"reflect"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zreflect"
)

func TestNonzero(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	v := reflect.ValueOf(0)
	t.EqualTrue(!zreflect.Nonzero(v))

	v = reflect.ValueOf(10.00)
	t.EqualTrue(zreflect.Nonzero(v))

	v = reflect.ValueOf(false)
	t.EqualTrue(!zreflect.Nonzero(v))
}

func TestCanInline(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	v := reflect.TypeOf(0)
	t.EqualTrue(zreflect.CanInline(v))

	v = reflect.TypeOf(&TestSt{})
	t.EqualTrue(!zreflect.CanInline(v))

	v = reflect.TypeOf(TestSt{Name: "yes"})
	t.EqualTrue(zreflect.CanInline(v))

	v = reflect.TypeOf(map[string]interface{}{"d": 10, "a": "zz"})
	t.EqualTrue(!zreflect.CanInline(v))

	v = reflect.TypeOf([...]int{10, 256})
	t.EqualTrue(zreflect.CanInline(v))

	v = reflect.TypeOf(func() {})
	t.EqualTrue(!zreflect.CanInline(v))

}
