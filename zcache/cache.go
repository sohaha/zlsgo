/*
 * @Author: seekwe
 * @Date:   2019-05-24 19:15:39
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-25 15:29:33
 */

package zcache

import (
	"errors"
	"sync"

	"github.com/sohaha/zlsgo/znet"
)

var (
	// ErrKeyNotFound ErrKeyNotFound
	ErrKeyNotFound = errors.New("key is not in cache")
	// ErrKeyNotFoundAndNotCallback ErrKeyNotFoundAndNotCallback
	ErrKeyNotFoundAndNotCallback = errors.New("key is not in cache and no callback is set")
	cache                        = make(map[string]*Table)
	mutex                        sync.RWMutex
)

// New New
func New(table string) *Table {
	mutex.RLock()
	t, ok := cache[table]
	mutex.RUnlock()

	if !ok {
		mutex.Lock()
		t, ok = cache[table]
		if !ok {
			t = &Table{
				name:  table,
				items: make(map[interface{}]*CacheItem),
			}
			cache[table] = t
		}
		mutex.Unlock()
	}

	return t
}

func NewHTTP(h znet.HandlerFunc) znet.HandlerFunc {
	_ = New("znet_http")
	return func(c *znet.Context) {
		c.String(200, "ok我是测试的那")
		c.String(200, "ok我是测试的那")
		// var log bytes.Buffer
		// rsp := io.MultiWriter(c.Writer, &log)
		// url := c.Request.URL.String()
		// encodeString := base64.StdEncoding.EncodeToString([]byte(url))
		// _, err := ca.Get(encodeString)
		// // c.String(200, "2是我赛坑囧 框架2")
		// // c.Writer.Writer.Write([]byte("hji"))
		// rsp.Write([]byte("hji"))
		// if err != nil {

		// }
		// // c.Abort()
		c.Next()
		// fmt.Println("ca", ca)
		// // from this point on use rsp instead of w, ie
		// // err := json.NewDecoder(c.Request.Body).Decode(&requestData)
		// // if err != nil {
		// // 	// writeError(rsp, "JSON request is not in correct format")
		// // 	return
		// // }
		// // c.Writer = rsp
		// fmt.Println(rsp)
		// fmt.Println(log.String(), "--", rsp)
		// rsp = io.MultiWriter(c.Writer, &log)
		// fmt.Println(log.String(), "--", c.Writer.Output)
		// io.ReadWriter(c.Writer)
		// fmt.Println(item, c.Writer)
		// fmt.Println(item, c.Writer)
		// fmt.Printf("%+v\n%[1]T", c.Writer)

	}
}
