// Package zcache cache operation
package zcache

import (
	"errors"
	"sync"

	"github.com/sohaha/zlsgo/zutil"
)

var (
	// ErrKeyNotFound ErrKeyNotFound
	ErrKeyNotFound = errors.New("key is not in cache")
	// ErrKeyNotFoundAndNotCallback ErrKeyNotFoundAndNotCallback
	ErrKeyNotFoundAndNotCallback = errors.New("key is not in cache and no callback is set")
	cache                        = make(map[string]*Table)
	mutex                        sync.RWMutex
)

// Deprecated: please use zcache.NewFast
// New new cache
func New(table string, accessCount ...bool) *Table {
	mutex.Lock()
	t, ok := cache[table]

	if !ok {
		t, ok = cache[table]
		if !ok {
			t = &Table{
				name:  table,
				items: make(map[string]*Item),
			}
			t.accessCount = zutil.NewBool(len(accessCount) > 0 && accessCount[0])
			cache[table] = t
		}
	}

	mutex.Unlock()
	return t
}
