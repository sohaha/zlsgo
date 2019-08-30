package znet

import (
	"fmt"
	"github.com/sohaha/zlsgo/zstring"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
)

var (
	expected = "hi"
	host     = "127.0.0.1"
)

var (
	one    sync.Once
	engine *Engine
)

func newServer() *Engine {
	one.Do(func() {
		engine = New("Web-test")
		engine.SetMode(DebugMode)
		engine.SetTimeout(3 * time.Second)
	})
	return engine
}

func newRequest(r *Engine, method string, url string, path string, handler ...HandlerFunc) *httptest.ResponseRecorder {
	method = strings.ToUpper(method)
	if len(handler) > 0 {
		firstHandler := handler[0]
		handlers := handler[1:]
		if path == "" {
			path = url
		}
		switch method {
		case "GET":
			r.GET(path, firstHandler, handlers...)
		case "POST":
			r.POST(path, firstHandler, handlers...)
		}
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, url, nil)
	req.Host = host
	r.ServeHTTP(w, req)
	return w
}

func TestWeb(t *testing.T) {
	T := zlsgo.NewTest(t)
	r := newServer()
	w := newRequest(r, "GET", "/", "/", func(c *Context) {
		_, _ = c.GetDataRaw()
		c.String(200, expected)
	})
	T.Equal(200, w.Code)
	T.Equal(expected, w.Body.String())
	r.GetMiddleware()
}

func TestPost(t *testing.T) {
	T := zlsgo.NewTest(t)
	r := newServer()
	w := newRequest(r, "POST", "/", "/", func(c *Context) {
		_, _ = c.GetDataRaw()
		c.String(200, expected)
	})
	T.Equal(200, w.Code)
	T.Equal(expected, w.Body.String())
}

func TestWebSetMode(T *testing.T) {
	t := zlsgo.NewTest(T)
	defer func() {
		if r := recover(); r != nil {
			t.Log("Recovered in f", r)
		}
	}()
	r := newServer()
	r.SetMode(DebugMode)
	t.Equal(true, r.IsDebug())
	r.SetMode(TestMode)
	r.SetMode(ReleaseMode)
	t.Equal(false, r.IsDebug())
	r.SetMode("")
	r.SetMode("unknownMode")
}

func TestWebRouter(T *testing.T) {
	t := zlsgo.NewTest(T)
	mux := newServer()

	testRouterNotFound(mux, t)
	testRouterCustomNotFound(mux, t)
	// testRouterPanicHandler(mux, t)
	testRouterCustomPanicHandler(mux, t)
	testRouterGET(mux, t)
}

func testRouterGET(r *Engine, t *zlsgo.TestUtil) {
	randString := zstring.Rand(5)

	w := newRequest(r, "GET", "/?id="+randString, "/", func(c *Context) {
		id := c.DefaultQuery("id", "not")
		host := c.Host()
		c.String(200, host+"|"+id)
	})

	t.Equal(200, w.Code)
	t.Equal("http://"+host+"|"+randString, w.Body.String())
}

func testRouterNotFound(r *Engine, t *zlsgo.TestUtil) {
	expectedText := "404 page not found\n"
	w := newRequest(r, "GET", "/404", "")
	t.Equal(404, w.Code)
	t.Equal(expectedText, w.Body.String())
}

func testRouterCustomNotFound(r *Engine, t *zlsgo.TestUtil) {
	expectedText := "is 404"
	r.NotFoundFunc(handleRes(expectedText))

	w := newRequest(r, "GET", "/404-2", "")
	t.Equal(200, w.Code)
	t.Equal(expectedText, w.Body.String())
}

func testRouterCustomPanicHandler(r *Engine, t *zlsgo.TestUtil) {
	expectedText := "panic"
	w := newRequest(r, "GET", "/panic", "", handleRes(expectedText))
	t.Equal(200, w.Code)
	t.Equal(expectedText, w.Body.String())
}

func handleRes(expected string) func(c *Context) {
	return func(c *Context) {
		_, _ = fmt.Fprint(c.Writer, expected)
	}
}

func TestGetInput(T *testing.T) {
	t := zlsgo.NewTest(T)
	r := newServer()
	getA := "abc"
	w := newRequest(r, "GET", "/"+getA+"?a="+getA, "/:name", func(c *Context) {
		a, _ := c.GetQuery("a")
		name := c.GetParam("name")
		GetAllQueryst := c.GetAllQueryst()
		t.Log(GetAllQueryst)
		t.Equal(getA, a)
		t.Equal(getA, name)
		t.Equal(url.Values{"a": []string{getA}}, GetAllQueryst)
		c.String(200, expected)
	})

	t.Equal(200, w.Code)
	t.Equal(expected, w.Body.String())
}
