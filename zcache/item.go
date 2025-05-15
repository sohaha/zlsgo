package zcache

import (
	"sync"
	"time"

	"github.com/sohaha/zlsgo/ztime"
)

// Item represents a single cache entry with associated metadata such as
// expiration time, access statistics, and lifecycle callbacks
type Item struct {
	createdTime    time.Time
	accessedTime   time.Time
	data           interface{}
	deleteCallback func(key string) bool
	key            string
	lifeSpan       time.Duration
	accessCount    int64
	sync.RWMutex
	intervalLifeSpan bool
}

// NewCacheItem creates a new cache item with the specified key, data, and lifespan.
// The item's creation and access times are initialized to the current time.
func NewCacheItem(key string, data interface{}, lifeSpan time.Duration) *Item {
	t := ztime.UnixMicro(ztime.Clock())
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

// keepAlive updates the item's access time to the current time and
// increments its access counter. This is used for tracking usage and
// implementing interval-based expiration policies.
func (item *Item) keepAlive() {
	item.Lock()
	item.accessedTime = ztime.UnixMicro(ztime.Clock())
	item.accessCount++
	item.Unlock()
}

// LifeSpan returns the duration for which this item will remain in the cache
// before expiring. A zero duration indicates that the item never expires.
func (item *Item) LifeSpan() time.Duration {
	return item.lifeSpan
}

// IntervalLifeSpan returns whether this item uses interval-based expiration,
// which resets the expiration timer each time the item is accessed.
func (item *Item) IntervalLifeSpan() bool {
	return item.intervalLifeSpan
}

// LifeSpanUint returns the item's lifespan converted to unsigned integer seconds.
// This is useful for interfaces that require the lifespan in seconds.
func (item *Item) LifeSpanUint() uint {
	return uint(item.lifeSpan / time.Second)
}

// AccessedTime returns the time when this item was last accessed.
// This is used for implementing expiration policies and usage analytics.
func (item *Item) AccessedTime() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.accessedTime
}

// CreatedTime returns the time when this item was created.
// This is used for calculating absolute expiration times.
func (item *Item) CreatedTime() time.Time {
	return item.createdTime
}

// RemainingLife calculates and returns the remaining time until this item expires.
// Returns 0 if the item has no expiration (lifeSpan == 0) or if it has already expired.
func (item *Item) RemainingLife() time.Duration {
	if item.lifeSpan == 0 {
		return 0
	}
	return time.Until(item.createdTime.Add(item.lifeSpan))
}

// AccessCount returns the number of times this item has been accessed.
// This is useful for implementing cache eviction policies based on usage patterns.
func (item *Item) AccessCount() int64 {
	item.RLock()
	defer item.RUnlock()
	return item.accessCount
}

// Key returns the unique identifier for this cache item.
func (item *Item) Key() interface{} {
	return item.key
}

// Data returns the value stored in this cache item.
func (item *Item) Data() interface{} {
	item.RLock()
	defer item.RUnlock()
	return item.data
}

// SetDeleteCallback sets a function to be called before this item is deleted from the cache.
// If the callback returns false, the deletion is aborted.
func (item *Item) SetDeleteCallback(fn func(key string) bool) {
	item.Lock()
	item.deleteCallback = fn
	item.Unlock()
}
