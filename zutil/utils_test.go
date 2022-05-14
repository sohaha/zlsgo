package zutil_test

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

func TestWithRunContext(t *testing.T) {
	tt := zlsgo.NewTest(t)
	duration, mem := zutil.WithRunContext(func() {
		var b = zstring.Buffer()
		size := 110000
		count := math.Ceil(float64(size) / 1000)
		count64 := int64(count)
		var i int64
		var length int
		for i = 0; i < count64; i++ {
			if i == (count64 - 1) {
				length = int(int64(size) - (i)*1000)
			} else {
				length = 1000
			}
			b.WriteString(strings.Repeat("A", length))
		}
		_ = b.String()
	})

	tt.EqualExit(true, mem > 60000)
	t.Log(duration)
	t.Log(zfile.SizeFormat(mem))
}

func TestIfVal(T *testing.T) {
	t := zlsgo.NewTest(T)
	i := zutil.IfVal(true, 1, 2)
	t.EqualExit(1, i)
	i = zutil.IfVal(false, 1, 2)
	t.EqualExit(2, i)
}

func TestTryCatch(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	errMsg := errors.New("test error")
	err := zutil.TryCatch(func() error {
		return errMsg
	})
	tt.Log(err)
	t.EqualTrue(err != nil)
	t.Equal(errMsg, err)

	err = zutil.TryCatch(func() error {
		panic(123)
	})
	tt.Log(err)
	t.EqualTrue(err != nil)
	t.Equal(errors.New("123"), err)
}

func TestTryError(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	defer func() {
		if message := recover(); message != nil {
			if e, ok := message.(error); ok {
				t.EqualExit("test", e.Error())
			}
		}
	}()

	zutil.Try(func() {
		panic("test")
	}, func(e interface{}) {
		if err, ok := e.(error); ok {
			t.EqualExit("test", err.Error())
		}
	}, func() {
		t.Log("TestTryError ok")
	})

	zutil.Try(func() {
		zutil.CheckErr(errors.New("test"))
	}, func(e interface{}) {
		if err, ok := e.(error); ok {
			t.EqualExit("test", err.Error())
		}
	})

	zutil.Try(func() {
		panic(t)
	}, func(e interface{}) {
		if err, ok := e.(error); ok {
			t.EqualExit("test", err.Error())
		}
	})

	zutil.Try(func() {
		panic("test")
	}, nil)
}

func TestUtil(t *testing.T) {
	_, _ = zutil.GetParentProcessName()
	zutil.IsDoubleClickStartUp()
}

func TestMaximizeOpenFileLimit(t *testing.T) {
	l, err := zutil.MaxRlimit()
	t.Log(l, err)
}
