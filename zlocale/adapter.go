package zlocale

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/sohaha/zlsgo/ztime"
)

// LegacyCacheAdapter implements TemplateCache interface using the original map-based approach
// This maintains backward compatibility and provides a fallback option
type LegacyCacheAdapter struct {
	cache     map[string]*TemplateCacheEntry
	hitCount  int64
	missCount int64
	maxSize   int
	mutex     sync.RWMutex
}

// NewLegacyCacheAdapter creates a new legacy map-based cache adapter
func NewLegacyCacheAdapter(maxSize int) *LegacyCacheAdapter {
	return &LegacyCacheAdapter{
		cache:   make(map[string]*TemplateCacheEntry),
		maxSize: maxSize,
	}
}

// Get retrieves a cached template entry by key
func (l *LegacyCacheAdapter) Get(key string) (*TemplateCacheEntry, bool) {
	l.mutex.RLock()
	entry, found := l.cache[key]
	if found {
		entry.Accessed = ztime.UnixMicro(ztime.Clock())
		atomic.AddInt64(&entry.Hits, 1)
	}
	l.mutex.RUnlock()

	if !found {
		atomic.AddInt64(&l.missCount, 1)
		return nil, false
	}

	atomic.AddInt64(&l.hitCount, 1)
	return entry, true
}

// Set stores a template entry with optional expiration
func (l *LegacyCacheAdapter) Set(key string, entry *TemplateCacheEntry) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	now := ztime.UnixMicro(ztime.Clock())
	if entry.Created.IsZero() {
		entry.Created = now
	}
	entry.Accessed = now
	if entry.Hits == 0 {
		entry.Hits = 1
	}

	if len(l.cache) >= l.maxSize {
		var oldestKey string
		var oldestTime time.Time = now

		for k, v := range l.cache {
			if v.Accessed.Before(oldestTime) {
				oldestTime = v.Accessed
				oldestKey = k
			}
		}

		if oldestKey != "" {
			delete(l.cache, oldestKey)
		}
	}

	l.cache[key] = entry
}

// Delete removes a template from the cache
func (l *LegacyCacheAdapter) Delete(key string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	delete(l.cache, key)
}

// Clear removes all templates from the cache
func (l *LegacyCacheAdapter) Clear() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.cache = make(map[string]*TemplateCacheEntry)
}

// Count returns the number of cached templates
func (l *LegacyCacheAdapter) Count() int {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return len(l.cache)
}

// Stats returns cache statistics
func (l *LegacyCacheAdapter) Stats() CacheStats {
	hits := atomic.LoadInt64(&l.hitCount)
	misses := atomic.LoadInt64(&l.missCount)
	total := hits + misses

	hitRate := 0.0
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	return CacheStats{
		TotalItems:       l.Count(),
		HitCount:         hits,
		MissCount:        misses,
		HitRate:          hitRate,
		EvictionCount:    0, // Legacy adapter doesn't track evictions
		CleanerLevel:     0,
		IdleDuration:     0,
		IsCleanerRunning: false,
		MemoryUsage:      int64(l.Count()) * 1024, // Rough estimate
	}
}

// Close cleans up cache resources
func (l *LegacyCacheAdapter) Close() {
	l.Clear()
}
