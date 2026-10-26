package zhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestSSEErrorChannelNilWhenNoFailure verifies the happy path: a server that
// ends the stream cleanly never pushes a value on Error().
func TestSSEErrorChannelNilWhenNoFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: hello\n\n"))
	}))
	defer srv.Close()

	s, err := SSE(srv.URL, func(o *SSEOption) {
		o.RetryNum = 0
	}, NoRedirect(true))
	if err != nil {
		t.Fatalf("SSE() error = %v", err)
	}
	defer s.Close()

	go func() {
		for range s.Event() {
		}
	}()

	// The stream ends with io.EOF (no error); RetryNum=0 ends the loop with
	// err == nil, so Error() must not deliver a value. Note the engine only
	// closes eventCh when its loop exits, so poll instead of waiting on Done.
	deadline := time.After(3 * time.Second)
	for {
		select {
		case e := <-s.Error():
			t.Fatalf("Error() unexpectedly delivered: %v", e)
		case <-deadline:
			return
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func TestSSEErrorChannelSurfacesStreamFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: first\n\n"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		// Abort the connection mid-stream so the client stream errors out.
		hj := w.(http.Hijacker)
		conn, _, _ := hj.Hijack()
		_ = conn.Close()
	}))
	defer srv.Close()

	s, err := SSE(srv.URL, func(o *SSEOption) {
		o.RetryNum = 0
	}, NoRedirect(true))
	if err != nil {
		t.Fatalf("SSE() error = %v", err)
	}
	defer s.Close()

	// Consume the first event in the background so the engine goroutine can
	// proceed to the stream error.
	go func() {
		for range s.Event() {
		}
	}()

	deadline := time.After(5 * time.Second)
	for {
		select {
		case e := <-s.Error():
			if e == nil {
				t.Fatal("Error() delivered nil")
			}
			t.Logf("Error() delivered: %v", e)
			return
		case <-s.Done():
			t.Fatal("Done() closed before an error was delivered")
		case <-deadline:
			t.Fatal("timed out waiting for an error on Error()")
		}
	}
}
