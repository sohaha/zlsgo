package zhttp_test

import (
    "encoding/json"
    "io"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/sohaha/zlsgo"
    "github.com/sohaha/zlsgo/zhttp"
)

func TestFluent(t *testing.T) {
    tt := zlsgo.NewTest(t)

    t.Run("GET with Query", func(t *testing.T) {
        ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            _, _ = w.Write([]byte(r.URL.Query().Get("id")))
        }))
        defer ts.Close()

        h := zhttp.NewRequest().Query("id", 123)
        res, err := h.URL(ts.URL).GET()
        tt.NoError(err)
        tt.Equal("123", res.String())
        tt.Equal("GET", res.Request().Method)
    })

    t.Run("POST form", func(t *testing.T) {
        ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            _ = r.ParseForm()
            ct := r.Header.Get("Content-Type")
            if !strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
                http.Error(w, "bad content type", http.StatusBadRequest)
                return
            }
            _, _ = w.Write([]byte(r.PostForm.Get("name")))
        }))
        defer ts.Close()

        res, err := zhttp.NewRequest().URL(ts.URL).Form("name", "alice").POST()
        tt.NoError(err)
        tt.Equal("alice", res.String())
        tt.Equal("POST", res.Request().Method)
    })

    t.Run("POST JSON", func(t *testing.T) {
        ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            body, _ := io.ReadAll(r.Body)
            _ = r.Body.Close()
            ct := r.Header.Get("Content-Type")
            if !strings.HasPrefix(ct, "application/json") {
                http.Error(w, "bad content type", http.StatusBadRequest)
                return
            }
            _, _ = w.Write(body)
        }))
        defer ts.Close()

        payload := map[string]string{"name": "bob"}
        res, err := zhttp.NewRequest().URL(ts.URL).JSON(payload).POST()
        tt.NoError(err)
        var got map[string]string
        _ = json.Unmarshal([]byte(res.String()), &got)
        tt.Equal("bob", got["name"])
        tt.Equal("POST", res.Request().Method)
    })

    t.Run("default method via Do() is GET", func(t *testing.T) {
        ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            _, _ = w.Write([]byte(r.Method))
        }))
        defer ts.Close()

        res, err := zhttp.NewRequest().URL(ts.URL).Do()
        tt.NoError(err)
        tt.Equal("GET", res.String())
    })

    t.Run("error when URL is empty", func(t *testing.T) {
        res, err := zhttp.NewRequest().GET()
        tt.Equal((*zhttp.Res)(nil), res)
        tt.Equal(zhttp.ErrUrlNotSpecified, err)
    })

    t.Run("Header and Reset", func(t *testing.T) {
        ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            v := r.Header.Get("X-Test") + ":" + r.URL.RawQuery
            _, _ = w.Write([]byte(v))
        }))
        defer ts.Close()

        req := zhttp.NewRequest().Header("X-Test", "v1").Query("a", 1)
        // after reset, previous state should be cleared
        req.Reset()
        res, err := req.URL(ts.URL).Header("X-Test", "v2").GET()
        tt.NoError(err)
        tt.Equal("v2:", res.String())
    })
}
