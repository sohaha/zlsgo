package zdi_test

import (
	"github.com/sohaha/zlsgo/ztype"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/ztime"
)

func TestProvide(t *testing.T) {
	tt := zlsgo.NewTest(t)
	di := zdi.New()

	val := ztime.Now()

	di.Provide(func() *testSt {
		tt.Log("init testSt")
		return &testSt{Msg: val, Num: 666}
	})

	di.Provide(func(ts *testSt) *testSt2 {
		tt.Log("init testSt2")
		return &testSt2{Name: "2->" + ts.Msg}
	})

	override := di.Provide(func() time.Time {
		tt.Log("init time")
		return time.Now()
	})
	tt.Log(override)

	// overwrite the previous time.Time, *testSt
	override = di.Provide(func() (time.Time, *testSt) {
		tt.Log("init time2")
		return time.Now(), &testSt{Msg: val, Num: 999}
	})
	tt.Log(override)

	_, err := di.Invoke(func(t2 *testSt2, t1 *testSt, now time.Time) {
		tt.Log(t2.Name, t1.Num, now)
		tt.Equal(val, t1.Msg)
	})
	tt.NoError(err)

	di.Provide(func() *ztype.Type {
		tt.Log("test panic")
		panic("panic")
		return nil
	})

	_, err = di.Invoke(func(typ *ztype.Type) {
		tt.Log(typ)
	})
	tt.EqualTrue(err != nil)
	t.Log(err)
}
