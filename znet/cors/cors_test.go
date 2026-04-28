package cors_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	zls "github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/znet/cors"
	"github.com/sohaha/zlsgo/zstring"
)

func TestNewAllowHeaders(t *testing.T) {
	tt := zls.NewTest(t)

	r := znet.New("TestNewAllowHeaders")
	r.SetMode(znet.ProdMode)

	addAllowHeader, h := cors.NewAllowHeaders()
	r.Use(h)
	r.GET("/TestNewAllowHeaders", func(c *znet.Context) {
		c.Log.Debug("ok")
		c.String(200, zstring.Rand(10, "abc"))
	})
	addAllowHeader("AllowTest")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/TestNewAllowHeaders", nil)
	req.Header.Add("AllowTest", "https://qq.com")
	req.Header.Add("Origin", "https://qq.com")
	r.ServeHTTP(w, req)
	tt.Equal(http.StatusNoContent, w.Code)
	tt.Equal(0, w.Body.Len())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/TestNewAllowHeaders", nil)
	req.Header.Add("AllowTest", "https://qq.com")
	req.Header.Add("Origin", "https://qq.com")
	r.ServeHTTP(w, req)
	tt.Equal(http.StatusOK, w.Code)
	tt.Equal(10, w.Body.Len())
}

func TestDefault(t *testing.T) {
	tt := zls.NewTest(t)

	r := znet.New("TestDefault")
	r.SetMode(znet.ProdMode)

	r.Any("/cors", func(c *znet.Context) {
		c.String(200, zstring.Rand(10, "abc"))
	}, cors.New(&cors.Config{Domains: []string{"https://qq.com"}}))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/cors", nil)
	req.Header.Add("Origin", "https://qq.com")
	req.Host = "baidu.com"
	r.ServeHTTP(w, req)
	tt.Equal(http.StatusNoContent, w.Code)
	tt.Equal(0, w.Body.Len())

	r.Any("/cors2", func(c *znet.Context) {
		c.String(200, zstring.Rand(10, "abc"))
	}, cors.New(&cors.Config{Domains: []string{"*://?q.com"}}))
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/cors2", nil)
	req.Header.Add("Origin", "https://qq.com")
	req.Host = "baidu.com"
	r.ServeHTTP(w, req)
	tt.Equal(http.StatusNoContent, w.Code)
	tt.Equal(0, w.Body.Len())

	r.Any("/cors3", func(c *znet.Context) {
		c.String(200, zstring.Rand(10, "abc"))
	}, cors.New(&cors.Config{Domains: []string{"*://?q.com"}}))
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/cors3", nil)
	req.Header.Add("Origin", "https://qa.com")
	req.Host = "baidu.com"
	r.ServeHTTP(w, req)
	tt.Equal(http.StatusForbidden, w.Code)
	tt.Equal(0, w.Body.Len())

	r.Any("/cors3", func(c *znet.Context) {
		c.String(200, zstring.Rand(10, "abc"))
	}, cors.New(&cors.Config{Domains: []string{"*://?q.com"}, CustomHandler: func(conf *cors.Config, c *znet.Context) {
		c.Log.Debug(conf.Headers)
	}}))
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/cors3", nil)
	req.Header.Add("Origin", "https://qa.com")
	req.Host = "baidu.com"
	r.ServeHTTP(w, req)
	tt.Equal(http.StatusForbidden, w.Code)
	tt.Equal(0, w.Body.Len())
}

func TestAllowAll(t *testing.T) {
	tt := zls.NewTest(t)
	r := znet.New()
	r.Use(cors.AllowAll())

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("X-Custom-Header", "test")

	r.ServeHTTP(w, req)

	tt.Equal(200, w.Code)
	tt.Equal("https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	tt.EqualTrue(w.Header().Get("Access-Control-Allow-Headers") != "")
}

func TestAllowAllOrigins(t *testing.T) {
	tt := zls.NewTest(t)
	r := znet.New()
	r.Use(cors.AllowAllOrigins())

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://malicious.com")

	r.ServeHTTP(w, req)

	tt.Equal(200, w.Code)
	tt.Equal("https://malicious.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestSecureDefault(t *testing.T) {
	tt := zls.NewTest(t)
	r := znet.New()
	r.Use(cors.Default())

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://untrusted.com")

	r.ServeHTTP(w, req)

	tt.Equal(403, w.Code)
}

func TestOriginValidation(t *testing.T) {
	tt := zls.NewTest(t)
	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains: []string{"https://trusted.com"},
	}))

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "invalid-origin")

	r.ServeHTTP(w, req)
	tt.Equal(400, w.Code)

	longOrigin := "https://" + strings.Repeat("a", 2048) + ".com"
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", longOrigin)

	r.ServeHTTP(w, req)
	tt.Equal(400, w.Code)
}

func TestRefererFallback(t *testing.T) {
	tt := zls.NewTest(t)
	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains: []string{"http://example.com"},
	}))

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Referer", "http://example.com/page")

	r.ServeHTTP(w, req)
	tt.Equal(200, w.Code)
	tt.Equal("http://example.com", w.Header().Get("Access-Control-Allow-Origin"))

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("Referer", "http://untrusted.com/page")

	r.ServeHTTP(w, req)
	tt.Equal(403, w.Code)
}

func TestCustomMethods(t *testing.T) {
	tt := zls.NewTest(t)
	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains: []string{"*"},
		Methods: []string{"GET", "POST", "PUT"},
	}))

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")

	r.ServeHTTP(w, req)
	tt.Equal(200, w.Code)
	tt.Equal("GET, POST, PUT", w.Header().Get("Access-Control-Allow-Methods"))
}

func TestExposeHeaders(t *testing.T) {
	tt := zls.NewTest(t)
	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains:       []string{"*"},
		ExposeHeaders: []string{"X-Custom-Header", "X-Another-Header"},
	}))

	r.GET("/test", func(c *znet.Context) {
		c.SetHeader("X-Custom-Header", "value")
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")

	r.ServeHTTP(w, req)
	tt.Equal(200, w.Code)
	tt.Equal("X-Custom-Header, X-Another-Header", w.Header().Get("Access-Control-Expose-Headers"))
}

func TestCredentialsSupport(t *testing.T) {
	tt := zls.NewTest(t)
	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains:     []string{"http://example.com"},
		Credentials: []string{"true"},
	}))

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")

	r.ServeHTTP(w, req)
	tt.Equal(200, w.Code)
	tt.Equal("true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCustomHandler(t *testing.T) {
	tt := zls.NewTest(t)
	customHeaderSet := false

	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains: []string{"*"},
		CustomHandler: func(conf *cors.Config, c *znet.Context) {
			c.SetHeader("X-Custom-CORS", "custom-value")
			customHeaderSet = true
		},
	}))

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")

	r.ServeHTTP(w, req)
	tt.Equal(200, w.Code)
	tt.EqualTrue(customHeaderSet)
	tt.Equal("custom-value", w.Header().Get("X-Custom-CORS"))
}

func TestInvalidDomainFormatPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid domain format")
		}
	}()

	_ = cors.New(&cors.Config{
		Domains: []string{"invalid-domain-without-protocol"},
	})
}

func TestInvalidMethodPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid HTTP method")
		}
	}()

	_ = cors.New(&cors.Config{
		Domains: []string{"*"},
		Methods: []string{"INVALID-METHOD"},
	})
}

func TestPreflightDetailedTests(t *testing.T) {
	tt := zls.NewTest(t)
	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains: []string{"*"},
	}))

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	// Test standard preflight
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	r.ServeHTTP(w, req)
	tt.Equal(204, w.Code)
	tt.Equal(0, w.Body.Len())
	tt.EqualTrue(w.Header().Get("Access-Control-Allow-Methods") != "")
	tt.EqualTrue(w.Header().Get("Access-Control-Allow-Headers") != "")
}

func TestWildcardHeaders(t *testing.T) {
	tt := zls.NewTest(t)
	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains: []string{"*"},
		Headers: []string{"*"},
	}))

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "X-Custom-1, X-Custom-2")

	r.ServeHTTP(w, req)
	tt.Equal(204, w.Code)

	allowHeaders := w.Header().Get("Access-Control-Allow-Headers")
	tt.EqualTrue(allowHeaders != "")
}

func TestNoOrigin(t *testing.T) {
	tt := zls.NewTest(t)
	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains: []string{"*"},
	}))

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)
	tt.Equal(200, w.Code)
}

func TestOriginWithInvalidProtocol(t *testing.T) {
	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains: []string{"*"},
	}))

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	invalidOrigins := []string{
		"ftp://example.com",
		"javascript:alert('xss')",
		"data:text/html,<script>alert('xss')</script>",
	}

	for _, origin := range invalidOrigins {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", origin)

		r.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Errorf("Origin %s should be rejected with status 400, got %d", origin, w.Code)
		}
	}
}

func TestOriginWithPort(t *testing.T) {
	tt := zls.NewTest(t)
	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains: []string{"http://localhost:*"},
	}))

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	// Test with port
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:8080")

	r.ServeHTTP(w, req)
	tt.Equal(200, w.Code)
	tt.Equal("http://localhost:8080", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestDomainMatching(t *testing.T) {
	tt := zls.NewTest(t)
	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains: []string{"https://*.example.com"},
	}))

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	// Test subdomain matching
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://sub.example.com")

	r.ServeHTTP(w, req)
	tt.Equal(200, w.Code)

	// Test non-matching domain
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://other.com")

	r.ServeHTTP(w, req)
	tt.Equal(403, w.Code)
}

func TestCORSConcurrent(t *testing.T) {
	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains: []string{"https://example.com", "https://trusted.com"},
	}))

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	// Test concurrent requests from different origins
	origins := []string{
		"https://example.com",
		"https://trusted.com",
		"https://malicious.com",
		"http://untrusted.com",
	}

	// Run concurrent requests
	done := make(chan bool, len(origins)*10) // 10 requests per origin

	for i := 0; i < 10; i++ {
		for _, origin := range origins {
			go func(orig string) {
				defer func() { done <- true }()

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/test", nil)
				req.Header.Set("Origin", orig)

				r.ServeHTTP(w, req)

				// Verify expected behavior
				if orig == "https://example.com" || orig == "https://trusted.com" {
					if w.Code != 200 {
						t.Errorf("Expected status 200 for %s, got %d", orig, w.Code)
					}
					if w.Header().Get("Access-Control-Allow-Origin") == "" {
						t.Errorf("Missing CORS header for %s", orig)
					}
				} else {
					if w.Code != 403 {
						t.Errorf("Expected status 403 for %s, got %d", orig, w.Code)
					}
				}
			}(origin)
		}
	}

	// Wait for all goroutines to complete
	for i := 0; i < len(origins)*10; i++ {
		<-done
	}
}

func TestCORSConcurrentPreflight(t *testing.T) {
	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains: []string{"*"},
	}))

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	// Simulate concurrent preflight requests
	done := make(chan bool, 20)

	for i := 0; i < 20; i++ {
		go func(id int) {
			defer func() { done <- true }()

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("OPTIONS", "/test", nil)
			req.Header.Set("Origin", "http://example.com")
			req.Header.Set("Access-Control-Request-Method", "POST")
			req.Header.Set("Access-Control-Request-Headers", "Content-Type")

			r.ServeHTTP(w, req)

			if w.Code != 204 {
				t.Errorf("Request %d: expected status 204, got %d", id, w.Code)
			}

			// Verify CORS headers are present
			if w.Header().Get("Access-Control-Allow-Methods") == "" {
				t.Errorf("Request %d: missing Allow-Methods header", id)
			}
			if w.Header().Get("Access-Control-Allow-Headers") == "" {
				t.Errorf("Request %d: missing Allow-Headers header", id)
			}
		}(i)
	}

	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestCORSConcurrentMixed(t *testing.T) {
	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains: []string{"https://*.example.com"},
	}))

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	done := make(chan bool, 30)

	for i := 0; i < 15; i++ {
		go func(id int) {
			defer func() { done <- true }()

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", "https://sub.example.com")

			r.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("GET request %d: expected status 200, got %d", id, w.Code)
			}
		}(i)

		go func(id int) {
			defer func() { done <- true }()

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("OPTIONS", "/test", nil)
			req.Header.Set("Origin", "https://api.example.com")
			req.Header.Set("Access-Control-Request-Method", "GET")

			r.ServeHTTP(w, req)

			if w.Code != 204 {
				t.Errorf("OPTIONS request %d: expected status 204, got %d", id, w.Code)
			}
		}(i)
	}

	for i := 0; i < 30; i++ {
		<-done
	}
}

func TestCORSConcurrentCustomHandler(t *testing.T) {
	customHandlerCalled := make(map[int]bool)
	customHandlerMu := sync.Mutex{}

	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains: []string{"*"},
		CustomHandler: func(conf *cors.Config, c *znet.Context) {
			// Simulate some custom processing
			requestID := c.GetHeader("X-Request-ID")
			if requestID != "" {
				customHandlerMu.Lock()
				customHandlerCalled[len(customHandlerCalled)] = true
				customHandlerMu.Unlock()
			}
		},
	}))

	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", "http://example.com")
			req.Header.Set("X-Request-ID", string(rune('0'+id)))

			r.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("Request %d: expected status 200, got %d", id, w.Code)
			}
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	customHandlerMu.Lock()
	calls := len(customHandlerCalled)
	customHandlerMu.Unlock()

	if calls == 0 {
		t.Error("Expected custom handler to be called at least once")
	}
}
