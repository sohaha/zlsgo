package zcache_test

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zcache"
)

func TestLRUCacheExpire(t *testing.T) {
	tt := zlsgo.NewTest(t)
	l := zcache.NewFast()

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

	l := zcache.NewFast(func(o *zcache.Option) {
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
