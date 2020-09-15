/*
go test -bench='.*' -benchmem -run none  -race ./zpool
*/

package zpool_test

import (
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo/zpool"
)

const runCount = 100000

func demoFunc() {
	time.Sleep(time.Duration(10) * time.Millisecond)
}

func BenchmarkGoroutines(b *testing.B) {
	var wg sync.WaitGroup

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(runCount)
		for j := 0; j < runCount; j++ {
			go func() {
				demoFunc()
				wg.Done()
			}()
		}
		wg.Wait()
	}
	b.StopTimer()
}

func BenchmarkPoolWorkerNum100(b *testing.B) {
	var wg sync.WaitGroup
	p := zpool.New(100)
	defer p.Close()
	_ = p.PreInit()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(runCount)
		for j := 0; j < runCount; j++ {
			_ = p.Do(func() {
				demoFunc()
				wg.Done()
			})
		}
		wg.Wait()
	}
	b.StopTimer()
}

func BenchmarkPoolWorkerNum500(b *testing.B) {
	var wg sync.WaitGroup
	p := zpool.New(500)
	defer p.Close()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(runCount)
		for j := 0; j < runCount; j++ {
			_ = p.Do(func() {
				demoFunc()
				wg.Done()
			})
		}
		wg.Wait()
	}
	b.StopTimer()
}

func BenchmarkPoolWorkerNum1500(b *testing.B) {
	var wg sync.WaitGroup
	p := zpool.New(1500)
	defer p.Close()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(runCount)
		for j := 0; j < runCount; j++ {
			_ = p.Do(func() {
				demoFunc()
				wg.Done()
			})
		}
		wg.Wait()
	}
	b.StopTimer()
}

func BenchmarkPoolWorkerNum15000(b *testing.B) {
	var wg sync.WaitGroup
	p := zpool.New(15000)
	defer p.Close()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(runCount)
		for j := 0; j < runCount; j++ {
			_ = p.Do(func() {
				demoFunc()
				wg.Done()
			})
		}
		wg.Wait()
	}
	b.StopTimer()
}
