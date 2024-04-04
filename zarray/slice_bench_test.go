//go:build go1.18
// +build go1.18

package zarray_test

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/ztype"
)

func BenchmarkMap(b *testing.B) {
	b.Run("Map", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				zarray.Map(l2, func(i int, v int) string {
					return ztype.ToString(v) + "//"
				})
			}
		})
	})

	b.Run("ParallelMap", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				zarray.ParallelMap(l2, func(i int, v int) string {
					return ztype.ToString(v) + "//"
				}, uint(len(l2)))
			}
		})
	})

	b.Run("Map_timeConsuming", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				zarray.Map(l2, func(i int, v int) string {
					time.Sleep(time.Microsecond)
					return ztype.ToString(v) + "//"
				})
			}
		})
	})

	b.Run("ParallelMap_timeConsuming", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				zarray.ParallelMap(l2, func(i int, v int) string {
					time.Sleep(time.Microsecond)
					return ztype.ToString(v) + "//"
				}, uint(len(l2)))
			}
		})
	})

}
