package fast_test

import (
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
