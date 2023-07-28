package zutil_test

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zsync"
	"github.com/sohaha/zlsgo/zutil"
)

func TestNewChanUnbounded(t *testing.T) {
	tt := zlsgo.NewTest(t)
	ch := zutil.NewChan[any]()

	var wg zsync.WaitGroup

	for i := 0; i < 10; i++ {
		i := i
		wg.Go(func() {
			ch.In() <- i
		})
	}

	go func() {
		_ = wg.Wait()
		ch.Close()
	}()

	time.Sleep(time.Second)
	tt.Equal(10, ch.Len())
	for v := range ch.Out() {
		t.Log(v, ch.Len())
	}
}

func TestNewChanUnbuffered(t *testing.T) {
	tt := zlsgo.NewTest(t)
	ch := zutil.NewChan[any](0)

	var wg zsync.WaitGroup

	for i := 0; i < 10; i++ {
		i := i
		wg.Go(func() {
			ch.In() <- i
		})
	}

	go func() {
		_ = wg.Wait()
		ch.Close()
	}()

	time.Sleep(time.Second)
	tt.Equal(0, ch.Len())
	for v := range ch.Out() {
		t.Log(v, ch.Len())
	}
}

func TestNewChanBuffered(t *testing.T) {
	tt := zlsgo.NewTest(t)
	ch := zutil.NewChan[any](3)

	var wg zsync.WaitGroup

	for i := 0; i < 10; i++ {
		i := i
		wg.Go(func() {
			ch.In() <- i
		})
	}

	go func() {
		_ = wg.Wait()
		ch.Close()
	}()

	time.Sleep(time.Second)
	tt.Equal(3, ch.Len())
	for v := range ch.Out() {
		t.Log(v, ch.Len())
	}
}
