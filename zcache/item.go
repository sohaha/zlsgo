package zcache

import (
	"sync"
	"time"
)

// Item Item
type Item struct {
	sync.RWMutex
	key              string
	data             interface{}
	lifeSpan         time.Duration
	createdTime      time.Time
	accessedTime     time.Time
	accessCount      int64
	intervalLifeSpan bool
	deleteCallback   func(key string) bool
}

// NewCacheItem NewCacheItem
func NewCacheItem(key string, data interface{}, lifeSpan time.Duration) *Item {
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
	if item.lifeSpan == 0 {
		return -1
	}
	return time.Until(item.createdTime.Add(item.lifeSpan))
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
	item.RLock()
	defer item.RUnlock()
	return item.data
}

// SetDeleteCallback SetDeleteCallback
func (item *Item) SetDeleteCallback(fn func(key string) bool) {
	item.Lock()
	item.deleteCallback = fn
	item.Unlock()
}
