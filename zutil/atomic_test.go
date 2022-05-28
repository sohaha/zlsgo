package zutil_test

import (
	"math/rand"
	"sync"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zutil"
)

func TestBool(t *testing.T) {
	tt := zlsgo.NewTest(t)

	isOk := zutil.NewBool(false)
	tt.EqualTrue(!isOk.Load())

	isOk.Store(true)
	tt.EqualTrue(isOk.Load())

	isOk.Store(false)
	tt.EqualTrue(!isOk.Load())

	tt.EqualTrue(isOk.CAS(false, true))

	tt.EqualTrue(isOk.Load())

	// the current is true
	tt.EqualTrue(!isOk.CAS(false, true))

	tt.EqualTrue(isOk.CAS(true, true))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			isOk.Toggle()
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestInt32(t *testing.T) {
	tt := zlsgo.NewTest(t)

	count := zutil.NewInt32(0)
	tt.EqualTrue(count.Load() == 0)

	var wg sync.WaitGroup
	l := rand.Intn(10000) + 10000

	count.Store(100)

	for i := 0; i < l; i++ {
		wg.Add(1)
		go func() {
			count.Add(1)
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			count.Sub(1)
			wg.Done()
		}()
	}
	wg.Wait()

	tt.Equal(int32(100), count.Load())
	count.Swap(200)
	tt.Equal(int32(200), count.Load())
	count.CAS(200, 300)
	tt.Equal(int32(300), count.Load())
}

func TestInt64(t *testing.T) {
	tt := zlsgo.NewTest(t)

	count := zutil.NewInt64(0)
	tt.EqualTrue(count.Load() == 0)

	var wg sync.WaitGroup
	l := rand.Intn(10000) + 10000

	count.Store(100)

	for i := 0; i < l; i++ {
		wg.Add(1)
		go func() {
			count.Add(1)
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			count.Sub(1)
			wg.Done()
		}()
	}
	wg.Wait()

	tt.Equal(int64(100), count.Load())
	count.Swap(200)
	tt.Equal(int64(200), count.Load())
	count.CAS(200, 300)
	tt.Equal(int64(300), count.Load())
}
