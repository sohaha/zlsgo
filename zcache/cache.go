package zcache

import (
	"errors"
	"sync"

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

func Set(key, data interface{}, lifeSpan uint, interval ...bool) {
	ite.Set(key, data, lifeSpan, interval...)
}

func Delete(key interface{}) (*Item, error) {
	return ite.Delete(key)
}

func Clear() {
	ite.Clear()
}

func Get(key interface{}) (value interface{}, err error) {
	return ite.Get(key)
}

func GetRaw(key interface{}) (*Item, error) {
	return ite.GetRaw(key)
}

func GetLocked(key interface{}) (interface{}, func(data interface{}, lifeSpan uint, interval ...bool)) {
	return ite.GetLocked(key)
}

func SetDeleteCallback(f func(*Item) bool) {
	ite.SetDeleteCallback(f)
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
				items: make(map[interface{}]*Item),
			}
			cache[table] = t
		}
		mutex.Unlock()
	}

	return t
}
