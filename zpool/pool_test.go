/*
go test -race ./zpool -v
*/
package zpool_test

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zpool"
)

func memFn(fn func()) uint64 {
	var mem = runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem := mem.TotalAlloc
	fn()
	runtime.ReadMemStats(&mem)
	return mem.TotalAlloc - curMem
}

const (
	_   = 1 << (10 * iota)
	KiB // 1024
	MiB // 1048576
)

func TestPool(t *testing.T) {
	tt := zlsgo.NewTest(t)

	count := 10000
	workerNum := 160
	var curMem uint64
	for i := 0; i < 10; i++ {
		var g sync.WaitGroup
		var now time.Time
		runtime.GC()
		now = time.Now()
		curMem = memFn(func() {
			now = time.Now()
			for i := 0; i < count; i++ {
				ii := i
				g.Add(1)
				go func() {
					_ = ii
					time.Sleep(time.Millisecond * 10)
					g.Done()
				}()
			}
			g.Wait()
		})
		t.Logf("NoPool memory:%v goroutines:%v time:%v \n", zfile.SizeFormat(curMem), count, time.Since(now))
		runtime.GC()
		p := zpool.New(workerNum)
		now = time.Now()
		curMem = memFn(func() {
			for i := 0; i < count; i++ {
				ii := i
				g.Add(1)
				err := p.Do(func() {
					_ = ii
					time.Sleep(time.Millisecond)
					g.Done()
				})
				tt.EqualNil(err)
			}
			g.Wait()
		})
		t.Logf("Pool   memory:%v goroutines:%v time:%v \n", zfile.SizeFormat(curMem), p.Cap(), time.Since(now))
		p.Close()
	}

	p := zpool.New(workerNum)
	_ = p.PreInit()
	tt.EqualExit(uint(workerNum), p.Cap())
	p.Close()

	err := p.PreInit()
	tt.EqualExit(true, err != nil)

	p.Close()
	c := p.IsClosed()
	tt.EqualExit(true, c)

	err = p.Do(func() {})
	tt.EqualExit(true, err != nil)
}

func TestPoolCap(t *testing.T) {
	tt := zlsgo.NewTest(t)
	p := zpool.New(0)
	g := sync.WaitGroup{}
	g.Add(1)
	err := p.Do(func() {
		time.Sleep(time.Second / 100)
		g.Done()
	})
	tt.EqualNil(err)
	tt.EqualExit(uint(1), p.Cap())
	g.Wait()

	maxsize := 10
	p = zpool.New(1, maxsize)
	g = sync.WaitGroup{}
	g.Add(1)
	err = p.Do(func() {
		time.Sleep(time.Second / 100)
		g.Done()
	})
	tt.EqualNil(err)
	tt.EqualExit(uint(1), p.Cap())
	g.Wait()
	p.Pause()
	tt.EqualExit(uint(0), p.Cap())

	newSize := 5
	go func() {
		tt.EqualExit(uint(0), p.Cap())
		time.Sleep(time.Second)
		p.Continue(newSize)
		tt.EqualExit(uint(newSize), p.Cap())
	}()

	restarSum := 6
	g.Add(restarSum)

	for i := 0; i < restarSum; i++ {
		ii := i
		go func() {
			now := time.Now()
			err := p.Do(func() {
				time.Sleep(time.Second / 100)
				t.Log("continue", ii, time.Since(now))
				g.Done()
			})
			tt.EqualNil(err)
		}()
	}
	g.Wait()
	tt.EqualExit(uint(newSize), p.Cap())

	p.Continue(1000)
	tt.EqualExit(uint(maxsize), p.Cap())

	p.Close()
	p.Continue(1000)
}
