package cors_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
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
	// NewAllowHeaders allows all origins by default, so it should be 204
	tt.Equal(http.StatusNoContent, w.Code)
	tt.Equal(0, w.Body.Len())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/TestNewAllowHeaders", nil)
	req.Header.Add("AllowTest", "https://qq.com")
	req.Header.Add("Origin", "https://qq.com")
	r.ServeHTTP(w, req)
	// Should also be 200
	tt.Equal(http.StatusOK, w.Code)
	tt.Equal(10, w.Body.Len())
}

func TestDefault(t *testing.T) {
	tt := zls.NewTest(t)

	r := znet.New("TestDefault")
	r.SetMode(znet.ProdMode)

	r.Any("/cors", func(c *znet.Context) {
		c.String(200, zstring.Rand(10, "abc"))
	}, cors.New(&cors.Config{Domains: []string{"https://qq.com"}})) // Explicitly configure allowed domains
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

// Test new AllowAll function
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

// Test AllowAllOrigins function
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

// Test secure default configuration
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
	
	// Default configuration should reject unconfigured domains
	tt.Equal(403, w.Code)
}

// Test Origin validation
func TestOriginValidation(t *testing.T) {
	tt := zls.NewTest(t)
	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains: []string{"https://trusted.com"},
	}))
	
	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})
	
	// Test invalid Origin format
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "invalid-origin")
	
	r.ServeHTTP(w, req)
	tt.Equal(400, w.Code) // Bad Request for invalid origin
	
	// Test overly long Origin
	longOrigin := "https://" + strings.Repeat("a", 2048) + ".com"
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", longOrigin)
	
	r.ServeHTTP(w, req)
	tt.Equal(400, w.Code)
}

// Test configuration validation
func TestConfigValidation(t *testing.T) {
	tt := zls.NewTest(t)
	
	// Test invalid domain format
	defer func() {
		if r := recover(); r != nil {
			tt.Log("Expected panic for invalid domain format:", r)
		}
	}()
	
	cors.New(&cors.Config{
		Domains: []string{"invalid-domain"},
	})
	
	// If no panic, test fails
	t.Error("Expected panic for invalid domain format")
}