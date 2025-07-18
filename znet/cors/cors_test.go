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
	// NewAllowHeaders默认允许所有来源，所以应该是204
	tt.Equal(http.StatusNoContent, w.Code)
	tt.Equal(0, w.Body.Len())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/TestNewAllowHeaders", nil)
	req.Header.Add("AllowTest", "https://qq.com")
	req.Header.Add("Origin", "https://qq.com")
	r.ServeHTTP(w, req)
	// 同样应该是200
	tt.Equal(http.StatusOK, w.Code)
	tt.Equal(10, w.Body.Len())
}

func TestDefault(t *testing.T) {
	tt := zls.NewTest(t)

	r := znet.New("TestDefault")
	r.SetMode(znet.ProdMode)

	r.Any("/cors", func(c *znet.Context) {
		c.String(200, zstring.Rand(10, "abc"))
	}, cors.New(&cors.Config{Domains: []string{"https://qq.com"}})) // 明确配置允许的域名
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

// 测试新的AllowAll函数
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

// 测试AllowAllOrigins函数
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

// 测试安全的默认配置
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
	
	// 默认配置应该拒绝未配置的域名
	tt.Equal(403, w.Code)
}

// 测试Origin验证
func TestOriginValidation(t *testing.T) {
	tt := zls.NewTest(t)
	r := znet.New()
	r.Use(cors.New(&cors.Config{
		Domains: []string{"https://trusted.com"},
	}))
	
	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	})
	
	// 测试无效的Origin格式
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "invalid-origin")
	
	r.ServeHTTP(w, req)
	tt.Equal(400, w.Code) // Bad Request for invalid origin
	
	// 测试过长的Origin
	longOrigin := "https://" + strings.Repeat("a", 2048) + ".com"
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", longOrigin)
	
	r.ServeHTTP(w, req)
	tt.Equal(400, w.Code)
}

// 测试配置验证
func TestConfigValidation(t *testing.T) {
	tt := zls.NewTest(t)
	
	// 测试无效的域名格式
	defer func() {
		if r := recover(); r != nil {
			tt.Log("Expected panic for invalid domain format:", r)
		}
	}()
	
	cors.New(&cors.Config{
		Domains: []string{"invalid-domain"},
	})
	
	// 如果没有panic，测试失败
	t.Error("Expected panic for invalid domain format")
}