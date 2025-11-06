package zsync

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

func TestCoalescer(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Run("base", func(tt *zlsgo.TestUtil) {
		var count int32
		debounced := NewCoalescer(func() {
			atomic.AddInt32(&count, 1)
			time.Sleep(time.Millisecond * 10)
		})

		var wg WaitGroup
		for i := 0; i < 10; i++ {
			wg.Go(func() {
				debounced()
			})
		}
		wg.Wait()

		tt.Equal(int32(2), atomic.LoadInt32(&count))
	})

	tt.Run("no-loss coalescing", func(tt *zlsgo.TestUtil) {
		var count int32
		var debounced func()
		debounced = NewCoalescer(func() {
			atomic.AddInt32(&count, 1)
			for i := 0; i < 5; i++ {
				go debounced()
			}
			time.Sleep(time.Millisecond * 5)
		})

		for i := 0; i < 10; i++ {
			go debounced()
		}

		time.Sleep(time.Millisecond * 40)
		c := atomic.LoadInt32(&count)
		tt.EqualTrue(c >= 2 && c < 10)
	})

	tt.Run("panic releases running", func(tt *zlsgo.TestUtil) {
		var count int32
		first := int32(1)
		debounced := NewCoalescer(func() {
			if atomic.CompareAndSwapInt32(&first, 1, 0) {
				panic("boom")
			}
			atomic.AddInt32(&count, 1)
		})

		safeCall := func() {
			defer func() { _ = recover() }()
			debounced()
		}

		safeCall()

		debounced()
		debounced()

		time.Sleep(time.Millisecond * 10)
		v := atomic.LoadInt32(&count)
		tt.EqualTrue(v >= 1 && v <= 2)
	})

	tt.Run("nil function", func(tt *zlsgo.TestUtil) {
		defer func() {
			if r := recover(); r != nil {
				tt.Fatal("should not panic")
			}
		}()
		debounced := NewCoalescer(nil)
		debounced()
		tt.EqualTrue(true)
	})

	tt.Run("high frequency concurrency", func(tt *zlsgo.TestUtil) {
		var count int64
		debounced := NewCoalescer(func() {
			atomic.AddInt64(&count, 1)
			time.Sleep(time.Microsecond * 100)
		})

		var wg WaitGroup
		const n = 1000
		for i := 0; i < n; i++ {
			wg.Go(func() {
				debounced()
			})
		}
		wg.Wait()

		finalCount := atomic.LoadInt64(&count)
		tt.EqualTrue(finalCount > 0 && finalCount < n)
	})

	tt.Run("pending count accuracy", func(tt *zlsgo.TestUtil) {
		var execCount int32
		var pendingCalls int32
		debounced := NewCoalescer(func() {
			atomic.AddInt32(&execCount, 1)
			time.Sleep(time.Millisecond * 1)
		})

		var wg WaitGroup
		for i := 0; i < 50; i++ {
			wg.Go(func() {
				for j := 0; j < 10; j++ {
					debounced()
					atomic.AddInt32(&pendingCalls, 1)
				}
			})
		}
		wg.Wait()

		time.Sleep(time.Millisecond * 20)
		execs := atomic.LoadInt32(&execCount)
		pending := atomic.LoadInt32(&pendingCalls)
		tt.EqualTrue(execs > 0)
		tt.EqualTrue(pending > 0)
	})

	tt.Run("rapid successive calls", func(tt *zlsgo.TestUtil) {
		var count int32
		debounced := NewCoalescer(func() {
			atomic.AddInt32(&count, 1)
			time.Sleep(time.Millisecond)
		})

		var wg WaitGroup
		for i := 0; i < 500; i++ {
			wg.Go(func() { debounced() })
		}
		wg.Wait()

		time.Sleep(time.Millisecond * 10)
		finalCount := atomic.LoadInt32(&count)
		tt.EqualTrue(finalCount > 0 && finalCount < 500)
	})
}
