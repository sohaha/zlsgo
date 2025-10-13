package zcache

import (
	"sort"
	"sync"
	"time"

	"github.com/sohaha/zlsgo/ztime"
	"github.com/sohaha/zlsgo/zutil"
	// "github.com/sohaha/zlsgo/zlog"
	// "sync/atomic"
)

type (
	// CacheItemPair maps a cache key to its access counter for tracking usage statistics
	CacheItemPair struct {
		Key         string
		AccessCount int64
	}
	// CacheItemPairList represents a collection of CacheItemPair objects
	// that can be sorted by access count for analytics and cache management
	CacheItemPairList []CacheItemPair
	// Table represents a named cache container that manages a collection of cached items
	// with support for expiration, callbacks, and access tracking
	Table struct {
		items           map[string]*Item
		cleanupTimer    *time.Timer
		loadNotCallback func(key string, args ...interface{}) *Item
		addCallback     func(item *Item)
		deleteCallback  func(key string) bool
		accessCount     *zutil.Bool
		name            string
		cleanupInterval time.Duration
		sync.RWMutex
	}
)

// Count returns the total number of items currently stored in the cache table
func (table *Table) Count() int {
	table.RLock()
	count := len(table.items)
	table.RUnlock()
	return count
}

// ForEach iterates through all cache items and applies the provided function to each key-value pair.
// The iteration continues as long as the function returns true, and stops when it returns false.
// This method provides access to the cached data values directly.
func (table *Table) ForEach(trans func(key string, value interface{}) bool) {
	table.ForEachRaw(func(k string, v *Item) bool {
		return trans(k, v.Data())
	})
}

// ForEachRaw iterates through all cache items and applies the provided function to each key-item pair.
// The iteration continues as long as the function returns true, and stops when it returns false.
// This method provides access to the raw Item objects, including metadata.
func (table *Table) ForEachRaw(trans func(key string, value *Item) bool) {
	count := table.Count()
	table.RLock()
	items := make(map[string]*Item, count)
	for k, v := range table.items {
		items[k] = v
	}
	table.RUnlock()

	for k, v := range items {
		if !trans(k, v) {
			break
		}
	}
}

// SetLoadNotCallback sets a function to be called when a requested item is not found in the cache.
// The callback function can generate a new cache item based on the key and additional arguments.
func (table *Table) SetLoadNotCallback(f func(key string, args ...interface{}) *Item) {
	table.Lock()
	table.loadNotCallback = f
	table.Unlock()
}

// SetAddCallback sets a function to be called whenever a new item is added to the cache.
// This can be used for logging, synchronization with external storage, or other side effects.
func (table *Table) SetAddCallback(f func(*Item)) {
	table.Lock()
	defer table.Unlock()
	table.addCallback = f
}

// SetDeleteCallback sets a function to be called before an item is deleted from the cache.
// If the callback returns false, the deletion is aborted.
func (table *Table) SetDeleteCallback(f func(key string) bool) {
	table.Lock()
	defer table.Unlock()
	table.deleteCallback = f
}

// expirationCheck scans all items in the cache and removes expired ones.
// It also schedules the next cleanup based on the item with the earliest expiration time.
func (table *Table) expirationCheck() {
	now := ztime.UnixMicro(ztime.Clock())
	smallestDuration := time.Duration(0)
	table.Lock()
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}

	for key, item := range table.items {
		item.RLock()
		lifeSpan := item.lifeSpan
		accessedOn := item.accessedTime
		intervalLifeSpan := item.intervalLifeSpan
		item.RUnlock()
		if lifeSpan == 0 {
			continue
		}
		remainingLift := item.RemainingLife()
		if table.accessCount.Load() && intervalLifeSpan {
			lastTime := now.Sub(accessedOn)
			if lastTime >= lifeSpan {
				_, _ = table.deleteInternal(key)
			} else {
				lifeSpan = lifeSpan * 2
				item.Lock()
				item.lifeSpan = lifeSpan
				item.Unlock()
				nextDuration := lifeSpan - lastTime
				if smallestDuration == 0 || nextDuration < smallestDuration {
					smallestDuration = nextDuration
				}
			}
		} else if remainingLift <= 0 {
			_, _ = table.deleteInternal(key)
		} else {
			if smallestDuration == 0 || smallestDuration > remainingLift {
				smallestDuration = remainingLift
			}
		}
	}
	table.cleanupInterval = smallestDuration

	if smallestDuration > 0 {
		table.cleanupTimer = time.AfterFunc(smallestDuration, func() {
			go table.expirationCheck()
		})
	}
	table.Unlock()
}

// addInternal adds an item to the cache and handles related operations such as
// invoking callbacks and scheduling expiration checks if needed.
func (table *Table) addInternal(item *Item) {
	table.Lock()
	table.items[item.key] = item

	expDur := table.cleanupInterval
	addedItem := table.addCallback
	table.Unlock()

	if addedItem != nil {
		addedItem(item)
	}
	item.RLock()
	lifeSpan := item.lifeSpan
	item.RUnlock()
	if lifeSpan > 0 && (expDur == 0 || lifeSpan < expDur) {
		go table.expirationCheck()
	}
}

// SetRaw adds or updates an item in the cache with the specified key, data, and lifespan.
// If intervalLifeSpan is true, the item's expiration time will be extended each time it is accessed.
// Returns the newly created or updated cache item.
func (table *Table) SetRaw(key string, data interface{}, lifeSpan time.Duration,
	intervalLifeSpan ...bool) *Item {
	item := NewCacheItem(key, data, lifeSpan)

	if len(intervalLifeSpan) > 0 && intervalLifeSpan[0] {
		table.accessCount.Store(true)
		item.intervalLifeSpan = intervalLifeSpan[0]
	}
	table.addInternal(item)

	return item
}

// Set adds or updates an item in the cache with the specified key, data, and lifespan in seconds.
// If interval is true, the item's expiration time will be extended each time it is accessed.
// Returns the newly created or updated cache item.
func (table *Table) Set(key string, data interface{}, lifeSpanSecond uint,
	interval ...bool) *Item {
	return table.SetRaw(key, data, time.Duration(lifeSpanSecond)*time.Second, interval...)
}

// deleteInternal removes an item from the cache and invokes any associated delete callbacks.
// Returns the removed item and any error that occurred during the operation.
func (table *Table) deleteInternal(key string) (*Item, error) {
	r, ok := table.items[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	deleteCallback := table.deleteCallback
	table.Unlock()
	if deleteCallback != nil && !deleteCallback(r.key) {
		table.Lock()
		r.RLock()
		r.accessedTime = ztime.UnixMicro(ztime.Clock())
		r.RUnlock()
		return r, nil
	}

	r.RLock()
	defer r.RUnlock()
	if r.deleteCallback != nil && !r.deleteCallback(r.key) {
		table.Lock()
		r.RLock()
		r.accessedTime = ztime.UnixMicro(ztime.Clock())
		r.RUnlock()
		return r, nil
	}

	table.Lock()
	delete(table.items, key)
	return r, nil
}

// Delete removes an item with the specified key from the cache.
// Returns the removed item and any error that occurred during the operation.
func (table *Table) Delete(key string) (*Item, error) {
	table.Lock()
	defer table.Unlock()

	return table.deleteInternal(key)
}

// Exists checks if an item with the specified key exists in the cache.
// Returns true if the item exists, false otherwise.
func (table *Table) Exists(key string) bool {
	table.RLock()
	defer table.RUnlock()
	_, ok := table.items[key]

	return ok
}

// Add adds a new item to the cache only if the key does not already exist.
// Returns true if the item was added, false if the key already exists.
// If intervalLifeSpan is true, the item's expiration time will be extended each time it is accessed.
func (table *Table) Add(key string, data interface{}, lifeSpan time.Duration, intervalLifeSpan ...bool) bool {
	table.Lock()
	_, ok := table.items[key]
	table.Unlock()
	if ok {
		return false
	}

	item := NewCacheItem(key, data, lifeSpan)
	if len(intervalLifeSpan) > 0 {
		item.intervalLifeSpan = intervalLifeSpan[0]
	}
	table.addInternal(item)

	return true
}

// MustGet retrieves an item from the cache or creates it if it doesn't exist.
// If the item doesn't exist, the provided function is called to generate the data,
// which is then stored in the cache with the specified parameters.
// Returns the item's data and any error that occurred during the operation.
func (table *Table) MustGet(key string, do func(set func(data interface{},
	lifeSpan time.Duration, interval ...bool)) (
	err error)) (data interface{}, err error) {
	table.Lock()
	r, ok := table.items[key]
	if ok {
		table.Unlock()
		r.keepAlive()
		return r.Data(), nil
	}
	item := NewCacheItem(key, "", 0)
	item.Lock()
	table.items[key] = item
	table.Unlock()
	err = do(func(data interface{},
		lifeSpan time.Duration, interval ...bool) {
		item.data = data
		item.lifeSpan = lifeSpan
		if len(interval) > 0 {
			item.intervalLifeSpan = interval[0]
		}
	})
	item.Unlock()

	if err != nil {
		table.Lock()
		delete(table.items, key)
		table.Unlock()
		return
	}

	data = item.data
	table.addInternal(item)
	return
}

// GetT retrieves a cache item by its key.
// If the item exists, its access time is updated if access counting is enabled.
// If the item doesn't exist and a load callback is set, it attempts to load the item.
// Returns the item and any error that occurred during the operation.
func (table *Table) GetT(key string, args ...interface{}) (*Item, error) {
	table.RLock()
	r, ok := table.items[key]
	table.RUnlock()

	if ok {
		if table.accessCount.Load() {
			r.keepAlive()
		}
		return r, nil
	}

	loadData := table.loadNotCallback
	if loadData != nil {
		item := loadData(key, args...)
		if item != nil {
			table.SetRaw(key, item.data, item.lifeSpan)
			return item, nil
		}

		return nil, ErrKeyNotFoundAndNotCallback
	}

	return nil, ErrKeyNotFound
}

// Get retrieves the data associated with the specified key.
// Returns the data value and any error that occurred during the operation.
func (table *Table) Get(key string, args ...interface{}) (value interface{}, err error) {
	var data *Item
	data, err = table.GetT(key, args...)
	if err != nil {
		return
	}
	value = data.Data()
	return
}

// GetString retrieves the data associated with the specified key as a string.
// Returns the string value and any error that occurred during the operation.
func (table *Table) GetString(key string, args ...interface{}) (value string, err error) {
	data, err := table.Get(key, args...)
	if err != nil {
		return
	}
	value, _ = data.(string)
	return
}

// GetInt retrieves the data associated with the specified key as an integer.
// Returns the integer value and any error that occurred during the operation.
func (table *Table) GetInt(key string, args ...interface{}) (value int, err error) {
	data, err := table.Get(key, args...)
	if err != nil {
		return
	}
	value, _ = data.(int)

	return
}

// Clear removes all items from the cache and stops any cleanup timers.
func (table *Table) Clear() {
	table.Lock()
	table.items = make(map[string]*Item)
	table.cleanupInterval = 0
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
	table.Unlock()
}

// Swap implements sort.Interface for CacheItemPairList
func (p CacheItemPairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
// Len implements sort.Interface for CacheItemPairList
func (p CacheItemPairList) Len() int           { return len(p) }
// Less implements sort.Interface for CacheItemPairList, sorting by access count in descending order
func (p CacheItemPairList) Less(i, j int) bool { return p[i].AccessCount > p[j].AccessCount }

// MostAccessed returns the most frequently accessed items in the cache.
// The count parameter specifies the maximum number of items to return.
// Items are sorted by access count in descending order.
func (table *Table) MostAccessed(count int64) []*Item {
	table.RLock()
	defer table.RUnlock()

	p := make(CacheItemPairList, len(table.items))
	i := 0
	for k, v := range table.items {
		p[i] = CacheItemPair{k, v.accessCount}
		i++
	}
	sort.Sort(p)

	var r []*Item
	c := int64(0)
	for _, v := range p {
		if c >= count {
			break
		}

		item, ok := table.items[v.Key]
		if ok {
			r = append(r, item)
		}
		c++
	}

	return r
}
