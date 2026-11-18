package znet

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/ztype"
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

// TestHostIgnoresUpstreamForwardedHost ensures a forwarded upstream address
// never downgrades a Host header that already carries the public domain.
func TestHostIgnoresUpstreamForwardedHost(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	r := newHostProbeEngine(tt)

	got := doHostProbe(tt, r, func(req *http.Request) {
		req.Host = "zlsgo.example.com"
		req.Header.Set("X-Forwarded-Host", "127.0.0.1:8080")
		req.Header.Set("X-Forwarded-Proto", "https")
	})

	t.EqualExit("https://zlsgo.example.com|zlsgo.example.com", got)
}

// TestRedirectLocationBehindProxy ensures Context.Redirect emits the public
// domain in the Location header when the request comes through a proxy.
func TestRedirectLocationBehindProxy(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	r := New("forwarded-redirect-" + tt.Name())
	r.SetMode(TestMode)
	r.GET("/go", func(c *Context) {
		c.Redirect("/login")
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/go", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "127.0.0.1:8080"
	req.Header.Set("X-Forwarded-Host", "zlsgo.example.com")
	req.Header.Set("X-Forwarded-Proto", "https")
	r.ServeHTTP(w, req)

	t.EqualExit(http.StatusFound, w.Code)
	t.EqualExit("https://zlsgo.example.com/login", w.Header().Get("Location"))
}

// TestHostSchemeFromXScheme mirrors a proxy that terminates TLS and names the
// original scheme in X-Scheme instead of X-Forwarded-Proto, while still
// forwarding the upstream Host with an explicit :443 port.
func TestHostSchemeFromXScheme(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	r := newHostProbeEngine(tt)

	got := doHostProbe(tt, r, func(req *http.Request) {
		req.Host = "zlsgo.example.com:443"
		req.Header.Set("X-Forwarded-Host", "zlsgo.example.com")
		req.Header.Set("X-Scheme", "https")
	})

	t.EqualExit("https://zlsgo.example.com|zlsgo.example.com", got)
}

// TestHostSchemeFromXForwardedScheme covers the de-facto X-Forwarded-Scheme header.
func TestHostSchemeFromXForwardedScheme(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	r := newHostProbeEngine(tt)

	got := doHostProbe(tt, r, func(req *http.Request) {
		req.Host = "10.0.0.8:8080"
		req.Header.Set("X-Forwarded-Scheme", "https")
	})

	t.EqualExit("https://10.0.0.8:8080|10.0.0.8:8080", got)
}

// TestHostDropsDefaultPort keeps c.Host() canonical when the proxy forwards an
// explicit default port, e.g. https://example.com:443.
func TestHostDropsDefaultPort(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	r := newHostProbeEngine(tt)

	got := doHostProbe(tt, r, func(req *http.Request) {
		req.Host = "127.0.0.1:8080"
		req.Header.Set("X-Forwarded-Host", "zlsgo.example.com:443")
		req.Header.Set("X-Forwarded-Proto", "https")
	})

	t.EqualExit("https://zlsgo.example.com|zlsgo.example.com:443", got)
}

// TestHostKeepsNonDefaultPort ensures only the scheme's default port is removed.
func TestHostKeepsNonDefaultPort(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	r := newHostProbeEngine(tt)

	got := doHostProbe(tt, r, func(req *http.Request) {
		req.Host = "127.0.0.1:8080"
		req.Header.Set("X-Forwarded-Host", "zlsgo.example.com:8443")
		req.Header.Set("X-Forwarded-Proto", "https")
	})

	t.EqualExit("https://zlsgo.example.com:8443|zlsgo.example.com:8443", got)
}

// TestHostTrimsPortFromUpstreamHost covers `proxy_set_header Host $http_host`
// with `X-Forwarded-Proto $scheme`: the host comes from the forwarded Host
// header, which may carry an explicit :443, and must still render canonically.
func TestHostTrimsPortFromUpstreamHost(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	r := newHostProbeEngine(tt)

	got := doHostProbe(tt, r, func(req *http.Request) {
		req.Host = "zlsgo.example.com:443"
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Real-Ip", "203.0.113.7")
		req.Header.Set("X-Forwarded-For", "203.0.113.7")
	})

	t.EqualExit("https://zlsgo.example.com|zlsgo.example.com:443", got)
}

// TestHostTrustsLocalProxyPeers ensures loopback and private peers are treated
// as trusted proxies for both address families.
func TestHostTrustsLocalProxyPeers(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	restore := TrustedProxies
	TrustedProxies = []string{"0.0.0.0/0", "::/0"}
	defer func() { TrustedProxies = restore }()

	r := newHostProbeEngine(tt)
	for _, addr := range []string{"127.0.0.1:54321", "[::1]:54321", "172.17.0.1:54321"} {
		got := doHostProbe(tt, r, func(req *http.Request) {
			req.RemoteAddr = addr
			req.Header.Set("X-Forwarded-Host", "zlsgo.example.com")
			req.Header.Set("X-Forwarded-Proto", "https")
		})
		t.EqualExit("https://zlsgo.example.com|zlsgo.example.com", got)
	}
}

// forwardedHostController mirrors an application controller whose payload
// exposes the resolved domain and the client IP.
type forwardedHostController struct{}

func (h *forwardedHostController) GET(c *Context) (ztype.Map, error) {
	return ztype.Map{
		"hello":  c.GetClientIP(),
		"domain": c.Host(),
	}, nil
}

func doControllerRequest(tt *testing.T, mutate func(req *http.Request)) *httptest.ResponseRecorder {
	t := zlsgo.NewTest(tt)
	r := New("controller-host-" + tt.Name())
	r.SetMode(TestMode)
	t.NoError(r.BindStruct("/", &forwardedHostController{}))

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:3788/", nil)
	t.NoError(err)
	mutate(req)
	r.ServeHTTP(w, req)
	t.EqualExit(200, w.Code)
	return w
}

// TestControllerHostBehindProxy asserts the controller payload reports the
// public https domain when nginx forwards both host and proto.
func TestControllerHostBehindProxy(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	w := doControllerRequest(tt, func(req *http.Request) {
		req.Host = "zlsgo.example.com:443"
		req.Header.Set("X-Forwarded-Host", "zlsgo.example.com")
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Real-Ip", "203.0.113.7")
	})

	res := zjson.Parse(w.Body.String())
	t.EqualExit(0, res.Get("code").Int())
	t.EqualExit("https://zlsgo.example.com", res.Get("data.domain").String())
	t.EqualExit("203.0.113.7", res.Get("data.hello").String())
}

// TestControllerHostWithoutForwardedProto documents the reported symptom: when
// the proxy forwards Host but omits X-Forwarded-Proto, c.Host() can only see a
// plain HTTP request and reports http://.
func TestControllerHostWithoutForwardedProto(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	w := doControllerRequest(tt, func(req *http.Request) {
		req.Host = "zlsgo.example.com"
		req.Header.Set("X-Real-Ip", "203.0.113.7")
	})

	res := zjson.Parse(w.Body.String())
	t.EqualExit("http://zlsgo.example.com", res.Get("data.domain").String())
}
