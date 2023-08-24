package zdi_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zreflect"
)

type testFastr func(v1 time.Time, v2 *testSt) string

func (f testFastr) Invoke(args []interface{}) ([]reflect.Value, error) {
	s := f(args[0].(time.Time), args[1].(*testSt))
	return []reflect.Value{zreflect.ValueOf(s)}, nil
}

type testFastRest func(v1 time.Time, v2 *testSt) string

func (f testFastRest) Invoke(args []interface{}) ([]interface{}, error) {
	s := f(args[0].(time.Time), args[1].(*testSt))
	return []interface{}{s}, nil
}

func TestFastInvoke(t *testing.T) {
	tt := zlsgo.NewTest(t)
	di := zdi.New()

	val := time.Now()
	override := di.Maps(val, &testSt{Msg: "The is FastInvoke"})
	tt.Equal(0, len(override))

	f := testFastr(func(v1 time.Time, v2 *testSt) string {
		t.Log(v1, v2)
		return "yes"
	})

	t.Log(zdi.IsPreInvoker(f))

	invoke, err := di.Invoke(f)
	t.Log(invoke, err)

	fr := testFastRest(func(v1 time.Time, v2 *testSt) string {
		t.Log(v1, v2)
		return "yes"
	})

	t.Log(zdi.IsPreInvoker(fr))

	invoke, err = di.Invoke(fr)
	t.Log(invoke, err)
}

func BenchmarkFast(b *testing.B) {
	di := zdi.New()
	now := time.Now()
	_ = di.Maps(now, &testSt{Msg: ""})
	f := testFastr(func(v1 time.Time, v2 *testSt) string {
		if v1 != now {
			b.Error("not equal")
			b.Fail()
		}
		return "yes"
	})
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := di.Invoke(f)
		if err != nil {
			b.Error("not equal")
			b.Fail()
		}
	}
}

func BenchmarkNoFast(b *testing.B) {
	di := zdi.New()
	now := time.Now()
	_ = di.Maps(now, &testSt{Msg: ""})
	f := func(v1 time.Time, v2 *testSt) string {
		if v1 != now {
			b.Error("not equal")
			b.Fail()
		}
		return "yes"
	}
	for i := 0; i < b.N; i++ {
		_, err := di.Invoke(f)
		if err != nil {
			b.Error("not equal")
			b.Fail()
		}
	}
}

func BenchmarkNoFast2(b *testing.B) {
	di := zdi.New()
	now := time.Now()
	_ = di.Maps(now, &testSt{Msg: ""})
	f := func(v1 time.Time, v2 *testSt) string {
		if v1 != now {
			b.Error("not equal")
			b.Fail()
		}
		return "yes"
	}
	for i := 0; i < b.N; i++ {
		_, err := di.Invoke(testFastRest(f))
		if err != nil {
			b.Error("not equal")
			b.Fail()
		}
	}
}
