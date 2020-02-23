package zcache_test

import (
	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/zlog"
	"testing"
	"time"
)

func TestCache(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	// 初始一个名为 demo 的缓存对象
	c := zcache.New("demo")

	data := 666
	key := "name"

	// 设置缓存key为name,值为666,过期时间为10秒
	// 等同 c.SetRaw(key, data, 10*time.Second)
	c.Set(key, data, 10)

	t.EqualExit(1, c.Count())

	// 或取缓存数据
	name, err := c.Get(key)
	if err != nil {
		t.Fatal("cache name err: ", err)
	}

	// 判断缓存 key 是否存在
	t.EqualExit(true, c.Exists(key))

	// 如果缓存不存在则添加反之不生效
	t.EqualExit(false, c.Add("name", 999, 5*time.Second))

	t.EqualExit(data+10, name.(int)+10)

	// 删除缓存
	_, _ = c.Delete(key)
	t.EqualExit(false, c.Exists(key))

	_, err = c.Delete(key)
	t.Equal(zcache.ErrKeyNotFound, err)

	t.EqualExit(true, c.Add("name", 999, 5*time.Second))

	c.SetLoadNotCallback(func(key interface{}, args ...interface{}) *zcache.Item {
		return c.Set(key, "88", 10)
	})

	hoho, _ := c.Get("key2")
	t.EqualExit("88", hoho.(string))
	tt.Log(c.MostAccessed(1))
}

func TestDefCache(tt *testing.T) {
	t := zlsgo.NewTest(tt)

	key := "test_cache_key"
	key2 := "test_cache_key_2"
	key3 := "test_cache_key_3"

	zcache.SetLogger(zlog.Log)
	zcache.SetDeleteCallback(func(item *zcache.Item) bool {
		tt.Log("删除", item.Key())
		return true
	})

	data := "cache_data"
	zcache.Set(key, data, 1)
	zcache.Set(key2, data, 1)

	a, e := zcache.Get(key)
	tt.Log(a, e)

	ar, err := zcache.GetRaw(key)
	t.EqualExit(nil, err)
	tt.Log(ar.AccessCount())
	tt.Log(ar.RemainingLife())
	ar.AccessedTime()
	ar.LifeSpanUint()
	ar.LifeSpan()
	ar.RemainingLife()
	ar.Data()
	ar.Key()
	ar.CreatedTime()

	ar.SetDeleteCallback(func(item *zcache.Item) bool {
		tt.Log("拦截不删除", item.Key(), item.AccessCount())
		return false
	})

	a, e = zcache.Get(key)
	tt.Log(key, a, e)

	time.Sleep(time.Millisecond * 1100)

	a, e = zcache.Get(key)
	tt.Log(key, a, e)

	a, e = zcache.Get(key2)
	tt.Log(key2, a, e)

	a, e = zcache.Get(key3)
	tt.Log(key3, a, e)

	time.Sleep(time.Millisecond * 1100)

	a, e = zcache.Get(key)
	tt.Log(key, a, e)

	zcache.Set(key2, data, 1)
	a, e = zcache.Get(key2)
	tt.Log(key2, a, e)

	zcache.Set(key3, data, 1, true)
	a, e = zcache.Get(key3)
	tt.Log(key3, a, e)

	time.Sleep(time.Millisecond * 900)
	a, e = zcache.Get(key)
	tt.Log(key, a, e)

	a, e = zcache.Get(key2)
	tt.Log(key2, a, e)

	a, e = zcache.Get(key3)
	tt.Log(key3, a, e)

	time.Sleep(time.Millisecond * 900)
	a, e = zcache.Get(key)
	tt.Log(key, a, e)

	a, e = zcache.Get(key2)
	t.EqualExit(true, e != nil)

	a, e = zcache.Get(key3)
	t.EqualExit(data, a)
	t.EqualExit(nil, e)

	time.Sleep(time.Millisecond * 900)

	a, e = zcache.Get(key)
	tt.Log(key, a, e)

	_, e = zcache.Get(key2)
	t.EqualExit(true, e != nil)

	_, e = zcache.Get(key3)
	t.EqualExit(nil, e)
	_, _ = zcache.Delete(key3)
	a, e = zcache.Get(key3)
	tt.Log(key3, a, e)
	t.EqualExit(nil, a)
	t.EqualExit(true, e != nil)

	a, e = zcache.Get(key)
	tt.Log(key, a, e)
	t.EqualExit(data, a)
	t.EqualExit(nil, e)

	zcache.Clear()

	a, e = zcache.Get(key)
	tt.Log(key, a, e)
	t.EqualExit(nil, a)
	t.EqualExit(true, e != nil)
}

func TestCacheForEach(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	c := zcache.New("demo2")
	c.SetLogger(zlog.Log)

	data := 666

	c.Set("name1", data, 1)
	c.Set("name2", "name--2", 1)

	c.ForEach(func(key, value interface{}) {
		data, _ := c.GetRaw(key)
		tt.Log(data.Key(), data.Data(), data.LifeSpan())
	})

	time.Sleep(time.Millisecond * 1100)
	t.EqualExit(0, c.Count())
}
