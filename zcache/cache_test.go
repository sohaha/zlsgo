package zcache_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/zstring"
)

func TestCache(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	// Initialize a cache object named demo
	c := zcache.New(zstring.Rand(7))

	data := 666
	key := "name"

	// Set cache key to name, value to 666, expiration time to 10 seconds
	// Equivalent to c.SetRaw(key, Raw, 10*time.Second)
	c.Set(key, data, 10)

	t.EqualExit(1, c.Count())

	// Get cache data
	name, err := c.Get(key)
	if err != nil {
		t.Fatal("cache name err: ", err)
	}

	// Check if cache key exists
	t.EqualExit(true, c.Exists(key))

	// Add if cache doesn't exist, otherwise no effect
	t.EqualExit(false, c.Add("name", 999, 5*time.Second))
	t.EqualExit(data+10, name.(int)+10)
	c.SetAddCallback(func(item *zcache.Item) {
		t.Log("SetAddCallback", item.Data())
	})
	c.Add("name", 999, 5*time.Second)
	// Delete cache
	_, _ = c.Delete(key)
	t.EqualExit(false, c.Exists(key))

	_, err = c.Delete(key)
	t.Equal(zcache.ErrKeyNotFound, err)

	t.EqualExit(true, c.Add("name", 999, 5*time.Second))

	c.SetLoadNotCallback(func(key string, args ...interface{}) *zcache.Item {
		return c.Set(key, "88", 10)
	})

	hoho, _ := c.Get("key2")
	t.EqualExit("88", hoho.(string))
	tt.Log(c.MostAccessed(1))
}

func TestCacheForEach(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	c := zcache.New("CacheForEach")

	data := 666

	c.Set("name1", data, 1)
	c.Set("name2", "name--2", 1)

	tt.Log("ForEach:")
	c.ForEach(func(key string, value interface{}) bool {
		data, _ := c.GetT(key)
		tt.Log("ForEach", key)
		tt.Log(data.Key(), data.Data(), data.LifeSpan())
		return true
	})

	i := 0
	c.ForEach(func(key string, value interface{}) bool {
		i++
		return false
	})

	t.Equal(1, i)
	time.Sleep(time.Millisecond * 1100)
	t.EqualExit(0, c.Count())
}

func TestOther(t *testing.T) {
	tt := zlsgo.NewTest(t)

	c := zcache.New("TestOther")
	c.Set("TestOther", "123", 1)
	s, err := c.GetString("TestOther")
	tt.EqualNil(err)
	tt.Equal("123", s)

	c.Set("TestOther", 123, 1)
	i, err := c.GetInt("TestOther")
	tt.EqualNil(err)
	tt.Equal(123, i)
}

func TestAccessCount(t *testing.T) {
	tt := zlsgo.NewTest(t)
	cache := zcache.New("AccessCount", true)

	cache.SetRaw("TestOther", 123, 100*time.Millisecond, true)
	i, err := cache.GetInt("TestOther")
	tt.EqualNil(err)
	tt.Equal(123, i)
	time.Sleep(90 * time.Millisecond)
	i, err = cache.GetInt("TestOther")
	tt.EqualNil(err)
	tt.Equal(123, i)
	time.Sleep(90 * time.Millisecond)
	i, err = cache.GetInt("TestOther")
	tt.EqualNil(err)
	tt.Equal(123, i)
	time.Sleep(time.Second * 1)
	i, err = cache.GetInt("TestOther")
	t.Log(i, err)
}

// func TestExportJSON(t *testing.T) {
// 	cache := zcache.New("ExportJSON")
// 	cache.Set("tmp1", &testSt{Name: "isName", Key: 100}, 1, true)
// 	cache.Set("tmp2", 666, 2)
// 	cache.Set("tmp3", "is string", 2)
// 	jsonData := cache.ExportJSON()
// 	t.Log(jsonData)
// }

func TestDo(t *testing.T) {
	var g sync.WaitGroup
	c := zcache.New("TestOther")
	for i := 1; i <= 10; i++ {
		g.Add(1)
		go func(ii int) {
			if ii > 8 {
				time.Sleep(time.Duration(210*(ii-8)) * time.Millisecond)
			}
			v, o := c.MustGet("do", func(set func(data interface{},
				lifeSpan time.Duration, interval ...bool)) (err error) {
				if ii < 9 {
					set(ii, 200*time.Millisecond)
					return nil
				} else if ii == 9 {
					set("ok", 200*time.Millisecond)
					return nil
				}
				return errors.New("不设置")
			})
			t.Log(ii, o, v)
			g.Done()
		}(i)
	}
	g.Wait()
}
