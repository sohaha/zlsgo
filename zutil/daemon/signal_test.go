package daemon

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	zls "github.com/sohaha/zlsgo"
)

func TestSignal(t *testing.T) {
	tt := zls.NewTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tip := "test"
	signalChan, _ := SignalChan()
	now := time.Now()
	select {
	case <-ctx.Done():
		tip = "timeout"
	case <-signalChan:
		tip = "signal"
	}
	t.Log(time.Since(now), tip)
	tt.Equal("timeout", tip)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case <-ctx.Done():
		t.Log("timeout")
	case k, ok := <-SingleKillSignal():
		tip = "kill"
		t.Log(k, ok)
	}
}

func TestSingleKillSignal(t *testing.T) {
	tt := zls.NewTest(t)
	go func() {
		time.Sleep(time.Second * 1)
		process, err := os.FindProcess(os.Getpid())
		tt.NoError(err, true)
		process.Signal(os.Interrupt)
	}()

	now := time.Now()
	isKill := <-SingleKillSignal()
	tt.Log(isKill)
	tt.EqualTrue(time.Since(now) > time.Second*1)

	ReSingleKillSignal()
	go func() {
		time.Sleep(time.Second * 2)
		process, err := os.FindProcess(os.Getpid())
		tt.NoError(err, true)
		process.Signal(os.Interrupt)
	}()

	now = time.Now()
	isKill = <-SingleKillSignal()
	tt.Log(isKill)
	tt.EqualTrue(time.Since(now) > time.Second*2)
}

func TestIsSudo(t *testing.T) {
	tt := zls.NewTest(t)
	tt.Log(IsSudo())
}

func TestConcurrentSignalAccess(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch := SingleKillSignal()
			ReSingleKillSignal()
			<-ch
		}()
	}

	wg.Wait()
}

func TestReSingleKillSignalCoverage(t *testing.T) {
	ReSingleKillSignal()

	ch := SingleKillSignal()

	ReSingleKillSignal()

	process, _ := os.FindProcess(os.Getpid())
	process.Signal(os.Interrupt)

	select {
	case <-ch:
	case <-time.After(time.Second * 2):
		t.Error("Expected to receive signal")
	}
}

func TestSignalChannel(t *testing.T) {
	sig, stop := SignalChan()
	if sig == nil {
		t.Error("Signal channel should not be nil")
	}
	if stop == nil {
		t.Error("Stop function should not be nil")
	}

	stop()

	select {
	case <-sig:
	default:
	}
}