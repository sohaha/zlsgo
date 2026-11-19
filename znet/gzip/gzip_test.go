package gzip_test

import (
	"bytes"
	stdgzip "compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/znet/gzip"
	"github.com/sohaha/zlsgo/zstring"
)

const host = "127.0.0.1"

var (
	r    *znet.Engine
	size = 10000
)

// minCompressedSize is the lower bound the compressed benchmark body must
// exceed. The payload is 10000 characters drawn from a 3-letter alphabet, so
// gzip still emits a few hundred bytes; anything at or below this means the
// response was not really compressed. TestGzip asserts the same lower bound.
const minCompressedSize = 100

func init() {
	r = znet.New()
	r.SetMode(znet.ProdMode)

	r.GET("/gzip", func(c *znet.Context) {
		c.String(200, zstring.Rand(size, "abc"))
	}, gzip.Default())

}

func TestGzip(t *testing.T) {
	tt := zlsgo.NewTest(t)

	go func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/gzip", nil)
		r.ServeHTTP(w, req)
		tt.Equal(200, w.Code)
		tt.Equal(size, w.Body.Len())
	}()

	var g sync.WaitGroup
	for i := 0; i < 1000; i++ {
		g.Add(1)
		go func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/gzip", nil)
			req.Header.Add("Accept-Encoding", "gzip")
			req.Host = host
			r.ServeHTTP(w, req)
			tt.Equal(200, w.Code)
			tt.EqualTrue(w.Body.Len() > 100)
			tt.EqualTrue(w.Body.Len() < size)
			g.Done()
		}()
	}
	g.Wait()
}

func TestGzipReadable(t *testing.T) {
	tt := zlsgo.NewTest(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/gzip", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	req.Host = host
	r.ServeHTTP(w, req)

	tt.Equal(200, w.Code)

	gr, err := stdgzip.NewReader(bytes.NewReader(w.Body.Bytes()))
	if !tt.NoError(err, true) {
		return
	}
	defer gr.Close()

	data, err := io.ReadAll(gr)
	tt.NoError(err, true)
	tt.Equal(size, len(data))
}

func BenchmarkGzipDoNotUse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/gzip", nil)
		req.Header.Add("Accept-Encoding1", "not-gzip")
		r.ServeHTTP(w, req)
		if w.Code != 200 || size != w.Body.Len() {
			b.Fail()
		}
	}
}

func BenchmarkGzipUse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/gzip", nil)
		req.Header.Add("Accept-Encoding", "gzip")
		req.Host = host
		r.ServeHTTP(w, req)
		if w.Code != 200 || size <= w.Body.Len() || w.Body.Len() <= minCompressedSize {
			b.Fail()
		}
	}
}
