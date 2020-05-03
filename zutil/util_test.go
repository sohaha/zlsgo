package zutil

import (
	"errors"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

func TestWithLockContext(T *testing.T) {
	t := zlsgo.NewTest(T)
	now := time.Now().UnixNano()
	for i := 0; i < 5; i++ {
		WithLockContext(func() {
			time.Sleep(100 * time.Millisecond)
			newNow := time.Now()
			el := (newNow.UnixNano() - now) > 1000000
			t.Log(el)
			t.Equal(true, el)
			now = newNow.UnixNano()
		})
	}
}

func TestWithRunTimeContext(T *testing.T) {
	t := zlsgo.NewTest(T)
	now := time.Now().UnixNano()
	for i := 0; i < 5; i++ {
		WithRunTimeContext(func() {
			time.Sleep(1 * time.Millisecond)
			newNow := time.Now()
			t.Equal(true, (newNow.UnixNano()-now) > 1000000)
			now = newNow.UnixNano()
		}, func(duration time.Duration) {
			t.Log(duration.String())
		})
	}
}

func TestIfVal(T *testing.T) {
	t := zlsgo.NewTest(T)
	i := IfVal(true, 1, 2)
	t.EqualExit(1, i)
	i = IfVal(false, 1, 2)
	t.EqualExit(2, i)
}

func TestTryError(T *testing.T) {
	t := zlsgo.NewTest(T)
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
