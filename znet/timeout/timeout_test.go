package timeout

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/znet"
)

func TestWebTimeout(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	r := newServer()
	body := ""
	w1 := newRequest(r, "GET", "/timeout_1", func(c *znet.Context) {
		tt.Log("timeout_1")
		c.String(201, "timeout_1")
	}, New(2*time.Second, func(c *znet.Context) {
		tt.Log("timeout_1 Second")
	}), func(c *znet.Context) {
		c.Next()
	})
	body = w1.Body.String()
	tt.Log("code:", w1.Code)
	tt.Log("body:", body)
	t.Equal(201, w1.Code)
	t.Equal("timeout_1", body)

	w2 := newRequest(r, "GET", "/timeout_2", func(c *znet.Context) {
		time.Sleep(2 * time.Second)
		t.Log("run 2")
		c.String(200, "timeout_2")
	}, New(1*time.Second))
	t.Equal(504, w2.Code)
	t.Equal("", w2.Body.String())

	w3 := newRequest(r, "GET", "/timeout_3", func(c *znet.Context) {
		time.Sleep(2 * time.Second)
		c.String(200, "timeout_3")
	}, New(1*time.Second, func(c *znet.Context) {
		tt.Log("3 timeout")
		c.String(210, "is timeout")
	}))
	t.Equal(210, w3.Code)
	t.Equal("is timeout", w3.Body.String())
	tt.Log(w3.Body.String())

	w4 := newRequest(r, "GET", "/timeout_4", func(c *znet.Context) {
		time.Sleep(3 * time.Second)
		c.String(200, "timeout_2")
	}, New(1*time.Second, func(c *znet.Context) {
		c.String(211, "ok")
	}))
	t.Equal(211, w4.Code)
	t.Equal("ok", w4.Body.String())

	w5 := newRequest(r, "GET", "/timeout_5", func(c *znet.Context) {
		c.String(200, "timeout_5")
	}, New(1*time.Second, func(c *znet.Context) {
		c.String(211, "ok")
	}))
	t.Equal(200, w5.Code)
	t.Equal("timeout_5", w5.Body.String())
}

var (
	one    sync.Once
	Engine *znet.Engine
)

func newServer() *znet.Engine {
	one.Do(func() {
		Engine = znet.New()
		Engine.SetMode(znet.DebugMode)
	})
	return Engine
}

func newRequest(r *znet.Engine, method string, path string, handler ...znet.Handler) *httptest.ResponseRecorder {
	method = strings.ToUpper(method)
	if len(handler) > 0 {
		firstHandler := handler[0]
		handlers := handler[1:]
		switch method {
		case "GET":
			r.GET(path, firstHandler, handlers...)
		case "POST":
			r.POST(path, firstHandler, handlers...)
		}
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	r.ServeHTTP(w, req)
	return w
}
