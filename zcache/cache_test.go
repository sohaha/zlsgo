package zcache

import (
	"github.com/sohaha/zlsgo"
	"testing"
	"time"
)

func TestCache(T *testing.T) {
	t := zlsgo.NewTest(T)
	// 初始一个名为 demo 的缓存对象
	c := New("demo")

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
	t.Equal(ErrKeyNotFound, err)

	t.EqualExit(true, c.Add("name", 999, 5*time.Second))

	c.SetLoadNotCallback(func(key interface{}, args ...interface{}) *CacheItem {
		return c.Set(key, "88", 10)
	})

	hoho, _ := c.Get("key2")
	t.EqualExit("88", hoho.(string))
}

func TestCacheForEach(T *testing.T) {
	t := zlsgo.NewTest(T)
	c := New("demo2")

	data := 666

	c.Set("name1", data, 1)
	c.Set("name2", data, 1)

	c.ForEach(func(key interface{}, item *CacheItem) {
		t.Log(key)
	})

	time.Sleep(2 * time.Second)

	t.EqualExit(0, c.Count())
}
