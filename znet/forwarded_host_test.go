package znet

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sohaha/zlsgo"
)

// newHostProbeEngine returns an engine that writes c.Host() and c.Request.Host to the response body.
func newHostProbeEngine(t *testing.T) *Engine {
	r := New("forwarded-host-" + t.Name())
	r.SetMode(TestMode)
	r.GET("/host", func(c *Context) {
		c.String(200, c.Host()+"|"+c.Request.Host)
	})
	return r
}

func doHostProbe(t *testing.T, r *Engine, mutate func(req *http.Request)) string {
	t.Helper()
	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/host", nil)
	if err != nil {
		t.Fatal(err)
	}
	if mutate != nil {
		mutate(req)
	}
	r.ServeHTTP(w, req)
	return w.Body.String()
}

// TestHostBehindNginx reproduces an nginx reverse-proxy scenario where nginx rewrites Host to the upstream address by default
// (proxy_set_header Host $proxy_host), leaving the original domain only in X-Forwarded-Host.
func TestHostBehindNginx(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	r := newHostProbeEngine(tt)

	got := doHostProbe(tt, r, func(req *http.Request) {
		req.Host = "127.0.0.1:8080"
		req.Header.Set("X-Forwarded-Host", "zlsgo.example.com")
		req.Header.Set("X-Forwarded-Proto", "https")
	})

	t.EqualExit("https://zlsgo.example.com|zlsgo.example.com", got)
}

// TestHostBehindRFC7239Proxy covers the RFC 7239 Forwarded header.
func TestHostBehindRFC7239Proxy(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	r := newHostProbeEngine(tt)

	got := doHostProbe(tt, r, func(req *http.Request) {
		req.Host = "127.0.0.1:8080"
		req.Header.Set("Forwarded", `for=203.0.113.7;proto=https;host="api.example.com:8443"`)
	})

	t.EqualExit("https://api.example.com:8443|api.example.com:8443", got)
}

// TestHostDirectAccess ensures behavior is unchanged when no proxy headers are present.
func TestHostDirectAccess(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	r := newHostProbeEngine(tt)

	got := doHostProbe(tt, r, func(req *http.Request) {
		req.Host = "10.0.0.8:8080"
	})

	t.EqualExit("http://10.0.0.8:8080|10.0.0.8:8080", got)
}

// TestHostRejectsInjectedForwardedHost ensures untrusted forwarding headers cannot taint the response.
func TestHostRejectsInjectedForwardedHost(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	r := newHostProbeEngine(tt)

	got := doHostProbe(tt, r, func(req *http.Request) {
		req.Host = "10.0.0.8:8080"
		req.Header.Set("X-Forwarded-Host", "evil.example.com/redirect?to=bad")
	})

	t.EqualExit("http://10.0.0.8:8080|10.0.0.8:8080", got)
}
