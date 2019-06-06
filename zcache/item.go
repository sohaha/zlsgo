/*
 * @Author: seekwe
 * @Date:   2019-05-24 19:16:09
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-25 13:53:54
 */

package zcache

import (
	"sync"
	"time"
)

// CacheItem CacheItem
type CacheItem struct {
	sync.RWMutex
	key            interface{}
	data           interface{}
	lifeSpan       time.Duration
	createdTime    time.Time
	accessedTime   time.Time
	accessCount    int64
	deleteCallback func(key interface{})
}

// newCacheItem newCacheItem
func newCacheItem(key interface{}, lifeSpan time.Duration, data interface{}) *CacheItem {
	t := time.Now()
	return &CacheItem{
		key:            key,
		lifeSpan:       lifeSpan,
		createdTime:    t,
		accessedTime:   t,
		accessCount:    0,
		deleteCallback: nil,
		data:           data,
	}
}

// KeepAlive KeepAlive
func (item *CacheItem) KeepAlive() {
	item.Lock()
	defer item.Unlock()
	item.accessedTime = time.Now()
	item.accessCount++
}

// LifeSpan LifeSpan
func (item *CacheItem) LifeSpan() time.Duration {
	return item.lifeSpan
}

// AccessedTime AccessedTime
func (item *CacheItem) AccessedTime() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.accessedTime
}

// CreatedTime CreatedTime
func (item *CacheItem) CreatedTime() time.Time {
	return item.createdTime
}

// AccessCount AccessCount
func (item *CacheItem) AccessCount() int64 {
	item.RLock()
	defer item.RUnlock()
	return item.accessCount
}

// Key Key
func (item *CacheItem) Key() interface{} {
	return item.key
}

// Data Data
func (item *CacheItem) Data() interface{} {
	return item.data
}

// SetDeleteCallback SetDeleteCallback
func (item *CacheItem) SetDeleteCallback(f func(interface{})) {
	item.Lock()
	defer item.Unlock()
	item.deleteCallback = f
}
