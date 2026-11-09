package znet

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// TestConcurrentRequestsDoNotRaceOnSharedMiddleware ensures that concurrent requests
// do not race on the shared middleware slice. The fix uses a full slice expression
// to converge the capacity, ensuring each request's chain lands in a private underlying array.
func TestConcurrentRequestsDoNotRaceOnSharedMiddleware(t *testing.T) {
	e := New()
	e.SetMode(TestMode)
	e.Use(func(c *Context) { c.Next() })
	e.GET("/race-probe", func(c *Context) {
		c.String(200, "ok")
	})

	srv := httptest.NewServer(e)
	defer srv.Close()

	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(srv.URL + "/race-probe")
			if err != nil {
				t.Error(err)
				return
			}
			resp.Body.Close()
		}()
	}
	wg.Wait()
}
