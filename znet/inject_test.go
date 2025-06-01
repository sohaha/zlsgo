package znet

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/ztype"
)

var _ zdi.PreInvoker = (*CustomInvoker)(nil)

type CustomInvoker func(ctx *Context) (b []byte)

func (fn CustomInvoker) Invoke(i []interface{}) ([]reflect.Value, error) {
	c := i[0].(*Context)
	b := fn(c)
	c.Byte(404, b)
	return nil, nil
}

func TestInject(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := newServer()

	r.Log.Discard()

	w := newRequest(r, "GET", "/NoInject", "/NoInject", func(c *Context) {
		c.String(200, "NoInject")
	})
	tt.Equal(200, w.Code)
	tt.Equal("NoInject", w.Body.String())

	w = newRequest(r, "GET", "/Inject", "/Inject", func() (int, string) {
		return 403, "Inject"
	})
	tt.Equal(403, w.Code)
	tt.Equal("Inject", w.Body.String())

	rewriteError := ""
	w = newRequest(r, "GET", "/InjectErr", "/InjectErr", func() (int, string, error) {
		return 403, "test InjectErr", errors.New("test InjectErr")
	})
	tt.Equal(500, w.Code)
	tt.Equal("test InjectErr", w.Body.String())
	tt.Equal("", rewriteError)
	w = newRequest(r, "GET", "/InjectErrRewrite", "/InjectErrRewrite", func() (int, string, error) {
		return 403, "test InjectErr", errors.New("InjectErrRewrite")
	}, RewriteErrorHandler(func(c *Context, err error) {
		tt.Equal("InjectErrRewrite", err.Error())
		rewriteError = err.Error()
	}))
	tt.Equal(403, w.Code)
	tt.Equal("test InjectErr", w.Body.String())
	tt.Equal("InjectErrRewrite", rewriteError)

	w = newRequest(r, "GET", "/InjectCustom", "/InjectCustom", CustomInvoker(func(ctx *Context) (b []byte) {
		return []byte("InjectCustom")
	}), func() {
	})
	tt.Equal(404, w.Code)
	tt.Equal("InjectCustom", w.Body.String())

	w = newRequest(r, "GET", "/InjectAny", "/InjectAny", func(ctx *Context) (c uint, api ApiData, err error) {
		return 302, ApiData{Code: 301, Msg: "InjectAny"}, nil
	})
	tt.Equal(302, w.Code)
	tt.Equal("application/json; charset=utf-8", w.Header().Get("Content-Type"))
	tt.Equal("InjectAny", zjson.Get(w.Body.String(), "msg").String())

	w = httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "__404__", nil)
	r.ServeHTTP(w, req)
	t.Log(w)
}

func TestInjectMiddleware(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := newServer()

	now := time.Now()
	r.Injector().Map(now)

	w := newRequest(r, "GET", "/TestInjectMiddleware", "/TestInjectMiddleware", func() (int, string) {
		t.Log("run")
		return 403, "Inject"
	}, Recovery(func(c *Context, err error) {
		zlog.Error("Recovery", err)
	}), func(c *Context) {
		c.Next()
	}, func(n time.Time) error {
		tt.Equal(now, n)
		return errors.New("return exit")
	})
	tt.Equal(500, w.Code)
	tt.Equal("return exit", w.Body.String())

	pc := make([]int, 0)
	w = newRequest(r, "GET", "/TestInjectMiddleware2", "/TestInjectMiddleware2", func() (int, string) {
		t.Log("run")
		return 403, "Inject"
	}, func(c *Context) {
		pc = append(pc, 1)
		c.Next()
		pc = append(pc, 9)
	}, func(c *Context) string {
		pc = append(pc, 2)
		c.Next()
		pc = append(pc, 8)
		return "middleware"
	}, func(c *Context) error {
		pc = append(pc, 3)
		c.Next()
		pc = append(pc, 7)
		var s string
		err := c.Injector().Resolve(&s)
		tt.NoError(err)
		tt.Equal("test", s)
		return nil
	}, func() {
		pc = append(pc, 4)
	}, func(c *Context) {
		c.Next()
		pc = append(pc, 6)
	}, func(c *Context) {
		pc = append(pc, 5)
		c.Injector().Map("test")
		c.Next()
	})
	tt.Equal(403, w.Code)
	tt.Equal([]int{1, 2, 3, 4, 5, 6, 7, 8, 9}, pc)
	tt.Equal("middleware", w.Body.String())
}

type customRenderer struct {
	Err  error
	Text string
}

func (c *customRenderer) Content(ctx *Context) (content []byte) {
	if c.Err != nil {
		ctx.SetStatus(500)
		return []byte(c.Err.Error())
	}
	ctx.SetStatus(200)
	return []byte(c.Text)
}

type Custom0 func(*Context) string

func (t Custom0) Invoke(args []interface{}) ([]reflect.Value, error) {
	c := args[0].(*Context)
	str := t(c)
	c.String(200, "[0]:"+str)
	return nil, nil
}

type custom1 func(*Context) string

func (t custom1) Invoke(args []interface{}) ([]reflect.Value, error) {
	c := args[0].(*Context)
	str := t(c)
	c.String(200, "[1]:"+str)
	return nil, nil
}

func TestCustomRenderer(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := newServer()

	w := newRequest(r, "GET", "/TestCustomRenderer", "/TestCustomRenderer", func(c *Context) *customRenderer {
		return &customRenderer{Text: "test custom renderer"}
	})
	tt.Equal(200, w.Code)
	tt.Equal("test custom renderer", w.Body.String())

	w = newRequest(r, "GET", "/TestCustomRendererError", "/TestCustomRendererError", func(c *Context) *customRenderer {
		return &customRenderer{Err: errors.New("test custom renderer error")}
	})
	tt.Equal(500, w.Code)
	tt.Equal("test custom renderer error", w.Body.String())

	RegisterRender(Custom0(nil))
	r.GET("/BindStructCustom_0/", func(c *Context) string {
		return "BindStructCustom_0"
	})

	r.Group("BindStructCustom_1", func(g *Engine) {
		g.RegisterRender((custom1)(nil))
		g.GET("/", func(c *Context) string {
			return "BindStructCustom_1"
		})
	})

	r.Group("BindStructCustom_2", func(g *Engine) {
		g.GET("/", func(c *Context) string {
			return "BindStructCustom_2"
		})
	})
	methods := [][]string{
		{"GET", "/BindStructCustom_0/"},
		{"GET", "/BindStructCustom_1/"},
		{"GET", "/BindStructCustom_2/"},
	}
	for _, v := range methods {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(v[0], v[1], nil)
		r.ServeHTTP(w, req)
		code := 200
		if len(v) > 2 {
			code = ztype.ToInt(v[2])
		}
		tt.Equal(code, w.Code)
		t.Log("Test:", v[0], v[1])
		t.Log(w.Code, w.Body.String())
		switch v[1] {
		case "/BindStructCustom_0/":
			tt.Equal("[0]:BindStructCustom_0", w.Body.String())
		case "/BindStructCustom_1/":
			tt.Equal("[1]:BindStructCustom_1", w.Body.String())
		case "/BindStructCustom_2/":
			tt.Equal("[0]:BindStructCustom_2", w.Body.String())
		}
	}
}

func BenchmarkInjectNo(b *testing.B) {
	r := newServer()
	path := "/BenchmarkInjectNo"
	r.SetMode(QuietMode)
	r.GET(path, func(c *Context) {
		c.String(200, path)
	})
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", path, nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 || w.Body.String() != path {
			b.Fail()
		}
	}
}

func BenchmarkInjectFast(b *testing.B) {
	r := newServer()
	r.SetMode(QuietMode)
	path := "/BenchmarkInjectFast"
	r.GET(path, func() (int, string) {
		return 200, path
	})
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", path, nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 || w.Body.String() != path {
			b.Fail()
		}
	}
}

func BenchmarkInjectBasis(b *testing.B) {
	r := newServer()
	r.SetMode(QuietMode)
	path := "/BenchmarkInjectBasis"
	r.GET(path, func() (int, []byte) {
		return 200, []byte(path)
	})
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", path, nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 || w.Body.String() != path {
			b.Fail()
		}
	}
}
