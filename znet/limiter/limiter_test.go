package limiter_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/znet/limiter"
)

var (
	one    sync.Once
	engine *znet.Engine
)

func newServer() *znet.Engine {
	one.Do(func() {
		engine = znet.New("limiter_test")
		engine.AddAddr("3787")
	})
	return engine
}

func TestNew(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	r := newServer()

	rule := limiter.NewRule()
	rule.AddRule(time.Second, 3)
	rule.AddRule(time.Second*2, 4, 5)
	r.GET("/limiterCustomize", func(c *znet.Context) {
		c.String(200, "ok")
	}, func(c *znet.Context) {
		if !rule.AllowVisitByIP(c.GetClientIP()) {
			c.String(http.StatusTooManyRequests, "超过限制")
			c.Abort()
			return
		}
		c.Next()
	})

	r.GET("/limiterCustomizeUser", func(c *znet.Context) {
		c.String(200, "ok")
	}, func(c *znet.Context) {
		if !rule.AllowVisit("username") {
			c.Abort()
			c.String(http.StatusTooManyRequests, "超过限制")
			return
		}
		c.Next()
	})

	r.GET("/limiter", func(c *znet.Context) {
		c.String(200, "ok")
	}, limiter.New(3, func(c *znet.Context) {
		c.String(http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
	}))

	r.GET("/limiter2", func(c *znet.Context) {
		c.String(200, "ok")
	}, limiter.New(3))

	ii := run("/limiterCustomize", r)
	t.EqualExit(int64(3), ii)

	tt.Log(rule.Remaining("username"))
	ii = run("/limiterCustomizeUser", r)
	t.EqualExit(int64(3), ii)
	t.EqualExit([]int{0, 1}, rule.Remaining("username"))
	tt.Log(rule.GetOnline())

	ii = run("/limiter", r)
	t.EqualExit(int64(3), ii)

	ii = run("/limiter2", r)
	t.EqualExit(int64(3), ii)

	time.Sleep(time.Second)

	ii = run("/limiterCustomizeUser", r)
	t.EqualExit(int64(1), ii)
	t.EqualExit([]int{0, 0}, rule.Remaining("username"))

	time.Sleep(time.Second * 2)

	ii = run("/limiterCustomizeUser", r)
	t.EqualExit(int64(3), ii)
	t.EqualExit([]int{0, 1}, rule.Remaining("username"))
}

func run(url string, r *znet.Engine) int64 {
	var wg sync.WaitGroup
	var ii int64
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("X-Real-Ip", "192.168.1.1")
			r.ServeHTTP(w, req)
			if w.Code == http.StatusOK {
				atomic.AddInt64(&ii, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	return ii
}
