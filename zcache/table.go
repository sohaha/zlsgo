/*
 * @Author: seekwe
 * @Date:   2019-05-24 19:18:58
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-25 13:47:29
 */

package zcache

import (
	"log"
	"sort"
	"sync"
	"time"
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
		items map[interface{}]*CacheItem

		cleanupTimer    *time.Timer
		cleanupInterval time.Duration

		logger *log.Logger

		loadNotCallback func(key interface{}, args ...interface{}) *CacheItem
		addCallback     func(item *CacheItem)
		deleteCallback  func(item *CacheItem)
	}
)

// Count Count
func (table *Table) Count() int {
	table.RLock()
	defer table.RUnlock()
	return len(table.items)
}

// ForEach ForEach
func (table *Table) ForEach(trans func(key interface{}, item *CacheItem)) {
	table.RLock()
	defer table.RUnlock()

	for k, v := range table.items {
		trans(k, v)
	}
}

// SetLoadNotCallback SetLoadNotCallback
func (table *Table) SetLoadNotCallback(f func(interface{}, ...interface{}) *CacheItem) {
	table.Lock()
	defer table.Unlock()
	table.loadNotCallback = f
}

// SetAddCallback SetAddCallback
func (table *Table) SetAddCallback(f func(*CacheItem)) {
	table.Lock()
	defer table.Unlock()
	table.addCallback = f
}

// SetDeleteCallback SetDeleteCallback
func (table *Table) SetDeleteCallback(f func(*CacheItem)) {
	table.Lock()
	defer table.Unlock()
	table.deleteCallback = f
}

// SetLogger SetLogger
func (table *Table) SetLogger(logger *log.Logger) {
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
		if now.Sub(accessedOn) >= lifeSpan {
			table.deleteInternal(key)
		} else {
			if smallestDuration == 0 || lifeSpan-now.Sub(accessedOn) < smallestDuration {
				smallestDuration = lifeSpan - now.Sub(accessedOn)
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

func (table *Table) addInternal(item *CacheItem) {
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

// Set Set
func (table *Table) Set(key interface{}, lifeSpan time.Duration, data interface{}) *CacheItem {
	item := newCacheItem(key, lifeSpan, data)
	table.Lock()
	table.addInternal(item)

	return item
}

func (table *Table) deleteInternal(key interface{}) (*CacheItem, error) {
	r, ok := table.items[key]
	if !ok {
		return nil, ErrKeyNotFound
	}
	deleteCallback := table.deleteCallback
	table.Unlock()

	if deleteCallback != nil {
		deleteCallback(r)
	}

	r.RLock()
	defer r.RUnlock()
	if r.deleteCallback != nil {
		r.deleteCallback(key)
	}

	table.Lock()
	table.log("Deleting item with key", key, "created on", r.createdTime, "and hit", r.accessCount, "times from table", table.name)
	delete(table.items, key)

	return r, nil
}

// Delete Delete
func (table *Table) Delete(key interface{}) (*CacheItem, error) {
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

// NotFoundAdd NotFoundAdd
func (table *Table) NotFoundAdd(key interface{}, lifeSpan time.Duration, data interface{}) bool {
	table.Lock()

	if _, ok := table.items[key]; ok {
		table.Unlock()
		return false
	}

	item := newCacheItem(key, lifeSpan, data)
	table.addInternal(item)

	return true
}

// Get Get
func (table *Table) Get(key interface{}, args ...interface{}) (*CacheItem, error) {
	table.RLock()
	r, ok := table.items[key]
	loadData := table.loadNotCallback
	table.RUnlock()

	if ok {
		r.KeepAlive()
		return r, nil
	}

	if loadData != nil {
		item := loadData(key, args...)
		if item != nil {
			table.Set(key, item.lifeSpan, item.data)
			return item, nil
		}

		return nil, ErrKeyNotFoundAndNotCallback
	}

	return nil, ErrKeyNotFound
}

// Clear Clear
func (table *Table) Clear() {
	table.Lock()
	defer table.Unlock()

	table.log("Flushing table", table.name)

	table.items = make(map[interface{}]*CacheItem)
	table.cleanupInterval = 0
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
}

func (p CacheItemPairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p CacheItemPairList) Len() int           { return len(p) }
func (p CacheItemPairList) Less(i, j int) bool { return p[i].AccessCount > p[j].AccessCount }

// MostAccessed MostAccessed
func (table *Table) MostAccessed(count int64) []*CacheItem {
	table.RLock()
	defer table.RUnlock()

	p := make(CacheItemPairList, len(table.items))
	i := 0
	for k, v := range table.items {
		p[i] = CacheItemPair{k, v.accessCount}
		i++
	}
	sort.Sort(p)

	var r []*CacheItem
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

	table.logger.Println(v...)
}
