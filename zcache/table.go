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
	// CacheItemPair maps key to access counter
	CacheItemPair struct {
		Key         string
		AccessCount int64
	}
	// CacheItemPairList CacheItemPairList
	CacheItemPairList []CacheItemPair
	// Table Table
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

// Count get the number of caches
func (table *Table) Count() int {
	table.RLock()
	defer table.RUnlock()
	return len(table.items)
}

// ForEach traversing the cache
func (table *Table) ForEach(trans func(key string, value interface{}) bool) {
	table.ForEachRaw(func(k string, v *Item) bool {
		return trans(k, v.Data())
	})
}

// ForEachRaw traversing the cache
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

// SetLoadNotCallback SetLoadNotCallback
func (table *Table) SetLoadNotCallback(f func(key string, args ...interface{}) *Item) {
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
func (table *Table) SetDeleteCallback(f func(key string) bool) {
	table.Lock()
	defer table.Unlock()
	table.deleteCallback = f
}

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

// SetRaw set cache
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

// Set set cache whether to automatically renew
func (table *Table) Set(key string, data interface{}, lifeSpanSecond uint,
	interval ...bool) *Item {
	return table.SetRaw(key, data, time.Duration(lifeSpanSecond)*time.Second, interval...)
}

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

// Delete Delete cache
func (table *Table) Delete(key string) (*Item, error) {
	table.Lock()
	defer table.Unlock()

	return table.deleteInternal(key)
}

// Exists Exists
func (table *Table) Exists(key string) bool {
	table.RLock()
	defer table.RUnlock()
	_, ok := table.items[key]

	return ok
}

// Add if the cache does not exist then adding does not take effect
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

// MustGet get the Raw of the specified key, set if it does not exist
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

// GetT GetT
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

// Get get the Raw of the specified key
func (table *Table) Get(key string, args ...interface{}) (value interface{}, err error) {
	var data *Item
	data, err = table.GetT(key, args...)
	if err != nil {
		return
	}
	value = data.Data()
	return
}

func (table *Table) GetString(key string, args ...interface{}) (value string, err error) {
	data, err := table.Get(key, args...)
	if err != nil {
		return
	}
	value, _ = data.(string)
	return
}

func (table *Table) GetInt(key string, args ...interface{}) (value int, err error) {
	data, err := table.Get(key, args...)
	if err != nil {
		return
	}
	value, _ = data.(int)

	return
}

// Clear Clear
func (table *Table) Clear() {
	table.Lock()
	table.items = make(map[string]*Item)
	table.cleanupInterval = 0
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
	table.Unlock()
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
