package zutil_test

import (
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zsync"
	"github.com/sohaha/zlsgo/zutil"
)

func TestOnce(t *testing.T) {
	tt := zlsgo.NewTest(t)
	v := zutil.NewInt32(1)
	r := zutil.Once(func() interface{} {
		v.Add(1)
		return v.Load()
	})

	var wg zsync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Go(func() {
			value := (r()).(int32)
			tt.Equal(int32(2), value)
		})
	}
	wg.Wait()
}

func TestOnceWithError(t *testing.T) {
	tt := zlsgo.NewTest(t)
	v := zutil.NewInt32(1)
	r := zutil.OnceWithError(func() (interface{}, error) {
		v.Add(1)
		return v.Load(), nil
	})

	var wg zsync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Go(func() {
			value, err := r()
			tt.NoError(err)
			tt.NotNil(value)
		})
	}
	wg.Wait()

	i := 0
	r2 := zutil.OnceWithError(func() (interface{}, error) {
		i++
		if i <= 100 {
			panic("error test")
		}
		return i, nil
	})

	for i := 0; i < 100; i++ {
		wg.Go(func() {
			value, err := r2()
			tt.NotNil(err)
			if !tt.IsNil(value) {
				tt.Log(value, err)
			}
		})
	}
	wg.Wait()

	value, err := r2()
	tt.Log(value, err)
}

func TestOnceNested(t *testing.T) {
	tt := zlsgo.NewTest(t)
	v := 1
	r := zutil.Once(func() interface{} {
		v = v + 1
		return v
	})

	v2 := 2
	r2 := zutil.Once(func() interface{} {
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

func TestGuard(t *testing.T) {
	tt := zlsgo.NewTest(t)
	v := 1
	r := zutil.Guard(func() int {
		v = v + 1
		time.Sleep(time.Second / 5)
		return v
	})

	errNum := zutil.NewInt32(0)
	successNum := zutil.NewInt32(0)
	var wg zsync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Go(func() {
			_, err := r()
			if err != nil {
				errNum.Add(1)
			} else {
				successNum.Add(1)
			}
		})
	}
	time.Sleep(time.Second / 3)

	_, err := r()
	if err != nil {
		errNum.Add(1)
	} else {
		successNum.Add(1)
	}

	wg.Wait()

	tt.Equal(errNum.Load(), int32(9))
	tt.Equal(successNum.Load(), int32(2))
}
