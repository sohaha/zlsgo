package zdi_test

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/ztime"
)

type testSt2 struct {
	Name string
}

func (t *testSt2) String() string {
	return t.Name
}

type Itest interface {
	String() string
}

func TestInterfaceOf(t *testing.T) {
	tt := zlsgo.NewTest(t)
	di := zdi.New()

	val := time.Now()
	_ = di.Map(val)

	ok := "is ok"
	invoke, err := di.Invoke(func(s Itest) string {
		tt.Equal(s.String(), val.String())
		return ok
	})
	tt.NoError(err)
	tt.Equal(ok, invoke[0].String())

	val2 := &testSt2{Name: "val"}
	o := di.Map(val2, zdi.WithInterface((*Itest)(nil)))
	tt.EqualNil(o)

	invoke, err = di.Invoke(func(s Itest) string {
		tt.Equal(s.String(), "val")
		return ok
	})
	tt.NoError(err)
	tt.Equal(ok, invoke[0].String())

	_, err = di.Invoke(func(s Itest, t time.Time) string {
		tt.Equal(s.String(), "val")
		tt.Equal(t.String(), val.String())
		return ok
	})
	tt.NoError(err)

	invoke, err = di.Invoke(func(t time.Time) string {
		return ok
	})
	tt.NoError(err)
	tt.Equal(ok, invoke[0].String())
}

func TestInvoke(t *testing.T) {
	tt := zlsgo.NewTest(t)
	di := zdi.New()

	test := &testSt{Msg: ztime.Now(), Num: 666}
	_ = di.Map(test)

	ok := "is ok"
	invoke, err := di.Invoke(func(s *testSt) string {
		tt.Equal(s, test)
		return ok
	})
	tt.NoError(err)
	tt.Equal(ok, invoke[0].String())

	invoke, err = di.Invoke(func(s testSt) string {
		tt.Equal(s, test)
		return ok
	})
	tt.EqualTrue(err != nil)
	t.Log(err, invoke)

	test2 := testSt{Msg: ztime.Now(), Num: 666}
	_ = di.Map(test2)
	invoke, err = di.Invoke(func(s testSt, rs *testSt) int64 {
		tt.Equal(s, test2)
		tt.Equal(rs, test)
		return 18
	})
	tt.NoError(err)
	tt.Equal(int64(18), int64(invoke[0].Int()))
}
