package zcache

import (
	"sort"
	"sync"
	"time"

	"github.com/sohaha/zlsgo/zlog"
)

type (
	// CacheItemPair maps key to access counter
	CacheItemPair struct {
		Key         interface{}
		AccessCount int64
	}
	// CacheItemPairList CacheItemPairList
	CacheItemPairList []CacheItemPair

	// Table Table
	Table struct {
		sync.RWMutex

		name  string
		items map[interface{}]*Item

		cleanupTimer    *time.Timer
		cleanupInterval time.Duration

		logger *zlog.Logger

		loadNotCallback func(key interface{}, args ...interface{}) *Item
		addCallback     func(item *Item)
		deleteCallback  func(item *Item) bool
	}
)

// Count get the number of caches
func (table *Table) Count() int {
	table.RLock()
	defer table.RUnlock()
	return len(table.items)
}

// ForEach traversing the cache
func (table *Table) ForEach(trans func(key, value interface{})) {
	table.RLock()
	items := make(map[interface{}]interface{}, table.Count())
	for k, v := range table.items {
		items[k] = v.Data()
	}
	table.RUnlock()

	for k, v := range items {
		trans(k, v)
	}
}

// SetLoadNotCallback SetLoadNotCallback
func (table *Table) SetLoadNotCallback(f func(key interface{}, args ...interface{}) *Item) {
	table.Lock()
	defer table.Unlock()
	table.loadNotCallback = f
}

// SetAddCallback SetAddCallback
func (table *Table) SetAddCallback(f func(*Item)) {
	table.Lock()
	defer table.Unlock()
	table.addCallback = f
}

// SetDeleteCallback SetDeleteCallback
func (table *Table) SetDeleteCallback(f func(*Item) bool) {
	table.Lock()
	defer table.Unlock()
	table.deleteCallback = f
}

// SetLogger SetLogger
func (table *Table) SetLogger(logger *zlog.Logger) {
	table.Lock()
	defer table.Unlock()
	table.logger = logger
}

func (table *Table) expirationCheck() {
	table.Lock()
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
	if table.cleanupInterval > 0 {
		table.log("Expiration check triggered after", table.cleanupInterval, "for table", table.name)
	} else {
		table.log("Expiration check installed for table", table.name)
	}

	now := time.Now()
	smallestDuration := 0 * time.Second
	for key, item := range table.items {
		item.RLock()
		lifeSpan := item.lifeSpan
		accessedOn := item.accessedTime
		item.RUnlock()

		if lifeSpan == 0 {
			continue
		}
		if item.intervalLifeSpan {
			if now.Sub(accessedOn) >= lifeSpan {
				_, _ = table.deleteInternal(key)
			} else {
				nextDuration := lifeSpan - now.Sub(accessedOn)
				if smallestDuration == 0 || nextDuration < smallestDuration {
					smallestDuration = nextDuration
				}
			}
		} else if item.RemainingLife() <= 0 {
			_, _ = table.deleteInternal(key)
		} else {
			remainingLift := item.RemainingLife()
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

func (table *Table) addInternal(item *Item) {
	table.log("Adding item with key", item.key, "and lifespan of", item.lifeSpan, "to table", table.name)
	table.items[item.key] = item

	expDur := table.cleanupInterval
	addedItem := table.addCallback
	table.Unlock()

	if addedItem != nil {
		addedItem(item)
	}

	if item.lifeSpan > 0 && (expDur == 0 || item.lifeSpan < expDur) {
		table.expirationCheck()
	}
}

// SetRaw set cache
func (table *Table) SetRaw(key, data interface{}, lifeSpan time.Duration, intervalLifeSpan ...bool) *Item {
	item := newCacheItem(key, data, lifeSpan)
	if len(intervalLifeSpan) > 0 {
		item.intervalLifeSpan = intervalLifeSpan[0]
	}
	table.Lock()
	table.addInternal(item)

	return item
}

// Set set cache whether to automatically renew
func (table *Table) Set(key, data interface{}, lifeSpan uint, interval ...bool) *Item {
	return table.SetRaw(key, data, time.Duration(lifeSpan)*time.Second, interval...)
}

func (table *Table) deleteInternal(key interface{}) (*Item, error) {
	r, ok := table.items[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	table.log("Deleting item with key", key, "created on", r.createdTime, "and hit", r.accessCount, "times from table", table.name)
	deleteCallback := table.deleteCallback
	table.Unlock()

	if deleteCallback != nil && !deleteCallback(r) {
		table.Lock()
		r.accessedTime = time.Now()
		return r, nil
	}

	r.RLock()
	defer r.RUnlock()
	if r.deleteCallback != nil && !r.deleteCallback(r) {
		table.Lock()
		r.accessedTime = time.Now()
		return r, nil
	}

	table.Lock()
	delete(table.items, key)

	return r, nil
}

// Delete Delete cache
func (table *Table) Delete(key interface{}) (*Item, error) {
	table.Lock()
	defer table.Unlock()

	return table.deleteInternal(key)
}

// Exists Exists
func (table *Table) Exists(key interface{}) bool {
	table.RLock()
	defer table.RUnlock()
	_, ok := table.items[key]

	return ok
}

// Add if the cache does not exist then adding does not take effect
func (table *Table) Add(key interface{}, data interface{}, lifeSpan time.Duration, intervalLifeSpan ...bool) bool {
	table.Lock()

	if _, ok := table.items[key]; ok {
		table.Unlock()
		return false
	}

	item := newCacheItem(key, data, lifeSpan)
	if len(intervalLifeSpan) > 0 {
		item.intervalLifeSpan = intervalLifeSpan[0]
	}
	table.addInternal(item)

	return true
}

// GetLocked if the cache is not obtained, it will be locked directly
func (table *Table) GetLocked(key interface{}) (interface{}, func(data interface{}, lifeSpan uint, interval ...bool)) {
	table.RLock()
	r, ok := table.items[key]
	table.RUnlock()
	if ok {
		r.keepAlive()
		return r.Data(), nil
	}
	tmp := table.SetRaw(key, "", 0)
	tmp.Lock()
	return nil, func(data interface{}, lifeSpan uint, interval ...bool) {
		tmp.Unlock()
		item := newCacheItem(key, data, time.Duration(lifeSpan)*time.Second)
		if len(interval) > 0 {
			item.intervalLifeSpan = interval[0]
		}
		table.Lock()
		table.addInternal(item)
	}
}

// GetRaw GetRaw
func (table *Table) GetRaw(key interface{}, args ...interface{}) (*Item, error) {
	table.RLock()
	r, ok := table.items[key]
	table.RUnlock()

	if ok {
		r.keepAlive()
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

func (table *Table) Get(key interface{}, args ...interface{}) (value interface{}, err error) {
	data, err := table.GetRaw(key, args...)
	if err != nil {
		return
	}
	value = data.Data()
	return
}

// Clear Clear
func (table *Table) Clear() {
	table.Lock()
	defer table.Unlock()

	table.log("Flushing table", table.name)

	table.items = make(map[interface{}]*Item)
	table.cleanupInterval = 0
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
}

func (p CacheItemPairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p CacheItemPairList) Len() int           { return len(p) }
func (p CacheItemPairList) Less(i, j int) bool { return p[i].AccessCount > p[j].AccessCount }

// MostAccessed MostAccessed
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

func (table *Table) log(v ...interface{}) {
	if table.logger == nil {
		return
	}

	table.logger.Debug(v...)
}
