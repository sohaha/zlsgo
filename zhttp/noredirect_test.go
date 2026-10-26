package zhttp

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// newRedirectServer serves "/old" with a 302 to "/new" and "/new" with "ok".
func newRedirectServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/old", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/new", http.StatusFound)
	})
	mux.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// TestNoRedirectLeavesClientUntouched guards the regression where a per-request
// NoRedirect option rewrote the shared http.Client.CheckRedirect and restored
// it via defer, racing with concurrent requests on the same Engine.
func TestNoRedirectLeavesClientUntouched(t *testing.T) {
	srv := newRedirectServer(t)
	e := New()

	res, err := e.Get(srv.URL+"/old", NoRedirect(true))
	if err != nil {
		t.Fatalf("NoRedirect request error = %v", err)
	}
	if got := res.StatusCode(); got != http.StatusFound {
		t.Fatalf("NoRedirect status = %d, want %d", got, http.StatusFound)
	}
	if e.Client().CheckRedirect != nil {
		t.Fatal("shared client CheckRedirect was modified")
	}

	// The engine's default behaviour must still follow redirects.
	res, err = e.Get(srv.URL + "/old")
	if err != nil {
		t.Fatalf("follow request error = %v", err)
	}
	if got := res.String(); got != "ok" {
		t.Fatalf("followed body = %q, want %q", got, "ok")
	}
}

// TestNoRedirectConcurrentRequests runs redirect-disabling requests in
// parallel with plain ones on a single Engine. Run with -race.
func TestNoRedirectConcurrentRequests(t *testing.T) {
	srv := newRedirectServer(t)
	e := New()

	var wg sync.WaitGroup
	errs := make(chan error, 32)
	for i := 0; i < 16; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			res, err := e.Get(srv.URL+"/old", NoRedirect(true))
			if err != nil {
				errs <- err
				return
			}
			if code := res.StatusCode(); code != http.StatusFound {
				errs <- &statusError{got: code}
			}
		}()
		go func() {
			defer wg.Done()
			res, err := e.Get(srv.URL + "/old")
			if err != nil {
				errs <- err
				return
			}
			if body := res.String(); body != "ok" {
				errs <- &bodyError{got: body}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent request failed: %v", err)
	}
}

// TestDoWithArgsNoRedirect covers the RequestArgs path.
func TestDoWithArgsNoRedirect(t *testing.T) {
	srv := newRedirectServer(t)
	e := New()

	res, err := e.DoWithArgs(http.MethodGet, srv.URL+"/old", &RequestArgs{NoRedirect: true})
	if err != nil {
		t.Fatalf("DoWithArgs error = %v", err)
	}
	if code := res.StatusCode(); code != http.StatusFound {
		t.Fatalf("status = %d, want %d", code, http.StatusFound)
	}
	if e.Client().CheckRedirect != nil {
		t.Fatal("shared client CheckRedirect was modified")
	}
}

type statusError struct{ got int }

func (e *statusError) Error() string { return "unexpected status" }

type bodyError struct{ got string }

func (e *bodyError) Error() string { return "unexpected body" }
