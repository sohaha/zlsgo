package zsync

import (
	"runtime"
	"sync"
	"sync/atomic"
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

func TestRBMutexTokenCorrectUsage(t *testing.T) {
	tt := zlsgo.NewTest(t)

	mu := NewRBMutex()
	counter := 0

	token1 := mu.RLock()
	tt.Equal(0, counter)

	token2 := mu.RLock()
	tt.Equal(0, counter)

	mu.RUnlock(token2)
	mu.RUnlock(token1)

	mu.Lock()
	counter = 42
	mu.Unlock()

	tt.Equal(42, counter)
}

func TestRBMutexCorrectPairing(t *testing.T) {
	mu := NewRBMutex()
	var wg WaitGroup
	var shared int

	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Go(func() {
			token := mu.RLock()
			v := shared
			_ = v
			mu.RUnlock(token)
		})
	}

	wg.Go(func() {
		for i := 0; i < iterations; i++ {
			mu.Lock()
			shared++
			mu.Unlock()
		}
	})

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if shared != iterations {
		t.Errorf("Expected shared=%d, got %d", iterations, shared)
	}
}

// TestRBMutexProtectsMap verifies read-write exclusion with a read-heavy map.
// Without the writer re-check in the bias path, -race reports a map data race.
func TestRBMutexProtectsMap(t *testing.T) {
	mu := NewRBMutex()
	shared := make(map[int]int)

	const (
		rounds  = 50
		readers = 8
		writers = 4
		ops     = 500
	)

	var wg sync.WaitGroup
	for round := 0; round < rounds; round++ {
		for i := 0; i < readers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < ops; j++ {
					token := mu.RLock()
					_ = shared[j&127]
					mu.RUnlock(token)
				}
			}()
		}
		for i := 0; i < writers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < ops; j++ {
					mu.Lock()
					shared[j&127] = j
					mu.Unlock()
				}
			}()
		}
		wg.Wait()
	}
}

// TestRBMutexNoOverlap asserts that no reader is ever inside the critical
// section while a writer is inside its own. It fails loudly on a future
// regression instead of relying on -race to notice.
func TestRBMutexNoOverlap(t *testing.T) {
	mu := NewRBMutex()
	var (
		active  atomic.Int64
		writing atomic.Bool
		overlap atomic.Int64
		wg      sync.WaitGroup
	)

	const (
		readers = 8
		writers = 3
		ops     = 3000
	)

	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				token := mu.RLock()
				active.Add(1)
				if writing.Load() {
					overlap.Add(1)
				}
				active.Add(-1)
				mu.RUnlock(token)
			}
		}()
	}

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < ops/10; j++ {
				mu.Lock()
				if active.Load() != 0 {
					overlap.Add(1)
				}
				writing.Store(true)
				runtime.Gosched()
				writing.Store(false)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	if n := overlap.Load(); n != 0 {
		t.Fatalf("read and write critical sections overlapped %d times", n)
	}
}

func TestRBMutexNestedReadLocks(t *testing.T) {
	mu := NewRBMutex()
	shared := 0

	token1 := mu.RLock()
	_ = shared

	token2 := mu.RLock()
	_ = shared

	mu.RUnlock(token2)
	mu.RUnlock(token1)

	mu.Lock()
	shared = 42
	mu.Unlock()

	token3 := mu.RLock()
	if shared != 42 {
		t.Errorf("Expected shared=42, got %d", shared)
	}
	mu.RUnlock(token3)
}
