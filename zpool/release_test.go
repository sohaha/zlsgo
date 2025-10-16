package zpool

import (
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zutil"
)

func TestPoolRelease(t *testing.T) {
	tt := zlsgo.NewTest(t)

	p := New(10)
	_ = p.PreInit()
	for i := 0; i < 4; i++ {
		_ = p.Do(func() {
			time.Sleep(time.Second)
		})
	}
	tt.Equal(uint(10), p.Cap())
	timer, mem := zutil.WithRunContext(func() {
		p.Close()
	})

	tt.EqualTrue(timer >= time.Second-100*time.Millisecond)
	t.Log(timer.String(), zfile.SizeFormat(mem))
	tt.Equal(uint(0), p.Cap())
}

func TestPoolAutoRelease(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var g sync.WaitGroup
	p := New(10)
	p.releaseTime = time.Second / 2
	_ = p.PreInit()

	for i := 0; i < 4; i++ {
		g.Add(1)
		_ = p.Do(func() {
			g.Done()
		})
	}
	g.Wait()
	tt.Equal(uint(10), p.Cap())
	time.Sleep(time.Second)

	tt.EqualTrue(p.Cap() <= uint(1))

	for i := 0; i < 6; i++ {
		g.Add(1)
		_ = p.Do(func() {
			time.Sleep(time.Second)
			tt.Equal(uint(6), p.Cap())
			g.Done()
		})
	}
	g.Wait()
	tt.Equal(uint(6), p.Cap())

	time.Sleep(time.Second)

	for i := 0; i < 6; i++ {
		g.Add(1)
		_ = p.Do(func() {
			g.Done()
		})
	}
	g.Wait()
	tt.EqualTrue(p.Cap() >= uint(1))
	time.Sleep(time.Second)
	tt.EqualTrue(p.Cap() <= uint(1))
}
