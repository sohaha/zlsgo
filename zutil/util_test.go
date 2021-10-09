package zutil

import (
	"errors"
	"math"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zstring"
)

func TestWithLockContext(t *testing.T) {
	tt := zlsgo.NewTest(t)
	now := time.Now()
	var lock sync.Mutex
	var g sync.WaitGroup
	var ii = 1
	for i := 0; i < 5; i++ {
		g.Add(1)
		go WithLockContext(&lock, func() {
			time.Sleep(100 * time.Millisecond)
			diffTime := time.Since(now)
			expect := time.Duration(ii) * 100 * time.Millisecond
			el := (diffTime) > expect
			t.Log(diffTime, expect)
			tt.Equal(true, el)
			ii++
			g.Done()
		})
	}
	g.Wait()
}

func TestWithRunTimeContext(T *testing.T) {
	t := zlsgo.NewTest(T)
	now := time.Now().UnixNano()
	for i := 0; i < 5; i++ {
		duration := WithRunTimeContext(func() {
			time.Sleep(1 * time.Millisecond)
			newNow := time.Now()
			t.Equal(true, (newNow.UnixNano()-now) > 1000000)
			now = newNow.UnixNano()
		})
		t.Log(duration.String())
	}
}

func TestWithRunMemContext(t *testing.T) {
	tt := zlsgo.NewTest(t)
	mem := WithRunMemContext(func() {
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
	t.Log(zfile.SizeFormat(mem))
}

func TestIfVal(T *testing.T) {
	t := zlsgo.NewTest(T)
	i := IfVal(true, 1, 2)
	t.EqualExit(1, i)
	i = IfVal(false, 1, 2)
	t.EqualExit(2, i)
}

func TestTryCatch(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	errMsg := errors.New("test error")
	err := TryCatch(func() error {
		return errMsg
	})
	tt.Log(err)
	t.EqualTrue(err != nil)
	t.Equal(errMsg, err)

	err = TryCatch(func() error {
		panic(123)
		return nil
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

	Try(func() {
		panic("test")
	}, func(e interface{}) {
		if err, ok := e.(error); ok {
			t.EqualExit("test", err.Error())
		}
	}, func() {
		t.Log("TestTryError ok")
	})

	Try(func() {
		CheckErr(errors.New("test"))
	}, func(e interface{}) {
		if err, ok := e.(error); ok {
			t.EqualExit("test", err.Error())
		}
	})

	Try(func() {
		panic(t)
	}, func(e interface{}) {
		if err, ok := e.(error); ok {
			t.EqualExit("test", err.Error())
		}
	})

	Try(func() {
		panic("test")
	}, nil)
}

func TestUtil(t *testing.T) {
	_, _ = GetParentProcessName()
	IsDoubleClickStartUp()
}

func TestMaximizeOpenFileLimit(t *testing.T) {
	l, err := MaxRlimit()
	t.Log(l, err)
}
