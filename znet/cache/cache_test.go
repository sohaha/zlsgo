package cache_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/znet/cache"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zsync"
)

var (
	r *znet.Engine
)

func init() {
	r = znet.New()
	r.SetMode(znet.ProdMode)

	r.GET("/cache", func(c *znet.Context) {
		c.String(200, zstring.Rand(10))
	}, cache.New(func(conf *cache.Config) {
		conf.Expiration = 10 * time.Second
		conf.Custom = func(c *znet.Context) (key string, expiration time.Duration) {
			return cache.QueryKey(c), 0
		}
	}))
}

func qvalue() string {
	q := map[string]string{
		"b": "2",
		"a": "1",
		"c": "3",
	}

	val := ""
	for k, v := range q {
		val += k + "=" + v + "&"
	}
	return val
}

func TestCache(t *testing.T) {
	tt := zlsgo.NewTest(t)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/cache?"+qvalue(), nil)
	r.ServeHTTP(w, req)
	str := w.Body

	var wg zsync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Go(func() {
			w := httptest.NewRecorder()
			url := "/cache?" + qvalue()
			req, _ := http.NewRequest("GET", url, nil)
			r.ServeHTTP(w, req)
			tt.Equal(200, w.Code)
			tt.Equal(10, w.Body.Len())
			tt.Equal(str, w.Body)
		})
	}
	_ = wg.Wait()
}
