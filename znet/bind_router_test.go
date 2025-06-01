package znet

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/ztype"
)

type testErrController struct{}

func (t *testErrController) Init(e *Engine) error {
	return errors.New("test error")
}

type testController struct{}

func (t *testController) Init(e *Engine) {
	e.Log.Debug("initialization")
	_ = RegisterRender((invokerCodeText)(nil), (*CustomInvoker)(nil))
}

func (t *testController) GETUser(_ *Context) {
}

func (t *testController) GETGetUser(_ *Context) (b []byte) {
	return []byte("GETUser")
}

func (t *testController) POSTUserInfo(_ *Context) {
}

func (t *testController) PUTUserInfo(_ *Context) {
}

func (t *testController) DELETEUserInfo(_ *Context) {
}

func (t *testController) PATCHUserInfo(_ *Context) {
}

func (t *testController) HEADUserInfo(_ *Context) {
}

func (t *testController) OPTIONSUserInfo(_ *Context) {
}

func (t *testController) AnyOk(c *Context) error {
	fmt.Println(c.Request.Method)
	return errors.New("ok")
}

func (t *testController) IDGET(_ *Context) {
}

func (t *testController) IDGETUser(_ *Context) {
}

func (t *testController) FullGETFile(_ *Context) {
}

func TestBindStruct(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := newServer()
	r.PanicHandler(func(c *Context, err error) {
		t.Log("PanicHandler", err)
	})
	prefix := "/test"
	err := r.BindStruct(prefix, &testErrController{})
	tt.Equal("test error", err.Error())

	err = r.BindStruct(prefix, &testController{}, func(c *Context) {
		t.Log("go", c.Request.URL)
		t.Log(c.GetAllParam())
		c.Next()
	})
	t.Log(err)
	tt.EqualNil(err)
	r.BindStructDelimiter = ""
	r.BindStructSuffix = ".go"
	err = r.BindStruct(prefix, &testController{}, func(c *Context) {
		t.Log("go", c.Request.URL)
		t.Log(c.GetAllParam())
		c.Next()
	})
	tt.Log(err)
	tt.EqualNil(err)
	methods := [][]string{
		{"GET", prefix + "/user"},
		{"GET", prefix + "/get-user"},
		{"POST", prefix + "/ok", "500"},
		{"POST", prefix + "/user-info"},
		{"PUT", prefix + "/user-info"},
		{"DELETE", prefix + "/user-info"},
		{"PATCH", prefix + "/user-info"},
		{"OPTIONS", prefix + "/user-info"},
		{"POST", prefix + "/UserInfo.go"},
		{"GET", prefix + "/user/233"},
		{"GET", prefix + "/User/233"},
		{"GET", prefix + "/File/File233"},
		{"GET", prefix + "/file/File233"},
		{"GET", prefix + "/ok", "500"},
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
	}

	err = r.BindStruct(prefix, nil)
	tt.Log(err)
}

func TestBindStructCase(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := newServer()
	r.BindStructCase = func(s string) string {
		tt.Log(s)
		if s == "UserInfo" {
			return "new-user-info"
		}
		return s
	}
	err := r.BindStruct("BindStructCase", &testController{}, func(c *Context) {
		t.Log("go", c.Request.URL)
		t.Log(c.GetAllParam())
		c.Next()
	})
	tt.NoError(err)

	methods := [][]string{
		{"POST", "/BindStructCase/new-user-info.go"},
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
	}
}
