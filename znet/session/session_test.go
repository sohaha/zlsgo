package session_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/znet/session"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
)

const host = "127.0.0.1"

var (
	r    *znet.Engine
	size = 10000
)

func init() {
	r = znet.New()
	r.SetMode(znet.ProdMode)

	r.Use(session.Default())
	r.GET("/session", func(c *znet.Context, s session.Session) {
		rand := s.Get("rand")
		if !rand.Exists() || ztype.ToBool(c.DefaultQuery("reset", "")) {
			r := zstring.Rand(6)
			s.Set("rand", r)
			rand = ztype.New(r)
			s.Save()
		}

		c.String(200, rand.String())
	})
	r.GET("/session2", func(c *znet.Context) error {
		s, err := session.Get(c)
		if err != nil {
			return err
		}
		c.String(204, s.Get("rand").String())
		return nil
	})
}

func TestSession(t *testing.T) {
	tt := zlsgo.NewTest(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/session", nil)
	r.ServeHTTP(w, req)
	s := getSession(w)
	rand := w.Body.String()
	tt.Equal(200, w.Code, true)
	tt.Log(s.Value, rand)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/session", nil)
	req.AddCookie(s)
	r.ServeHTTP(w, req)
	tt.Equal(200, w.Code, true)
	tt.Log(s.Value, w.Body.String())
	tt.Equal(rand, w.Body.String())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/session", nil)
	req.AddCookie(s)
	r.ServeHTTP(w, req)
	tt.Equal(200, w.Code, true)
	tt.Log(s.Value, w.Body.String())
	tt.Equal(rand, w.Body.String())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/session", nil)
	r.ServeHTTP(w, req)
	tt.Equal(200, w.Code, true)
	tt.Log(s.Value, w.Body.String())
	tt.EqualTrue(rand != w.Body.String())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/session", nil)
	req.AddCookie(s)
	r.ServeHTTP(w, req)
	tt.Equal(200, w.Code, true)
	tt.Log(s.Value, w.Body.String())
	tt.Equal(rand, w.Body.String())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/session?reset=true", nil)
	req.AddCookie(s)
	r.ServeHTTP(w, req)
	tt.Equal(200, w.Code, true)
	tt.Log(s.Value, w.Body.String())
	newRand := w.Body.String()
	tt.EqualTrue(rand != newRand)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/session", nil)
	req.AddCookie(s)
	r.ServeHTTP(w, req)
	tt.Equal(200, w.Code, true)
	tt.Log(s.Value, w.Body.String())
	tt.Equal(newRand, w.Body.String())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/session2", nil)
	req.AddCookie(s)
	r.ServeHTTP(w, req)
	tt.Equal(204, w.Code, true)
	tt.Log(s.Value, w.Body.String())
	tt.Equal(newRand, w.Body.String())
}

func getSession(w *httptest.ResponseRecorder) *http.Cookie {
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == "session_id" {
			return cookie
		}
	}
	return &http.Cookie{}
}
