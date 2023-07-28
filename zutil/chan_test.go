//go:build go1.18
// +build go1.18

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

	time.Sleep(time.Second / 4)
	tt.EqualTrue(ch.Len() >= 10)

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

	time.Sleep(time.Second / 4)
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

	time.Sleep(time.Second / 4)
	tt.EqualTrue(ch.Len() >= 3)

	for v := range ch.Out() {
		t.Log(v, ch.Len())
	}
}
