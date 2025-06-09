package zdi_test

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/ztime"
)

func TestBind(t *testing.T) {
	tt := zlsgo.NewTest(t)
	di := zdi.New()

	test1 := testSt{Msg: ztime.Now(), Num: 666}
	o := di.Map(test1)
	tt.EqualNil(o)
	zlog.Dump(test1)

	var test2 testSt
	override := di.Resolve(&test2)

	tt.Equal(test1, test2)
	zlog.Dump(override, test2)

	var test3 testSt
	override = di.Resolve(test3)
	tt.EqualTrue(override != nil)
	zlog.Dump(override, test3)

	test5 := &testSt{Msg: ztime.Now(), Num: 777}
	o = di.Map(test5)
	tt.EqualNil(o)

	var test4 *testSt
	err := di.Resolve(test4)
	tt.EqualTrue(err != nil)
	zlog.Dump(err, test4)

	var test6 *testSt
	err = di.Resolve(&test6)
	tt.NoError(err)
	tt.Equal(test5, test6)
	zlog.Dump(err, test6)
}

func TestApply(t *testing.T) {
	tt := zlsgo.NewTest(t)
	di := zdi.New()

	val := time.Now().String()
	o := di.Map(val)
	tt.EqualNil(o)

	var v testSt
	err := di.Apply(&v)
	tt.EqualNil(err)
	tt.Logf("%+v %+v\n", val, v)

	var s string
	err = di.Apply(&s)
	tt.EqualNil(err)
	tt.Logf("%+v\n", s)
}

func TestResolve(t *testing.T) {
	tt := zlsgo.NewTest(t)
	di := zdi.New()

	val := &testSt{Msg: "TestResolve", Num: 2}
	o := di.Map(val)
	tt.EqualNil(o)

	var v *testSt
	err := di.Resolve(&v)
	tt.Logf("%+v %+v\n", val, v)
	tt.EqualNil(err)
	tt.Equal(val.Msg, v.Msg)
	tt.Equal(val.Num, v.Num)

	v = &testSt{}
	err = di.Resolve(&v)
	tt.Logf("%+v %+v\n", val, v)
	tt.EqualNil(err)
	tt.Equal(val.Msg, v.Msg)
	tt.Equal(val.Num, v.Num)
}
