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
		wg    WaitGroup
		total = 100
		maps  = make(map[int]int)
		mu    = NewRBMutex()
	)

	for i := 0; i < total; i++ {
		maps[i] = i
	}

	for i := 0; i < total; i++ {
		ii := i
		wg.Go(func() {
			mu.Lock()
			maps[ii*2] = ii * 2
			mu.Unlock()
		})

		wg.Go(func() {
			token := mu.RLock()
			tt.Equal(ii, maps[ii])
			mu.RUnlock(token)
		})
	}

	wg.Wait()
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
