//go:build go1.18
// +build go1.18

package zutil_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zutil"
)

func TestOptional(t *testing.T) {
	tt := zlsgo.NewTest(t)

	o := zutil.Optional(TestSt{Name: "test"})
	tt.Equal("test", o.Name)
	tt.Equal(0, o.I)

	o = zutil.Optional(TestSt{Name: "test2"}, func(o *TestSt) {
		o.I = 1
	}, func(ts *TestSt) {
		ts.I = ts.I + 1
	})
	tt.Equal("test2", o.Name)
	tt.Equal(2, o.I)

	o2 := zutil.Optional(&TestSt{Name: "test"}, func(ts **TestSt) {
		(*ts).I = 1
	})
	tt.Equal("test", o2.Name)
	tt.Equal(1, o2.I)
}
