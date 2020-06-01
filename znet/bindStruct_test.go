package znet

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sohaha/zlsgo"
)

type testController struct {
}

func (t *testController) GetUser(_ *Context) {

}

func (t *testController) GetgetUser(_ *Context) {

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

func (t *testController) AnyOk(_ *Context) {

}

func (t *testController) No(_ *Context) {

}

func TestBindStruct(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := newServer()
	prefix := "/test"
	err := r.BindStruct(prefix, &testController{}, func(c *Context) {
		t.Log("go", c.Request.URL)
		c.Next()
	})
	tt.Log(err)
	tt.EqualNil(err)
	BindStructSuffix = ".go"
	err = r.BindStruct(prefix, &testController{}, func(c *Context) {
		t.Log("go", c.Request.URL)
		c.Next()
	})
	tt.Log(err)
	tt.EqualNil(err)
	methods := [][]string{
		{"GET", prefix + "/user"},
		{"GET", prefix + "/get-user"},
		{"POST", prefix + "/ok"},
		{"POST", prefix + "/user-info"},
		{"PUT", prefix + "/user-info"},
		{"DELETE", prefix + "/user-info"},
		{"PATCH", prefix + "/user-info"},
		{"OPTIONS", prefix + "/user-info"},
	}
	for _, v := range methods {
		w := httptest.NewRecorder()
		t.Log("Test:", v[0], v[1])
		req, _ := http.NewRequest(v[0], v[1], nil)
		r.ServeHTTP(w, req)
		tt.Equal(200, w.Code)
	}

	err = r.BindStruct(prefix, nil)
	tt.Log(err)
}
