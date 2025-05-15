// Package zcache provides caching functionality with support for expiration, callbacks,
// and various cache management strategies. It offers both in-memory and persistent
// caching solutions for Go applications.
package zcache

import (
	"errors"
	"sync"

	"github.com/sohaha/zlsgo/zutil"
)

var (
	// ErrKeyNotFound is returned when a requested key does not exist in the cache
	ErrKeyNotFound = errors.New("key is not in cache")
	// ErrKeyNotFoundAndNotCallback is returned when a key is not found and no callback function is provided
	ErrKeyNotFoundAndNotCallback = errors.New("key is not in cache and no callback is set")
	// Internal cache registry for named cache tables
	cache                        = make(map[string]*Table)
	mutex                        sync.RWMutex
)

// New creates or retrieves a named cache table with optional access counting.
// If a table with the specified name already exists, it is returned.
// If accessCount is true, the cache will track the number of times each item is accessed.
// 
// Deprecated: please use zcache.NewFast instead
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
