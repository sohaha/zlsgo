package znet

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zlog"
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
		c.String(211, "test InjectErrRewrite")
	}))
	tt.Equal(211, w.Code)
	tt.Equal("test InjectErrRewrite", w.Body.String())
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

	w := newRequest(r, "GET", "/TestInjectMiddleware", "/TestInjectMiddleware", func() (int, string) {
		t.Log("run")
		return 403, "Inject"
	}, Recovery(func(c *Context, err error) {
		zlog.Error("Recovery", err)
	}), func(c *Context) {
		c.Next()
	}, func() error {
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
		err := c.Injector.Resolve(&s)
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
		c.Injector.Map("test")
		c.Next()
	})
	tt.Equal(403, w.Code)
	tt.Equal([]int{1, 2, 3, 4, 5, 6, 7, 8, 9}, pc)
	tt.Equal("middleware", w.Body.String())
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
