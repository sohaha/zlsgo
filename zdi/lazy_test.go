package zdi_test

import (
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
		t.Log("init testSt")
		return &testSt{Msg: val, Num: 666}
	})

	di.Provide(func(tt *testSt) *testSt2 {
		t.Log("init testSt2")
		return &testSt2{Name: "2->" + tt.Msg}
	})

	override := di.Provide(func() time.Time {
		t.Log("init time")
		return time.Now()
	})

	// overwrite the previous time.Time, *testSt
	override = di.Provide(func() (time.Time, *testSt) {
		t.Log("init time2")
		return time.Now(), &testSt{Msg: val, Num: 999}
	})
	t.Log(override)

	_, err := di.Invoke(func(t2 *testSt2, t1 *testSt, now time.Time) {
		t.Log(t2.Name, t1.Num, now)
		tt.Equal(val, t1.Msg)
	})
	tt.NoError(err)
}
