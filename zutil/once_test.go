package zutil_test

import (
	"sync"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zutil"
)

func TestOnce(t *testing.T) {
	tt := zlsgo.NewTest(t)
	var v = 1
	var r = zutil.Once(func() interface{} {
		v = v + 1
		return v
	})
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value := (r()).(int)
			tt.Equal(2, value)
		}()
	}
	wg.Wait()
}

func TestOnceNested(t *testing.T) {
	tt := zlsgo.NewTest(t)
	var v = 1
	var r = zutil.Once(func() interface{} {
		v = v + 1
		return v
	})

	var v2 = 2
	var r2 = zutil.Once(func() interface{} {
		v2 = r().(int) + 2
		return v2
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value := (r2()).(int)
			tt.Equal(4, value)
		}()
	}
	wg.Wait()
}
