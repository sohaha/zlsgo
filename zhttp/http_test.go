package zhttp

import (
	zls "github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/znet"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

var (
	r   *znet.Engine
	one sync.Once
)

func TestHttp(T *testing.T) {
	t := zls.NewTest(T)
	g := get(t)

	queryHandler := func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		t.Log(query)
	}
	ts := httptest.NewServer(http.HandlerFunc(queryHandler))
	t.Log(ts.URL)
	docs, err := Get(ts.URL)
	t.Log(docs, err)
	t.Equal(len(g) > 0, true)
}

func get(t *zls.TestUtil) string {
	docs, err := Get("https://docs.73zls.com/zls-go/#/")
	if err != nil {
		t.Log(err)
		return ""
	}
	return docs.String()
}
