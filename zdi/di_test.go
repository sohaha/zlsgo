package zdi_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/ztime"
)

type testSt struct {
	Msg string `di:""`
	Num int
}

func TestBase(t *testing.T) {
	tt := zlsgo.NewTest(t)
	di := zdi.New()

	now := ztime.Now()
	test1 := &testSt{Msg: now, Num: 666}
	override := di.Maps(test1, testSt2{Name: "main"})
	tt.Equal(0, len(override))

	tt.Run("TestSetParent", func(t *testing.T, tt *zlsgo.TestUtil) {
		ndi := zdi.New(di)
		ndi.Map(testSt2{Name: "Current"})
		_, err := ndi.Invoke(func(t2 testSt2, t1 *testSt) {
			tt.Equal("Current", t2.Name)
			tt.Equal(666, t1.Num)
			tt.Equal(now, t1.Msg)
			t.Log(t2, t1)
		})
		tt.NoError(err)
	})
}
