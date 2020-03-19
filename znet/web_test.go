package znet

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zstring"

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

func newRequest(r *Engine, method string, urlAndBody interface{}, path string, handler ...HandlerFunc) *httptest.ResponseRecorder {
	var (
		body        io.Reader
		_url        string
		contentType string
	)
	method = strings.ToUpper(method)
	if u, ok := urlAndBody.(string); ok {
		_url = u
	} else if u, ok := urlAndBody.([]string); ok {
		_url = u[0]
		body = strings.NewReader(u[1])
		contentType = u[2]
	}
	if len(handler) > 0 {
		firstHandler := handler[0]
		handlers := handler[1:]
		if path == "" {
			path = _url
		}
		switch method {
		case "GET":
			r.GET(path, firstHandler, handlers...)
		case "POST":
			r.POST(path, firstHandler, handlers...)
		default:
			r.Customize(method, path, firstHandler, handlers...)
		}
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, _url, body)
	req.Host = host
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	r.ServeHTTP(w, req)
	return w
}

func TestWeb(tt *testing.T) {
	T := zlsgo.NewTest(tt)
	r := newServer()
	// r.SetMode(ReleaseMode)
	w := newRequest(r, "GET", "/", "/", func(c *Context) {
		tt.Log("TestWeb")
		_, _ = c.GetDataRaw()
		c.String(200, expected)
	})
	T.Equal(200, w.Code)
	T.Equal(expected, w.Body.String())
	r.GetMiddleware()
}

func TestFile(tt *testing.T) {
	T := zlsgo.NewTest(tt)
	r := newServer()
	w := newRequest(r, "GET", "/TestFile", "/TestFile", func(c *Context) {
		c.Next()
		tt.Log("PrevContent", c.PrevContent())
		tt.Log("PrevStatus", c.PrevStatus())
	}, func(c *Context) {
		tt.Log("TestFile")
		c.File("doc.go")
	})
	T.Equal(200, w.Code)
	tt.Log(len(w.Body.String()))

	w = newRequest(r, "GET", "/TestFile2", "/TestFile2", func(c *Context) {
		tt.Log("TestFile")
		c.File("doc_no.go")
	})
	T.Equal(404, w.Code)
	tt.Log(len(w.Body.String()))

	w = newRequest(r, "GET", "/TestFile3", "/TestFile3", func(c *Context) {
		c.Next()
		tt.Log("PrevContent", c.PrevContent())
		tt.Log("PrevStatus", c.PrevStatus())
		c.String(211, "file")
	}, func(c *Context) {
		tt.Log("TestFile")
		c.File("doc_no.go")
	})
	T.Equal(211, w.Code)
	tt.Log(len(w.Body.String()))
}

func TestPost(tt *testing.T) {
	T := zlsgo.NewTest(tt)
	r := newServer()
	r.SetMode(DebugMode)
	w := newRequest(r, "POST", "/Post", "/Post", func(c *Context) {
		c.WithValue("k1", "k1-data")
		tt.Log("==1==")
		c.Next()
		tt.Log("--1--")
		tt.Log("PrevContent", zstring.Bytes2String(c.PrevContent()))
		T.Equal(expected, zjson.Get(zstring.Bytes2String(c.PrevContent()), "msg").String())
		tt.Log("PrevContent2", zstring.Bytes2String(c.PrevContent()))
		tt.Log("PrevStatus", c.PrevStatus())
		c.SetStatus(211)
		c.JSON(211, &Api{
			Code: 0,
			Msg:  "replace",
			Data: nil,
		})
		tt.Log("PrevContent3", zstring.Bytes2String(c.PrevContent()))
		tt.Log(c.Value("k1"))
		tt.Log(c.Value("k2"))
		tt.Log(c.Value("k2-2"))
		tt.Log(c.Value("k3"))
		tt.Log(c.Value("k4"))
	}, func(c *Context) {
		c.WithValue("k2", "k2-data")
		tt.Log("==2==")
		c.Next()
		tt.Log("GetContentType", c.GetContentType())
		c.WithValue("k2-2", "k2-2-data")
	}, func(c *Context) {
		tt.Log("TestWeb")
		c.WithValue("k3", "k3-data")
		_, _ = c.GetDataRaw()
		c.JSON(201, Api{
			Code: 200,
			Msg:  expected,
			Data: nil,
		})
	})
	T.Equal(211, w.Code)
	T.Equal("replace", zjson.Get(w.Body.String(), "msg").String())

	w = newRequest(r, "POST", "/Post2", "/Post2",
		func(c *Context) {
			c.Abort(222)
		}, func(c *Context) {
			c.String(200, "ok")
		})
	T.Equal(222, w.Code)

	w = newRequest(r, "POST", "/Post3", "/Post3",
		func(c *Context) {
			c.Byte(200, []byte("ok"))
		})
	T.Equal(200, w.Code)
	T.Equal("ok", w.Body.String())
}

func TestHTML(tt *testing.T) {
	T := zlsgo.NewTest(tt)
	r := newServer()
	w := newRequest(r, "GET", "/TestHTML", "/TestHTML", func(c *Context) {
		tt.Log("TestHTML")
		c.HTML(202, `<html>123</html>`)
	})
	T.Equal(202, w.Code)
	T.EqualExit(`<html>123</html>`, w.Body.String())

	w = newRequest(r, "GET", "/TestHTML2", "/TestHTML2", func(c *Context) {
		tt.Log("TestHTML2")
		c.Template(202, `<html>{{.title}}</html>`, Data{"title": "ZlsGo"})
	})
	T.Equal(202, w.Code)
	T.EqualExit(`<html>ZlsGo</html>`, w.Body.String())
}

func TestMore(t *testing.T) {
	T := zlsgo.NewTest(t)
	r := newServer()
	r.SetMode(DebugMode)
	ShutdownDone = func() {

	}
	CloseHotRestart = true
	w := newRequest(r, "delete", "/", "/", func(c *Context) {
		_, _ = c.GetDataRaw()
		c.String(200, expected)
	})
	T.Equal(200, w.Code)
	T.Equal(expected, w.Body.String())
}

func TestShouldBind(T *testing.T) {
	t := zlsgo.NewTest(T)
	r := newServer()
	w := newRequest(r, "POST", []string{"/?c=999", "d[C][Cc]=88&arr=1&arr=2&d[a]=1&a=123&b=abc&name=seekwe&s=true&Abc=123&d[b]=9", "application/x-www-form-urlencoded"}, "/", func(c *Context) {
		ct := c.ContentType()
		t.Equal(MIMEPOSTForm, ct)
		_ = c.DefaultFormOrQuery("ok", "yes")
		all, err := c.GetPostFormAll()
		T.Log(all, err)
		r, e := c.GetDataRaw()
		T.Log(r, e)
		r, _ = c.GetPostForm("b")
		d, _ := c.GetPostFormMap("d")
		T.Log(d)
		arr2, _ := c.GetPostFormArray("arr")
		T.Log("arr2", arr2)
		t.EqualExit("abc", r)
		r, _ = c.GetQuery("c")
		t.EqualExit("999", r)

		ss := struct {
			Abc int
			d   int
			Arr struct {
				A int    `z:"a"`
				B string `z:"b"`
				C struct {
					Cc string `z:"cc"`
				} `z:"C"`
			} `z:"d"`
			Arr2   []string `z:"arr"`
			Status bool     `z:"s"`
			Name   string   `z:"name"`
		}{
			d:    99,
			Name: "是我",
		}
		err = c.Bind(&ss)
		T.Log(fmt.Sprintf("%v", ss), err)
		// err = Request(c.Request).Field(&ss,"")
		// t.Log(1, ss, err)
		c.String(210, expected)
	})
	t.EqualExit(210, w.Code)
	t.EqualExit(expected, w.Body.String())
}

func TestSetMode(T *testing.T) {
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
