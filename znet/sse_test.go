package znet

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestSSEStopDoesNotBlockWithPendingEvents(t *testing.T) {
	r := New("sse-test-" + t.Name())
	req := httptest.NewRequest("GET", "/sse", nil)
	w := httptest.NewRecorder()
	c := r.NewContext(w, req)
	sse := NewSSE(c)
	if err := sse.SendByte("1", []byte("data")); err != nil {
		t.Fatalf("send failed: %v", err)
	}

	done := make(chan struct{})
	go func() {
		sse.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Stop blocked with pending events")
	}
}

func TestSSESendByteAfterStop(t *testing.T) {
	r := New("sse-test-" + t.Name())
	req := httptest.NewRequest("GET", "/sse", nil)
	w := httptest.NewRecorder()
	c := r.NewContext(w, req)
	sse := NewSSE(c)
	sse.Stop()

	if err := sse.SendByte("1", []byte("data")); err == nil {
		t.Fatal("expected error when sending after stop")
	}
}

func TestSSESendCommentNonBlockingWhenFull(t *testing.T) {
	r := New("sse-test-" + t.Name())
	req := httptest.NewRequest("GET", "/sse", nil)
	w := httptest.NewRecorder()
	c := r.NewContext(w, req)
	sse := NewSSE(c)
	sse.events <- &sseEvent{Comment: "filled"}

	done := make(chan struct{})
	go func() {
		sse.sendComment()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("sendComment blocked")
	}
}
