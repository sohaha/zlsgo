/*
 * @Author: seekwe
 * @Date:   2019-05-09 12:48:09
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-29 17:49:22
 */

package znet

import (
	"fmt"
	"github.com/sohaha/zlsgo/zstring"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/sohaha/zlsgo"
)

var (
	defrouter, errorFormat, expected string
	host                             = "127.0.0.1"
)

func init() {
	defrouter = "hi"
	expected = "Hello World "
	errorFormat = "handler returned unexpected body: got %v want %v"
}

var (
	one    sync.Once
	engine *Engine
)

func newServer() *Engine {
	one.Do(func() {
		engine = New("Web-test")
		engine.SetMode(DebugMode)
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
	r.SetMode(TestMode)
	r.SetMode(ReleaseMode)
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
	// 	testRouterParam(mux, T)()
	// 	testRouterRegex(mux, T)()
	// 	testRouterPOST(mux, T)()
	// 	testRouterPUT(mux, T)()
	// 	testRouterPATCH(mux, T)()
	// 	testRouterDELETE(mux, T)()
	// 	testRouterUse(mux, T)()
	// 	testRouterGroup(mux, T)()
	// 	testRouterMatch(mux, T)()
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

// func testRouterParam(mux *Engine, T *zlsgo.TestUtil) func() {
// 	now := ztype.ToString(time.Now().Unix())
// 	expectedText := expected + "GETParam"
// 	currentDefrouter := "/" + defrouter + "/Param"
// 	mux.GET(currentDefrouter+"/id/:id", func(c *Context) {
// 		id := c.GetParam("id")
// 		handle(expectedText + id)(c)
// 	})
// 	mux.POST(currentDefrouter+"/id/:id/op/:now", func(c *Context) {
// 		_now := c.GetParam("now")
// 		id := c.GetParam("id")
// 		pa := c.GetAllParams()
// 		s := map[string]string{"id": now, "now": "123"}
// 		for k, _ := range s {
// 			T.Equal(k+pa[k], k+s[k])
// 		}

// 		handle(expectedText + id + "+" + _now)(c)
// 	})
// 	return func() {
// 		w := httptest.NewRecorder()
// 		req, _ := http.NewRequest(http.MethodGet, currentDefrouter+"/id/"+now, nil)
// 		mux.ServeHTTP(w, req)
// 		body := w.Body.String()
// 		T.T.Logf("body: %s", body)
// 		T.Equal(expectedText+now, body)

// 		w = httptest.NewRecorder()
// 		req, _ = http.NewRequest("POST", currentDefrouter+"/id/"+now+"/op/123", nil)
// 		mux.ServeHTTP(w, req)
// 		body = w.Body.String()
// 		T.T.Logf("body: %s", body)
// 		T.Equal(expectedText+now+"+123", body)
// 	}
// }

// func testRouterRegex(mux *Engine, T *zlsgo.TestUtil) func() {
// 	now := ztype.ToString(time.Now().Unix())
// 	expectedText := expected + "GETParam"
// 	currentDefrouter := "/" + defrouter + "/Regex"
// 	mux.GET(currentDefrouter+"/id2/{id:[0-9]+}", func(c *Context) {
// 		id := c.GetParam("id")
// 		handle(expectedText + id)(c)
// 	})

// 	return func() {
// 		w := httptest.NewRecorder()
// 		req, _ := http.NewRequest(http.MethodGet, currentDefrouter+"/id2/"+now, nil)
// 		mux.ServeHTTP(w, req)
// 		body := w.Body.String()
// 		T.T.Logf("body: %s", body)
// 		T.Equal(expectedText+now, body)

// 		w = httptest.NewRecorder()
// 		req, _ = http.NewRequest(http.MethodGet, currentDefrouter+"/id2/t111111", nil)
// 		mux.ServeHTTP(w, req)
// 		body = w.Body.String()
// 		T.Equal("404", body)

// 	}
// }

// func testRouterPOST(mux *Engine, T *zlsgo.TestUtil) func() {
// 	expectedText := expected + "POST"
// 	mux.POST(defrouter, handle(expectedText))
// 	w := httptest.NewRecorder()
// 	return func() {
// 		req, _ := http.NewRequest("POST", defrouter, nil)
// 		mux.ServeHTTP(w, req)
// 		T.Equal(expectedText, w.Body.String())
// 	}
// }

// func testRouterPUT(mux *Engine, T *zlsgo.TestUtil) func() {
// 	expectedText := expected + "PUT"
// 	mux.PUT(defrouter, handle(expectedText))
// 	w := httptest.NewRecorder()
// 	return func() {
// 		req, _ := http.NewRequest("PUT", defrouter, nil)
// 		mux.ServeHTTP(w, req)
// 		T.Equal(expectedText, w.Body.String())
// 	}
// }

// func testRouterDELETE(mux *Engine, T *zlsgo.TestUtil) func() {
// 	expectedText := expected + "DELETE"
// 	mux.DELETE(defrouter, handle(expectedText))
// 	w := httptest.NewRecorder()
// 	return func() {
// 		req, _ := http.NewRequest("DELETE", defrouter, nil)
// 		mux.ServeHTTP(w, req)
// 		T.Equal(expectedText, w.Body.String())
// 	}
// }

// func testRouterPATCH(mux *Engine, T *zlsgo.TestUtil) func() {
// 	expectedText := expected + "PATCH"
// 	mux.PATCH(defrouter, handle(expectedText))
// 	w := httptest.NewRecorder()
// 	return func() {
// 		req, _ := http.NewRequest("PATCH", defrouter, nil)
// 		mux.ServeHTTP(w, req)
// 		T.Equal(expectedText, w.Body.String())
// 	}
// }

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

// func testRouterMatch(mux *Engine, T *zlsgo.TestUtil) func() {
// 	requestUrl := "/id/1/id2/2"
// 	ok := mux.Match(requestUrl, "/id/:id1/id2/:id2")

// 	if !ok {
// 		T.T.Fatal("TestRouter_Match test fail")
// 	}

// 	errorRequestUrl := "#xx#1#oo#2"
// 	ok = mux.Match(errorRequestUrl, "/xx/:param1/oo/:param2")

// 	if ok {
// 		T.T.Fatal("TestRouter_Match test fail")
// 	}

// 	ok = mux.Match(requestUrl, errorRequestUrl)

// 	if ok {
// 		T.T.Fatal("TestRouter_Match test fail")
// 	}

// 	return func() {

// 	}
// }
// func testRouterGroup(mux *Engine, T *zlsgo.TestUtil) func() {
// 	var w *httptest.ResponseRecorder
// 	expectedText := expected + "Group"
// 	expectedTextPOST := expectedText + "POST"
// 	prefix := "app"
// 	group := mux.Group(prefix)
// 	group.GET("", handle(expectedText))
// 	group.GET(defrouter, handle(expectedText))
// 	group.POST(defrouter+"/", handle(expectedTextPOST))

// 	prefix2 := "/app"
// 	expectedText2 := expected + "Group2"
// 	group2 := mux.Group("/app")
// 	group2.GET("", handle(expectedText2))

// 	return func() {
// 		req, _ := http.NewRequest("GET", prefix, nil)
// 		w = httptest.NewRecorder()
// 		mux.ServeHTTP(w, req)
// 		T.Equal(expectedText2, w.Body.String())

// 		req, _ = http.NewRequest("GET", prefix+"/"+defrouter, nil)
// 		w = httptest.NewRecorder()
// 		mux.ServeHTTP(w, req)
// 		T.Equal(expectedText, w.Body.String())

// 		req, _ = http.NewRequest("POST", prefix+"/"+defrouter+"/", nil)
// 		w = httptest.NewRecorder()
// 		mux.ServeHTTP(w, req)
// 		T.Equal(expectedTextPOST, w.Body.String())

// 		req, _ = http.NewRequest("POST", prefix2+"/"+defrouter+"/", nil)
// 		w = httptest.NewRecorder()
// 		mux.ServeHTTP(w, req)
// 		T.Equal(expectedTextPOST, w.Body.String())
// 	}
// }

// func withUse(next HandlerFunc) HandlerFunc {
// 	return func(c *Context) {
// 		next(c)
// 	}
// }

// func testRouterUse(mux *Engine, T *zlsgo.TestUtil) func() {
// 	mux.Use(withUse)
// 	return func() {}
// }

func handleRes(expected string) func(c *Context) {
	return func(c *Context) {
		fmt.Fprint(c.Writer, expected)
	}
}
