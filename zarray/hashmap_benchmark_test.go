//go:build go1.18
// +build go1.18

package zarray_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/sohaha/zlsgo/zarray"
)

const size = 1 << 12

func BenchmarkGoSyncSet(b *testing.B) {
	var m sync.Map
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < size; i++ {
				m.Store(i, i)
			}
		}
	})
}

func BenchmarkHashMapSet(b *testing.B) {
	m := zarray.NewHashMap[int, int]()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < size; i++ {
				m.Set(i, i)
			}
		}
	})
}

func BenchmarkGoSyncGet(b *testing.B) {
	var m sync.Map
	for i := 0; i < size; i++ {
		m.Store(i, i)
	}
	var n int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if atomic.CompareAndSwapInt64(&n, 0, 1) {
				for pb.Next() {
					for i := 0; i < size; i++ {
						m.Store(i, i)
					}
				}
			} else {
				for pb.Next() {
					for i := 0; i < size; i++ {
						j, _ := m.Load(i)
						if j != i {
							b.Fail()
						}
					}
				}
			}
		}
	})
}

func BenchmarkHashMapGet(b *testing.B) {
	m := zarray.NewHashMap[int, int]()
	for i := 0; i < size; i++ {
		m.Set(i, i)
	}
	var n int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if atomic.CompareAndSwapInt64(&n, 0, 1) {
				for pb.Next() {
					for i := 0; i < size; i++ {
						m.Set(i, i)
					}
				}
			} else {
				for pb.Next() {
					for i := 0; i < size; i++ {
						j, _ := m.Get(i)
						if j != i {
							b.Fail()
						}
					}
				}
			}
		}
	})
}
