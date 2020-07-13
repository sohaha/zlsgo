package zcache_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zstring"
	//"github.com/sohaha/zlsgo/zstring"
)

func TestCache(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	// 初始一个名为 demo 的缓存对象
	c := zcache.New(zstring.Rand(7))

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
	c.SetAddCallback(func(item *zcache.Item) {
		t.Log("SetAddCallback", item.Data())
	})
	c.Add("name", 999, 5*time.Second)
	// 删除缓存
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

func TestDefCache(tt *testing.T) {
	t := zlsgo.NewTest(tt)

	key := "test_cache_def_key"
	key2 := "test_cache_def_key_2"
	key3 := "test_cache_def_key_3"

	tt.Log("TestDefCache")
	zcache.SetDeleteCallback(func(key string) bool {
		fmt.Println("删除", key)
		return true
	})

	data := "cache_def_data"
	tt.Log(data)
	zcache.Set(key, data, 1)
	zcache.Set(key2, data, 1)

	a, e := zcache.Get(key)
	tt.Log(a, e)

	ar, err := zcache.GetT(key)
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

	ar.SetDeleteCallback(func(key string) bool {
		tt.Log("拦截不删除", key)
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

	_, e = zcache.Get(key2)
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
	c := zcache.New("CacheForEach")
	c.SetLogger(zlog.Log)

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

	zcache.SetLogger(zlog.Log)
	zcache.Set("TestOther", "123", 1)
	s, err := zcache.GetString("TestOther")
	tt.EqualNil(err)
	tt.Equal("123", s)

	zcache.Set("TestOther", 123, 1)
	i, err := zcache.GetInt("TestOther")
	tt.EqualNil(err)
	tt.Equal(123, i)
}

type testSt struct {
	Name string
	Key  int
}

func TestAccessCount(t *testing.T) {
	tt := zlsgo.NewTest(t)
	cache := zcache.New("AccessCount", true)
	cache.SetLogger(zlog.Log)

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
	time.Sleep(110 * 4 * time.Millisecond)
	i, err = cache.GetInt("TestOther")
	tt.Equal(true, err != nil)
	tt.Equal(0, i)
}

// func TestExportJSON(t *testing.T) {
// 	cache := zcache.New("ExportJSON")
// 	cache.Set("tmp1", &testSt{Name: "isName", Key: 100}, 1, true)
// 	cache.Set("tmp2", 666, 2)
// 	cache.Set("tmp3", "is string", 2)
// 	jsonData := cache.ExportJSON()
// 	t.Log(jsonData)
// }

func TestToFile(t *testing.T) {
	tt := zlsgo.NewTest(t)
	cache := zcache.New("file")
	zfile.WriteFile("tmp.json", []byte(`{"tmp2":"XhAAFSp6Y2FjaGUucGVyc2lzdGVuY2VTdP+BAwEBDXBlcnNpc3RlbmNlU3QB/4IAAQMBBERhdGEBEAABCExpZmVTcGFuAQQAARBJbnRlcnZhbExpZmVTcGFuAQIAAAAV/4ISAQNpbnQEBAD+BTQB/O5rKAAA","tmp3":"XhAAFSp6Y2FjaGUucGVyc2lzdGVuY2VTdP+BAwEBDXBlcnNpc3RlbmNlU3QB/4IAAQMBBERhdGEBEAABCExpZmVTcGFuAQQAARBJbnRlcnZhbExpZmVTcGFuAQIAAAAf/4IcAQZzdHJpbmcMCwAJaXMgc3RyaW5nAfzuaygAAA==","tmp1":"XhAAFSp6Y2FjaGUucGVyc2lzdGVuY2VTdP+BAwEBDXBlcnNpc3RlbmNlU3QB/4IAAQMBBERhdGEBEAABCExpZmVTcGFuAQQAARBJbnRlcnZhbExpZmVTcGFuAQIAAABW/4I6ARMqemNhY2hlX3Rlc3QudGVzdFN0/4MDAQEGdGVzdFN0Af+EAAECAQROYW1lAQwAAQNLZXkBBAAAABj/hAwBBmlzTmFtZQH/yAAB/Hc1lAABAQA="}`))
	save := cache.PersistenceToFile("tmp.json", false, &testSt{})

	tmp1, err := cache.Get("tmp1")
	tt.EqualExit(nil, err)
	tt.Equal("isName", tmp1.(*testSt).Name)
	tmp2, err := cache.GetInt("tmp2")
	tt.EqualNil(err)
	tt.Equal(666, tmp2)
	tmp3, err := cache.GetString("tmp3")
	tt.EqualNil(err)
	tt.Equal("is string", tmp3)
	go func() {
		time.Sleep(500 * time.Millisecond)
		cache.Get("tmp1")
	}()
	time.Sleep(3 * time.Second)
	tmp1, err = cache.Get("tmp1")
	tt.EqualNil(err)
	t.Log(tmp1)
	tmp2, err = cache.GetInt("tmp2")
	tt.EqualTrue(err != nil)
	tmp3, err = cache.GetString("tmp3")
	tt.EqualTrue(err != nil)

	cache.Set("tmp0", 1, 2)
	tt.Equal(2, cache.Count())
	zfile.Rmdir("tmp.json")

	err = save()
	tt.EqualNil(err)

	zfile.Rmdir("tmp.json")
}

func TestDo(t *testing.T) {
	var g sync.WaitGroup
	for i := 1; i <= 10; i++ {
		g.Add(1)
		go func(ii int) {
			if ii > 8 {
				time.Sleep(time.Duration(210*(ii-8)) * time.Millisecond)
			}
			v, o := zcache.MustGet("do", func(set func(data interface{},
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
