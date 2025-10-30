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

func TestRemainingVisitsByIP(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	t.Parallel()

	rule := limiter.NewRule()
	rule.AddRule(time.Second, 5)

	remaining := rule.RemainingVisitsByIP("192.168.1.1")
	t.EqualExit(1, len(remaining))
	t.EqualExit(5, remaining[0])

	remaining = rule.RemainingVisitsByIP("invalid.ip")
	t.EqualExit(0, len(remaining))

	remaining = rule.RemainingVisitsByIP("")
	t.EqualExit(0, len(remaining))
}

func TestAllowVisitEdgeCases(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	t.Parallel()

	rule := limiter.NewRule()
	t.EqualTrue(rule.AllowVisit("test"))
	t.EqualTrue(rule.AllowVisit("test", "test2"))

	rule.AddRule(time.Second, 2)
	rule.AddRule(time.Second*2, 3)

	t.EqualTrue(rule.AllowVisit("user1"))
	t.EqualTrue(rule.AllowVisit("user1"))
	t.EqualFalse(rule.AllowVisit("user1"))
}

func TestAllowVisitByIPEdgeCases(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	t.Parallel()

	rule := limiter.NewRule()
	rule.AddRule(time.Second, 2)

	t.EqualTrue(rule.AllowVisitByIP("192.168.1.1"))
	t.EqualTrue(rule.AllowVisitByIP("invalid.ip.address"))
	t.EqualTrue(rule.AllowVisitByIP(""))
}

func TestLimiterRecovery(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	t.Parallel()

	rule := limiter.NewRule()
	rule.AddRule(time.Millisecond*100, 1, 1)

	for i := 0; i < 10; i++ {
		key := i
		rule.AllowVisit(key)
	}

	time.Sleep(time.Millisecond * 200)

	t.EqualTrue(rule.AllowVisit("test"))
}

func TestLimiterAddMethodEdgeCases(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	t.Parallel()

	rule := limiter.NewRule()
	rule.AddRule(time.Second, 2, 1)

	t.EqualTrue(rule.AllowVisit("user1"))
	t.EqualTrue(rule.AllowVisit("user2"))

	for i := 0; i < 5; i++ {
		key := i
		t.EqualTrue(rule.AllowVisit(key))
	}
}

func TestNewRuleWithZeroAllowed(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	t.Parallel()

	rule := limiter.NewRule()
	rule.AddRule(time.Second, 0)

	t.EqualTrue(rule.AllowVisit("test"))
	t.EqualFalse(rule.AllowVisit("test"))
}

func TestNewWithMultipleOverflowHandlers(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	t.Parallel()

	customHandler := func(c *znet.Context) {
		c.String(429, "Custom rate limit exceeded")
	}

	limiter := limiter.New(2, customHandler)

	r := znet.New("test")
	r.GET("/test", func(c *znet.Context) {
		c.String(200, "ok")
	}, limiter)

	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Real-Ip", "10.0.0.1")
		r.ServeHTTP(w, req)
		t.EqualExit(200, w.Code)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-Ip", "10.0.0.1")
	r.ServeHTTP(w, req)
	t.EqualExit(429, w.Code)
	t.EqualExit("Custom rate limit exceeded", w.Body.String())
}

func TestConcurrentAccess(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	t.Parallel()

	rule := limiter.NewRule()
	rule.AddRule(time.Second, 100)

	var wg sync.WaitGroup
	var successCount int64
	var failCount int64

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rule.AllowVisitByIP("127.0.0.1") {
				atomic.AddInt64(&successCount, 1)
			} else {
				atomic.AddInt64(&failCount, 1)
			}
		}()
	}

	wg.Wait()

	t.EqualExit(int64(100), successCount)
	t.EqualExit(int64(900), failCount)
}

func TestRuleSorting(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	t.Parallel()

	rule := limiter.NewRule()
	rule.AddRule(time.Minute, 10)
	rule.AddRule(time.Second, 1)
	rule.AddRule(time.Hour, 100)

	remaining := rule.Remaining("test")
	t.EqualExit(3, len(remaining))

	expected := []int{1, 10, 100}
	for i, val := range remaining {
		t.EqualExit(expected[i], val)
	}
}
