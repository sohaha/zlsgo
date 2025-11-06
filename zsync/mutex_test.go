package zsync

import (
	"runtime"
	"sync"
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestRBMutex(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var (
		wg      WaitGroup
		total   = 100
		counter int
		mu      = NewRBMutex()
	)

	for i := 0; i < total; i++ {
		wg.Go(func() {
			mu.Lock()
			counter++
			mu.Unlock()
		})
	}
	wg.Wait()
	tt.Equal(total, counter)

	counter = 42
	readValues := make([]int, total)
	for i := 0; i < total; i++ {
		ii := i
		wg.Go(func() {
			token := mu.RLock()
			readValues[ii] = counter
			mu.RUnlock(token)
		})
	}
	wg.Wait()

	for i := 0; i < total; i++ {
		tt.Equal(42, readValues[i])
	}
}

func BenchmarkRBMutexReadOnceAfterWrite(b *testing.B) {
	benchmarkReadOnceAfterWrite(b, func() func() {
		mu := NewRBMutex()
		shared := 0

		mu.Lock()
		shared = 42
		mu.Unlock()

		return func() {
			token := mu.RLock()
			_ = shared
			mu.RUnlock(token)
		}
	})
}

func BenchmarkRWMutexReadOnceAfterWrite(b *testing.B) {
	benchmarkReadOnceAfterWrite(b, func() func() {
		var mu sync.RWMutex
		shared := 0

		mu.Lock()
		shared = 42
		mu.Unlock()

		return func() {
			mu.RLock()
			_ = shared
			mu.RUnlock()
		}
	})
}

func benchmarkReadOnceAfterWrite(b *testing.B, setup func() func()) {
	readFn := setup()

	b.ReportAllocs()
	b.SetParallelism(runtime.GOMAXPROCS(0))
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			readFn()
		}
	})
}
