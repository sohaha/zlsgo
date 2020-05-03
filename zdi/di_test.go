package zdi

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestDI(T *testing.T) {
	t := zlsgo.NewTest(T)
	di := New()
	di.Bind("test", func() {
		t.Log("test di")
	})
	t.EqualExit(true, di.Exist("test"))
	err := di.SoftMake("test", func() {
		t.Log("test di 2")
	})
	t.Log(err)
	_ = di.SoftMake("test3", func() {
		t.Log("test di 3")
	})
	fn := di.Make("test").(func())
	fn()

	di.Make("test3").(func())()

	di.Remove("test")
}
