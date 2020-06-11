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
	"github.com/sohaha/zlsgo/zvalid"

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
		engine.PreHandle(func(context *Context) (stop bool) {
			return
		})
		CloseHotRestartFileMd5()
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

func TestWeb(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := newServer()

	_, ok := Server("Web-test")
	tt.EqualExit(true, ok)

	// r.SetMode(ReleaseMode)
	w := newRequest(r, "GET", "/", "/", func(c *Context) {
		t.Log("TestWeb")
		_, _ = c.GetDataRaw()
		c.String(200, expected)
	})
	tt.Equal(200, w.Code)
	tt.Equal(expected, w.Body.String())
	r.GetMiddleware()
}

func TestFile(tt *testing.T) {
	T := zlsgo.NewTest(tt)
	r := newServer()
	w := newRequest(r, "GET", "/TestFile", "/TestFile", func(c *Context) {
		c.Next()
		tt.Log("PrevContent", c.PrevContent())
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
		tt.Log("PrevContent", zstring.Bytes2String(c.PrevContent().Content))
		T.Equal(expected, zjson.Get(zstring.Bytes2String(c.PrevContent().Content), "msg").String())
		tt.Log("PrevContent2", zstring.Bytes2String(c.PrevContent().Content))
		tt.Log("PrevStatus", c.PrevContent().Code)
		c.SetStatus(211)
		c.JSON(211, &ApiData{
			Code: 0,
			Msg:  "replace",
			Data: nil,
		})
		tt.Log("PrevContent3", zstring.Bytes2String(c.PrevContent().Content))
		tt.Log(c.Value("k1"))
		tt.Log(c.Value("k2"))
		tt.Log(c.Value("k2-2"))
		tt.Log(c.Value("k3"))
		tt.Log(c.Value("k4"))
	}, func(c *Context) {
		c.WithValue("k2", "k2-data")
		tt.Log("==2==")
		c.Next()
		p := c.PrevContent()
		ctype := p.Type
		tt.Log("PrevContentType", ctype)
		c.WithValue("k2-2", "k2-2-data")
	}, func(c *Context) {
		tt.Log("TestWeb")
		c.WithValue("k3", "k3-data")
		_, _ = c.GetDataRaw()
		c.JSON(201, ApiData{
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

func TestCustomMethod(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := newServer()
	r.SetCustomMethodField("_m_")

	r.PUT("/CustomMethod", func(c *Context) {
		tt.EqualExit(true, c.IsAjax())
		c.String(200, `put`)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/CustomMethod", nil)

	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("_m_", "put")
	req.Host = host

	r.ServeHTTP(w, req)
	tt.Equal(200, w.Code)
	tt.Equal("put", w.Body.String())
}

func TestHTML(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := newServer()
	w := newRequest(r, "GET", "/TestHTML", "/TestHTML", func(c *Context) {
		t.Log("TestHTML")
		c.HTML(202, `<html>123</html>`)
	})
	tt.Equal(202, w.Code)
	tt.EqualExit(`<html>123</html>`, w.Body.String())

	w = newRequest(r, "GET", "/TestHTML2", "/TestHTML2", func(c *Context) {
		t.Log("TestHTML2")
		c.Template(202, `<html>{{.title}}</html>`, Data{"title": "ZlsGo"})
	})
	tt.Equal(202, w.Code)
	tt.EqualExit(`<html>ZlsGo</html>`, w.Body.String())
}

func TestMore(t *testing.T) {
	T := zlsgo.NewTest(t)
	r := newServer()
	r.SetMode(DebugMode)
	ShutdownDone = func() {

	}
	CloseHotRestart = true
	w := newRequest(r, "delete", []string{"/", `{"Name":"is json"}`, ContentTypeJSON}, "/", func(c *Context) {
		_, _ = c.GetDataRaw()
		c.String(200, expected)
		c.GetAllQuerystMaps()
		c.GetAllQueryst()
		c.Log.Debug(c.GetJSON("Name"))
		var u struct {
			Name string `json:"Name"`
		}
		err := c.Bind(&u)
		T.EqualNil(err)
		c.Log.Warn(u)
	})
	T.Equal(200, w.Code)
	T.Equal(expected, w.Body.String())
}

func TestShouldBind(T *testing.T) {
	t := zlsgo.NewTest(T)
	r := newServer()
	w := newRequest(r, "POST", []string{"/?id=666666666&c=999",
		"d[C][Cc]=88&d.z=566&arr=1&arr=2&d[a]=1&d[A]=1a&a=123a&b=abc&name=seekwe" +
			"&s=true" +
			"&Abc=123&d[b]=9", "application/x-www-form-urlencoded"}, "/", func(c *Context) {
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
			d1  int
			Arr struct {
				A string `json:"a"`
				B string `json:"b"`
				C int    `json:"C"`
				Z string `json:"A"`
			} `json:"d"`
			Arr2   []string `json:"arr"`
			Status bool     `json:"s"`
			Name   string   `json:"name"`
		}{
			d1:   99,
			Name: "是我",
		}
		err = c.Bind(&ss)
		T.Log(fmt.Sprintf("%v", ss), err)
		t.EqualExit(123, ss.Abc)
		t.EqualExit(88, ss.Arr.C)
		t.EqualExit("1a", ss.Arr.A)
		t.EqualExit([]string{"1", "2"}, ss.Arr2)
		t.EqualExit(true, ss.Status)
		T.Log(ss.Arr)

		var bind struct {
			Name string `json:"name"`
			Abc  int
			ID   uint   `json:"id"`
			A    string `json:"a"`
		}

		rule := c.ValidRule().Required()
		err = c.BindValid(&bind, map[string]zvalid.Engine{
			"name": rule,
			"Abc":  rule.IsNumber(),
		})
		T.Log(bind, err)
		t.EqualNil(err)
		t.Equal(bind.Name, "seekwe")

		var bind2 struct {
			Name string
			Abc  int
			A    string
		}

		err = c.BindValid(&bind2, map[string]zvalid.Engine{
			"name": rule.SetAlias("用户名"),
			"Abc":  rule.IsNumber(),
		})
		T.Log(bind2, err)
		// t.EqualNil(err)
		t.Equal(bind2.Name, "")

		var bind3 struct {
			Name string
			Abc  int
			A    string
		}

		c.Engine.BindTag = ""
		err = c.BindValid(&bind3, map[string]zvalid.Engine{
			"name": rule,
			"Abc":  rule.SetAlias("Abc").IsNumber(),
		})
		T.Log(bind3, err)
		// t.EqualNil(err)
		t.Equal(bind3.Name, "seekwe")

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
		u := c.Host(true)
		t.Equal(true, host != u)
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
	r.NotFoundHandler(handleRes(expectedText))

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
		t.EqualExit(false, c.IsAjax())
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
	r.GetPanicHandler()
}

func TestRecovery(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := New()
	r.Use(Recovery(r, func(c *Context, err error) {
		c.String(200, "ok")
	}))
	r.GET("/", func(c *Context) {
		panic("xxx")
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)
	tt.Equal("ok", w.Body.String())
	tt.Equal(200, w.Code)
}
