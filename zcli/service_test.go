package zcli

import (
	"testing"
	"time"
)

func TestService(t *testing.T) {
	s, err := LaunchService("test", "", func() {
		t.Log("TestService")
	})
	t.Log(s, err)
}

func TestAppStartReturnsImmediately(t *testing.T) {
	done := make(chan struct{})
	a := &app{
		run: func() {
			time.Sleep(100 * time.Millisecond)
			close(done)
		},
	}

	start := time.Now()
	if err := a.Start(nil); err != nil {
		t.Fatal(err)
	}
	if time.Since(start) > 50*time.Millisecond {
		t.Fatal("Start should return immediately")
	}
	if err := a.Stop(nil); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("run function did not complete")
	}
}

func TestAppStopReturnsRunError(t *testing.T) {
	a := &app{
		run: func() {
			panic("boom")
		},
	}

	if err := a.Start(nil); err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	err := a.Stop(nil)
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected panic error, got %v", err)
	}
}
