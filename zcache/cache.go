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
