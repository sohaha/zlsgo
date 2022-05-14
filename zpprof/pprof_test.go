package zpprof

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/znet"
)

func TestListenAndServe(t *testing.T) {
	_ = ListenAndServe("127.0.0.1:67890")
}

func TestRegister(t *testing.T) {
	tt := zlsgo.NewTest(t)
	r := znet.New("pprof-test")
	r.SetMode(znet.DebugMode)
	Register(r, "666")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/debug?token=666", strings.NewReader(""))
	r.ServeHTTP(w, req)
	t.Log(w.Body.String())
	tt.Equal(200, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/debug", nil)
	r.ServeHTTP(w, req)
	t.Log(w.Body.String())
	tt.Equal(401, w.Code)
}
