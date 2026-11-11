package zsync

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

// BenchmarkRBMutexReadUncontended measures the single-goroutine read path.
func BenchmarkRBMutexReadUncontended(b *testing.B) {
	mu := NewRBMutex()
	shared := 42
	b.ReportAllocs()
	for b.Loop() {
		token := mu.RLock()
		_ = shared
		mu.RUnlock(token)
	}
}

func BenchmarkRWMutexReadUncontended(b *testing.B) {
	var mu sync.RWMutex
	shared := 42
	b.ReportAllocs()
	for b.Loop() {
		mu.RLock()
		_ = shared
		mu.RUnlock()
	}
}

// BenchmarkRBMutexReadParallel measures pure reader scaling with no writer.
func BenchmarkRBMutexReadParallel(b *testing.B) {
	mu := NewRBMutex()
	shared := 42
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			token := mu.RLock()
			_ = shared
			mu.RUnlock(token)
		}
	})
}

func BenchmarkRWMutexReadParallel(b *testing.B) {
	var mu sync.RWMutex
	shared := 42
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.RLock()
			_ = shared
			mu.RUnlock()
		}
	})
}

// BenchmarkRBMutexReadMostly mixes occasional writers into a reader pool,
// which is the workload the read-bias path is supposed to win.
func BenchmarkRBMutexReadMostly(b *testing.B) {
	for _, writeEvery := range []uint32{4096, 256, 32} {
		b.Run(fmt.Sprintf("write1of%d", writeEvery), func(b *testing.B) {
			mu := NewRBMutex()
			shared := 0
			var seq atomic.Uint32
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if seq.Add(1)%writeEvery == 0 {
						mu.Lock()
						shared++
						mu.Unlock()
						continue
					}
					token := mu.RLock()
					_ = shared
					mu.RUnlock(token)
				}
			})
		})
	}
}

func BenchmarkRWMutexReadMostly(b *testing.B) {
	for _, writeEvery := range []uint32{4096, 256, 32} {
		b.Run(fmt.Sprintf("write1of%d", writeEvery), func(b *testing.B) {
			var mu sync.RWMutex
			shared := 0
			var seq atomic.Uint32
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if seq.Add(1)%writeEvery == 0 {
						mu.Lock()
						shared++
						mu.Unlock()
						continue
					}
					mu.RLock()
					_ = shared
					mu.RUnlock()
				}
			})
		})
	}
}

// BenchmarkRBMutexWriteParallel measures writer throughput while readers are idle.
func BenchmarkRBMutexWriteParallel(b *testing.B) {
	mu := NewRBMutex()
	shared := 0
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			shared++
			mu.Unlock()
		}
	})
}

func BenchmarkRWMutexWriteParallel(b *testing.B) {
	var mu sync.RWMutex
	shared := 0
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			shared++
			mu.Unlock()
		}
	})
}
