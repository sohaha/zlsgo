package cors_test

import (
	"net/http"
	"net/http/httptest"
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
	}, cors.Default())
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
