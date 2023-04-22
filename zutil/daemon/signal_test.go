package daemon

import (
	"context"
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
