// Package zcache cache operation
package zcache

import (
	"errors"
	"sync"
	"time"

	"github.com/sohaha/zlsgo/zlog"
)

var (
	// ErrKeyNotFound ErrKeyNotFound
	ErrKeyNotFound = errors.New("key is not in cache")
	// ErrKeyNotFoundAndNotCallback ErrKeyNotFoundAndNotCallback
	ErrKeyNotFoundAndNotCallback = errors.New("key is not in cache and no callback is set")
	cache                        = make(map[string]*Table)
	ite                          = New("defaultCache")
	mutex                        sync.RWMutex
)

func SetLogger(logger *zlog.Logger) {
	ite.SetLogger(logger)
}

func Set(key string, data interface{}, lifeSpan uint, interval ...bool) {
	ite.Set(key, data, lifeSpan, interval...)
}

func Delete(key string) (*Item, error) {
	return ite.Delete(key)
}

func Clear() {
	ite.Clear()
}

func Get(key string) (value interface{}, err error) {
	return ite.Get(key)
}

func GetInt(key string) (value int, err error) {
	return ite.GetInt(key)
}

func GetString(key string) (value string, err error) {
	return ite.GetString(key)
}

func GetT(key string) (*Item, error) {
	return ite.GetT(key)
}

// MustGet get the data of the specified key, set if it does not exist
func MustGet(key string, do func(set func(data interface{},
	lifeSpan time.Duration, interval ...bool)) (
	err error)) (data interface{}, err error) {
	return ite.MustGet(key, do)
}

func SetDeleteCallback(fn func(key string) bool) {
	ite.SetDeleteCallback(fn)
}

// New new cache
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
				items: make(map[string]*Item),
			}
			cache[table] = t
		}
		mutex.Unlock()
	}

	return t
}
