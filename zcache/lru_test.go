package zcache_test

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zcache"
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

	{
		time.Sleep(time.Millisecond * 3500)

		v, ok = l.Get("key1")
		tt.Equal("value1", v)
		tt.EqualTrue(ok)
		v, ok = l.Get("key3")
		tt.EqualTrue(!ok)
		tt.Equal(nil, v)
		v, ok = l.Get("key5")
		tt.EqualTrue(!ok)
		tt.Equal(nil, v)
	}

	{
		l.Delete("key1")

		v, ok = l.Get("key1")
		tt.EqualTrue(!ok)
		tt.Equal(nil, v)
		v, ok = l.Get("key3")
		tt.EqualTrue(!ok)
		tt.Equal(nil, v)
		v, ok = l.Get("key5")
		tt.EqualTrue(!ok)
		tt.Equal(nil, v)
	}

	{
		l.Set("key1", "new", time.Second/2)

		v, ok = l.Get("key1")
		tt.EqualTrue(ok)
		tt.Equal("new", v)
		v, ok = l.Get("key3")
		tt.EqualTrue(!ok)
		tt.Equal(nil, v)
		v, ok = l.Get("key5")
		tt.EqualTrue(!ok)
		tt.Equal(nil, v)
	}

	{
		time.Sleep(time.Millisecond * 1500)
		v, ok = l.Get("key1")
		tt.EqualTrue(!ok)
		tt.Equal(nil, v)
	}
}

func TestLRUCache(t *testing.T) {
	tt := zlsgo.NewTest(t)

	l := zcache.NewFast(func(o *zcache.Options) {
		o.Expiration = time.Second / 2
		o.Bucket = 4
		o.Cap = 10
		o.Callback = func(kind zcache.ActionKind, key string, ptr uintptr) {
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
		tt.EqualTrue(!ok)
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
