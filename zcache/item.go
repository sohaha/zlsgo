package zcache

import (
	"sync"
	"time"
)

// Item Item
type Item struct {
	sync.RWMutex
	key              interface{}
	data             interface{}
	lifeSpan         time.Duration
	createdTime      time.Time
	accessedTime     time.Time
	accessCount      int64
	intervalLifeSpan bool
	deleteCallback   func(item *Item) bool
}

// newCacheItem newCacheItem
func newCacheItem(key interface{}, data interface{}, lifeSpan time.Duration) *Item {
	t := time.Now()
	return &Item{
		key:              key,
		lifeSpan:         lifeSpan,
		createdTime:      t,
		accessedTime:     t,
		accessCount:      0,
		intervalLifeSpan: false,
		deleteCallback:   nil,
		data:             data,
	}
}

func (item *Item) keepAlive() {
	item.Lock()
	item.accessedTime = time.Now()
	item.accessCount++
	item.Unlock()
}

// LifeSpan LifeSpan
func (item *Item) LifeSpan() time.Duration {
	return item.lifeSpan
}

// LifeSpanUint LifeSpanUint
func (item *Item) LifeSpanUint() uint {
	return uint(item.lifeSpan / time.Second)
}

// AccessedTime AccessedTime
func (item *Item) AccessedTime() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.accessedTime
}

// CreatedTime CreatedTime
func (item *Item) CreatedTime() time.Time {
	return item.createdTime
}

// RemainingLife RemainingLife
func (item *Item) RemainingLife() time.Duration {
	return item.createdTime.Add(item.lifeSpan).Sub(time.Now())
}

// AccessCount AccessCount
func (item *Item) AccessCount() int64 {
	item.RLock()
	defer item.RUnlock()
	return item.accessCount
}

// Key item key
func (item *Item) Key() interface{} {
	return item.key
}

// Data data
func (item *Item) Data() interface{} {
	return item.data
}

// SetDeleteCallback SetDeleteCallback
func (item *Item) SetDeleteCallback(fn func(item *Item) bool) {
	item.Lock()
	item.deleteCallback = fn
	item.Unlock()
}
