package auth_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/znet/auth"
	"github.com/sohaha/zlsgo/zutil"
)

var authHandler = auth.New(auth.Accounts{
	"admin":  "123",
	"admin2": "456",
})

func TestNew(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := server().(*znet.Engine)

	w1 := newRequest(r, "/auth1", nil)
	tt.Equal(http.StatusUnauthorized, w1.Code)

	w2 := newRequest(r, "/auth2", []string{"admin", "123"})
	tt.Equal(http.StatusOK, w2.Code)
	tt.Equal("admin", w2.Body.String())

	w3 := newRequest(r, "/auth3", []string{"admin2", "456"})
	tt.Equal(http.StatusOK, w3.Code)
	tt.Equal("admin2", w3.Body.String())

	w4 := newRequest(r, "/auth4", []string{"admin3", "456"})
	tt.Equal(http.StatusUnauthorized, w4.Code)
	tt.Equal("", w4.Body.String())

}

var server = zutil.Once(func() interface{} {
	r := znet.New()
	// r.SetMode(znet.DebugMode)
	return r
})

func newRequest(r *znet.Engine, path string, account []string) *httptest.ResponseRecorder {
	r.GET(path, func(c *znet.Context) {
		c.String(200, c.MustValue(auth.UserKey, "").(string))
	}, authHandler)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", path, nil)
	if len(account) == 2 {
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(account[0]+":"+account[1])))
	}
	r.ServeHTTP(w, req)
	return w
}
