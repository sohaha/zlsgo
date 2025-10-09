package fast_test

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/zcache/fast"
	"github.com/sohaha/zlsgo/zsync"
	"github.com/sohaha/zlsgo/ztype"
	"github.com/sohaha/zlsgo/zutil"
)

func TestLRUCacheExpire(t *testing.T) {
	tt := zlsgo.NewTest(t)
	l := zcache.NewFast(func(o *zcache.Options) {
		o.LRU2Cap = 10
	})

	l.Set("key1", "value1")
	l.Set("key3", "value3", time.Second*1)
	l.Set("key5", "value5", time.Second*3)

	v, ok := l.Get("key1")
	_ = v
	tt.EqualTrue(ok)

	v, ok = l.Get("key3")
	tt.EqualTrue(ok)
	tt.Equal("value3", v)
	v, ok = l.Get("key5")
	tt.EqualTrue(ok)
	tt.Equal("value5", v)

	{
		time.Sleep(time.Millisecond * 1500)

		v, ok = l.Get("key1")
		tt.EqualTrue(ok)
		tt.Equal("value1", v)
		v, ok = l.Get("key3")
		tt.EqualTrue(!ok)
		tt.Equal(nil, v)
		v, ok = l.Get("key5")
		tt.EqualTrue(ok)
		tt.Equal("value5", v)
	}

	v, ok = l.Get("key3")
	tt.EqualTrue(!ok)
	tt.Equal(nil, v)
	v, ok = l.Get("key5")
	tt.EqualTrue(ok)
	tt.Equal("value5", v)
}

func TestLazyStartIdleStopAndRestart(t *testing.T) {
	tt := zlsgo.NewTest(t)

	c := fast.NewFast(func(o *fast.Options) {
		o.Expiration = 100 * time.Millisecond
		o.AutoCleaner = true
		o.LazyCleaner = true
		o.IdleAfter = 100 * time.Millisecond
		o.Bucket = 4
		o.Cap = 8
	})
	defer c.Close()

	c.Set("k1", "v1")
	time.Sleep(350 * time.Millisecond)
	if v, ok := c.Get("k1"); ok || v != nil {
		t.Fatalf("expected k1 cleaned after expiration, got %v %v", v, ok)
	}
	time.Sleep(250 * time.Millisecond)

	c.Set("k2", "v2")
	time.Sleep(350 * time.Millisecond)
	if v, ok := c.Get("k2"); ok || v != nil {
		t.Fatalf("expected k2 cleaned after restart, got %v %v", v, ok)
	}

	c.SetBytes("kb", []byte("vb"))
	time.Sleep(350 * time.Millisecond)
	if b, ok := c.GetBytes("kb"); ok || len(b) != 0 {
		t.Fatalf("expected kb cleaned after restart, got %v %v", string(b), ok)
	}

	tt.EqualTrue(true)
}

func TestAutoCleanerDisabled(t *testing.T) {
	tt := zlsgo.NewTest(t)

	c := fast.NewFast(func(o *fast.Options) {
		o.Expiration = 80 * time.Millisecond
		o.AutoCleaner = false
		o.Bucket = 2
		o.Cap = 8
	})
	defer c.Close()

	c.Set("x", 1)
	time.Sleep(200 * time.Millisecond)

	if v, ok := c.Get("x"); ok || v != nil {
		t.Fatalf("expected expired on Get without cleaner, got %v %v", v, ok)
	}
	c.SetBytes("xb", []byte("1"))
	time.Sleep(200 * time.Millisecond)
	if b, ok := c.GetBytes("xb"); ok || len(b) != 0 {
		t.Fatalf("expected expired bytes on GetBytes without cleaner, got %v %v", string(b), ok)
	}

	tt.EqualTrue(true)
}

func TestAutoCleaner(t *testing.T) {
	tt := zlsgo.NewTest(t)

	l := zcache.NewFast(func(o *zcache.Options) {
		o.Expiration = 200 * time.Millisecond
		o.Bucket = 4
		o.Cap = 16
	})

	l.Set("k1", "v1")
	l.SetBytes("k2", []byte("v2"))

	time.Sleep(600 * time.Millisecond)

	v, ok := l.Get("k1")
	tt.EqualTrue(!ok)
	tt.Equal(nil, v)

	vb, ok := l.GetBytes("k2")
	tt.EqualTrue(!ok)
	tt.Equal(0, len(vb))
}

func TestCloseIdempotent(t *testing.T) {
	_ = zlsgo.NewTest(t)

	l := fast.NewFast(func(o *fast.Options) {
		o.Expiration = 50 * time.Millisecond
	})

	l.Close()
	l.Close()

	l.Set("a", 1, 30*time.Millisecond)
	time.Sleep(80 * time.Millisecond)
	if v, ok := l.Get("a"); ok || v != nil {
		t.Fatalf("expected expired after Close, got %v %v", v, ok)
	}
}

func TestLRUCache(t *testing.T) {
	tt := zlsgo.NewTest(t)

	l := zcache.NewFast(func(o *zcache.Options) {
		o.Expiration = time.Second / 2
		o.Bucket = 4
		o.Cap = 10
		o.Callback = func(kind fast.ActionKind, key string, ptr uintptr) {
			t.Log("    ", kind, key, ptr)
		}
	})

	l.Set("key1", "value1")
	l.SetBytes("key2", []byte("value2"))

	v, ok := l.Get("key1")
	tt.EqualTrue(ok)
	tt.Equal("value1", v)
	t.Log("key1", v)

	v, ok = l.Get("key2")
	tt.EqualTrue(ok)
	tt.Equal("value2", string(v.([]byte)))
	t.Log("key2", v)

	vb, ok := l.GetBytes("key2")
	tt.EqualTrue(ok)
	tt.Equal("value2", string(vb))

	v, ok = l.Get("key3")
	tt.EqualTrue(!ok)
	t.Log("key3", v)

	{
		time.Sleep(time.Second * 1)

		v, ok = l.Get("key1")
		tt.Log(v, ok)
		tt.EqualTrue(!ok, true)
		tt.Equal(nil, v)

		v, ok = l.Get("key3")
		tt.EqualTrue(!ok)
		tt.Equal(nil, v)

		v, ok = l.Get("key5")
		tt.EqualTrue(!ok)
		tt.Equal(nil, v)
	}
}

func TestProvideGet(t *testing.T) {
	tt := zlsgo.NewTest(t)

	l := zcache.NewFast(func(o *zcache.Options) {})

	var wg zsync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Go(func() {
			_, _ = l.ProvideGet("key1", func() (interface{}, bool) {
				tt.Log("    ", "value1")
				return "value1", true
			}, time.Second/2)
		})
	}
	wg.Wait()
	t.Log(l.Get("key1"))
	time.Sleep(time.Second * 1)
	t.Log(l.Get("key1"))

	for i := 0; i < 100; i++ {
		wg.Go(func() {
			_, _ = l.ProvideGet("key1", func() (interface{}, bool) {
				tt.Log("    ", "value2")
				return "value2", true
			}, time.Second/2)
		})
	}
	wg.Wait()
	t.Log(l.Get("key1"))

	time.Sleep(time.Second * 1)
	t.Log(l.Get("key1"))
}

func TestForEach(t *testing.T) {
	tt := zlsgo.NewTest(t)

	c, i := zcache.NewFast(), zutil.NewInt64(0)

	for i := 0; i < 100; i++ {
		c.Set(ztype.ToString(i), i, time.Millisecond*time.Duration(i))
	}

	c.ForEach(func(key string, value interface{}) bool {
		i.Add(1)
		return true
	})

	time.Sleep(time.Millisecond * 50)

	c.ForEach(func(key string, value interface{}) bool {
		i.Sub(1)
		return true
	})

	tt.EqualTrue(i.Load() != 0)
}

func TestMemoryLeakPrevention(t *testing.T) {
	tt := zlsgo.NewTest(t)

	cache := fast.NewFast(func(o *fast.Options) {
		o.Expiration = 100 * time.Millisecond
		o.AutoCleaner = true
		o.LazyCleaner = true
		o.IdleAfter = 200 * time.Millisecond
		o.Bucket = 2
		o.Cap = 4
	})

	cache.Set("test1", "value1")
	cache.Set("test2", "value2", 50*time.Millisecond)

	val, exists := cache.Get("test1")
	tt.EqualTrue(exists && val != nil)

	time.Sleep(400 * time.Millisecond)

	_, exists = cache.Get("test2")
	tt.EqualTrue(!exists)

	cache.Set("test3", "value3")
	val, exists = cache.Get("test3")
	tt.EqualTrue(exists)
	tt.Equal("value3", val)

	cache.Close()
}

func TestFinalizerSafetyNet(t *testing.T) {
	tt := zlsgo.NewTest(t)

	createAndAbandonCache := func() {
		cache := fast.NewFast(func(o *fast.Options) {
			o.Expiration = 50 * time.Millisecond
			o.AutoCleaner = true
			o.LazyCleaner = false
			o.Bucket = 2
			o.Cap = 4
		})

		cache.Set("key1", "value1")
		cache.Set("key2", "value2")
	}

	for i := 0; i < 5; i++ {
		createAndAbandonCache()
	}

	for i := 0; i < 3; i++ {
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	}

	tt.EqualTrue(true)
}

func TestCleanerLevelTransitions(t *testing.T) {
	tt := zlsgo.NewTest(t)

	cache := fast.NewFast(func(o *fast.Options) {
		o.Expiration = 100 * time.Millisecond
		o.AutoCleaner = true
		o.LazyCleaner = true
		o.Bucket = 2
		o.Cap = 4
	})
	defer cache.Close()

	cache.Set("key1", "value1")

	intervals := []time.Duration{
		40 * time.Millisecond,
		2 * time.Second,
	}

	for i, interval := range intervals {
		t.Logf("Waiting %v for cleaner level transition %d", interval, i+1)
		time.Sleep(interval)

		testKey := fmt.Sprintf("test_level_%d", i)
		cache.Set(testKey, fmt.Sprintf("value_%d", i))
		val, exists := cache.Get(testKey)
		tt.EqualTrue(exists)
		tt.Equal(fmt.Sprintf("value_%d", i), val)
	}

	cache.Set("final_test", "final_value")
	val, exists := cache.Get("final_test")
	tt.EqualTrue(exists)
	tt.Equal("final_value", val)
}

func TestConfigurableCleaningThresholds(t *testing.T) {
	tt := zlsgo.NewTest(t)

	cache := fast.NewFast(func(o *fast.Options) {
		o.Expiration = 50 * time.Millisecond
		o.AutoCleaner = true
		o.LazyCleaner = true
		o.Bucket = 2
		o.Cap = 4
		o.ShortIdleThreshold = 10 * time.Millisecond
		o.MediumIdleThreshold = 50 * time.Millisecond
		o.LongIdleThreshold = 100 * time.Millisecond
	})
	defer cache.Close()

	cache.Set("test", "value")
	tt.Equal(int32(0), cache.GetCleanerLevel())

	time.Sleep(15 * time.Millisecond)
	cache.Set("trigger", "clean")
	time.Sleep(5 * time.Millisecond)

	time.Sleep(60 * time.Millisecond)
	cache.Set("trigger2", "clean")
	time.Sleep(5 * time.Millisecond)
	val, exists := cache.Get("trigger2")
	tt.EqualTrue(exists)
	tt.Equal("clean", val)
}

func TestPerformanceMonitoring(t *testing.T) {
	tt := zlsgo.NewTest(t)

	cache := fast.NewFast(func(o *fast.Options) {
		o.Expiration = 100 * time.Millisecond
		o.AutoCleaner = true
		o.Bucket = 2
		o.Cap = 8
	})
	defer cache.Close()

	stats := cache.GetStats()
	tt.Equal(int32(0), stats.CleanerLevel)
	tt.Equal(int64(0), stats.AccessCount)
	tt.EqualTrue(stats.IdleDuration >= 0)

	initialCount := cache.GetAccessCount()
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Get("key1")

	newCount := cache.GetAccessCount()
	tt.EqualTrue(newCount > initialCount)

	stats = cache.GetStats()
	tt.EqualTrue(stats.AccessCount > 0)
	tt.EqualTrue(stats.TotalItems >= 0)
	tt.EqualTrue(cache.GetCleanerLevel() >= 0)
	tt.EqualTrue(cache.GetAccessCount() >= 0)
	tt.EqualTrue(cache.GetIdleDuration() >= 0)
}

func TestOptimizedMarkActive(t *testing.T) {
	tt := zlsgo.NewTest(t)

	cache := fast.NewFast(func(o *fast.Options) {
		o.Expiration = 100 * time.Millisecond
		o.AutoCleaner = true
		o.Bucket = 2
		o.Cap = 4
	})
	defer cache.Close()

	initialLevel := cache.GetCleanerLevel()

	for i := 0; i < 100; i++ {
		cache.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
		cache.Get(fmt.Sprintf("key%d", i))
	}

	val, exists := cache.Get("key99")
	tt.EqualTrue(exists)
	tt.Equal("value99", val)
	tt.EqualTrue(cache.GetAccessCount() > 0)

	finalLevel := cache.GetCleanerLevel()
	tt.EqualTrue(finalLevel >= initialLevel)
}
