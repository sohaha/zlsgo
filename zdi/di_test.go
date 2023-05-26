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

	tt.Run("TestSetParent", func(tt *zlsgo.TestUtil) {
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

func TestMultiple(t *testing.T) {
	tt := zlsgo.NewTest(t)
	di := zdi.New()

	test1 := &testSt{Num: 1}
	test2 := &testSt{Num: 2}
	test3 := &testSt{Num: 3}

	di.Maps(test1, test2)
	di.Map(test3)

	_, err := di.Invoke(func(test *testSt) {
		t.Log(test)
	})
	tt.NoError(err)

	err = di.InvokeWithErrorOnly(func(test []*testSt) error {
		t.Log(test)
		return nil
	})
	tt.Log(err)
	tt.EqualTrue(err != nil)
}
