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
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zpool"
	"github.com/sohaha/zlsgo/zutil"
)

func TestBase(t *testing.T) {
	tt := zlsgo.NewTest(t)
	workerNum := 5
	p := zpool.New(workerNum)
	p.PanicFunc(func(err error) {
		zlog.Printf("panic: %v", err)
	})
	for i := 0; i < workerNum*2; i++ {
		ii := i
		err := p.Do(func() {
			t.Log("run", ii)
			time.Sleep(time.Millisecond * 300)
			t.Log("done", ii)
		})
		tt.EqualNil(err)
	}

	p.Wait()
}

func TestPool(t *testing.T) {
	tt := zlsgo.NewTest(t)

	count := 10000
	workerNum := 160
	var curMem int64
	for i := 0; i < 10; i++ {
		var g sync.WaitGroup
		var now time.Time
		runtime.GC()
		now = time.Now()
		_, curMem = zutil.WithRunContext(func() {
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
		_, curMem = zutil.WithRunContext(func() {
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
	p.Continue(newSize)
	tt.EqualExit(uint(newSize), p.Cap())

	restarSum := 7
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
	tt.EqualExit(uint(restarSum), p.Cap())

	p.Continue(1000)
	tt.EqualExit(uint(maxsize), p.Cap())

	p.Close()
	tt.EqualExit(uint(0), p.Cap())
}

func TestPoolPanicFunc(t *testing.T) {
	tt := zlsgo.NewTest(t)
	p := zpool.New(1)
	defErr := zerror.New(0, "test panic")
	var g sync.WaitGroup
	p.PanicFunc(func(err error) {
		g.Done()
		tt.Equal(err, defErr)
		t.Log(err)
	})
	i := 0

	g.Add(1)
	_ = p.Do(func() {
		zerror.Panic(defErr)
		i++
	})
	g.Wait()
	tt.EqualExit(0, i)

	g.Add(1)
	_ = p.Do(func() {
		i++
		zerror.Panic(defErr)
		i++
	})
	g.Wait()
	tt.EqualExit(1, i)

	g.Add(1)
	_ = p.Do(func() {
		i++
		g.Done()
	})

	g.Wait()
	tt.EqualExit(2, i)

	p.Pause()
	p.PanicFunc(func(err error) {
		t.Log("send again")
		defer g.Done()
	})
	p.Continue()
	g.Add(1)
	_ = p.Do(func() {
		i++
		zerror.Panic(defErr)
		i++
	})
	g.Wait()
	tt.EqualExit(3, i)
}

func TestPoolTimeout(t *testing.T) {
	tt := zlsgo.NewTest(t)
	p := zpool.New(1)
	for i := 0; i < 3; i++ {
		v := i
		err := p.DoWithTimeout(func() {
			t.Log(v)
			time.Sleep(time.Second)
		}, time.Second/3)
		t.Log(err)
		if v > 0 {
			tt.Equal(err, zpool.ErrWaitTimeout)
		}
		if err == zpool.ErrWaitTimeout {
			t.Log(v)
		}
	}
	p.Wait()
}

func TestPoolAuto(t *testing.T) {
	tt := zlsgo.NewTest(t)
	var g sync.WaitGroup
	p := zpool.New(1, 2)
	for i := 0; i < 4; i++ {
		g.Add(1)
		v := i
		err := p.DoWithTimeout(func() {
			time.Sleep(time.Second)
			t.Log("ok", v)
			g.Done()
		}, time.Second/6)
		t.Log(v, err)
		if v > 1 {
			tt.EqualTrue(err == zpool.ErrWaitTimeout)
		}
		if err == zpool.ErrWaitTimeout {
			go func() {
				time.Sleep(time.Second)
				t.Log("err", v)
				g.Done()
			}()
		}
	}
	g.Wait()
	tt.Log("done", p.Cap())
}
