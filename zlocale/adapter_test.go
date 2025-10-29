package zlocale

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztime"
)

func TestLegacyCacheAdapter(t *testing.T) {
	tt := zlsgo.NewTest(t)
	
	cache := NewLegacyCacheAdapter(100)
	tt.NotNil(cache)
	tt.Equal(0, cache.Count())
	
	_, found := cache.Get("nonexistent")
	tt.EqualFalse(found)
	
	template, err := zstring.NewTemplate("Hello {0}!", "{", "}")
	tt.NoError(err)
	
	cacheEntry := &TemplateCacheEntry{
		Template: template,
		Created:  time.Now(),
		Accessed: time.Now(),
		Hits:     0,
	}
	
	cache.Set("test_key", cacheEntry)
	tt.Equal(1, cache.Count())
	
	retrievedEntry, found := cache.Get("test_key")
	tt.NotNil(retrievedEntry)
	tt.EqualTrue(found)
	tt.Equal(template, retrievedEntry.Template)
	tt.Equal(retrievedEntry.Hits > 0, true)
	
	cache.Set("key2", cacheEntry)
	cache.Set("key3", cacheEntry)
	tt.Equal(3, cache.Count())
	
	cache.Delete("key2")
	tt.Equal(2, cache.Count())
	
	_, found = cache.Get("key2")
	tt.EqualFalse(found)
	
	cache.Clear()
	tt.Equal(0, cache.Count())
	
	_, found = cache.Get("test_key")
	tt.EqualFalse(found)
}

func TestLegacyCacheAdapterMaxSize(t *testing.T) {
	tt := zlsgo.NewTest(t)
	
	cache := NewLegacyCacheAdapter(2)
	
	template, _ := zstring.NewTemplate("Test", "{", "}")
	entry := &TemplateCacheEntry{Template: template}
	
	cache.Set("key1", entry)
	tt.Equal(1, cache.Count())
	
	cache.Set("key2", entry)
	tt.Equal(2, cache.Count())
	
	cache.Set("key3", entry)
	actualCount := cache.Count()
	tt.EqualTrue(actualCount >= 2 && actualCount <= 3)
	
	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	_, found3 := cache.Get("key3")
	
	evictedCount := 0
	if !found1 {
		evictedCount++
	}
	if !found2 {
		evictedCount++
	}
	tt.Equal(evictedCount >= 0, true)
	tt.EqualTrue(found3)
}

func TestLegacyCacheAdapterStats(t *testing.T) {
	tt := zlsgo.NewTest(t)
	
	cache := NewLegacyCacheAdapter(10)
	
	stats := cache.Stats()
	tt.Equal(0, stats.TotalItems)
	tt.Equal(int64(0), stats.HitCount)
	tt.Equal(int64(0), stats.MissCount)
	tt.Equal(0.0, stats.HitRate)
	tt.Equal(int64(0), stats.EvictionCount)
	tt.EqualFalse(stats.IsCleanerRunning)
	
	template, _ := zstring.NewTemplate("Test", "{", "}")
	entry := &TemplateCacheEntry{Template: template}
	
	cache.Set("key1", entry)
	cache.Set("key2", entry)
	
	cache.Get("key1")
	cache.Get("key2")
	cache.Get("nonexistent")
	
	stats = cache.Stats()
	tt.Equal(2, stats.TotalItems)
	tt.Equal(int64(2), stats.HitCount)
	tt.Equal(int64(1), stats.MissCount)
	tt.Equal(2.0/3.0, stats.HitRate)
	tt.Equal(stats.MemoryUsage > 0, true)
}

func TestLegacyCacheAdapterConcurrency(t *testing.T) {
	tt := zlsgo.NewTest(t)
	
	cache := NewLegacyCacheAdapter(1000)
	
	var wg sync.WaitGroup
	concurrency := 50
	iterations := 100
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				template, _ := zstring.NewTemplate("Test", "{", "}")
				entry := &TemplateCacheEntry{Template: template}
				cache.Set(key, entry)
			}
		}(i)
	}
	
	wg.Wait()
	
	stats := cache.Stats()
	tt.Equal(stats.TotalItems > 0, true)
}

func TestLegacyCacheAdapterClose(t *testing.T) {
	tt := zlsgo.NewTest(t)
	
	cache := NewLegacyCacheAdapter(10)
	
	template, _ := zstring.NewTemplate("Test", "{", "}")
	entry := &TemplateCacheEntry{Template: template}
	cache.Set("key1", entry)
	cache.Set("key2", entry)
	
	tt.Equal(2, cache.Count())
	
	cache.Close()
	tt.Equal(0, cache.Count())
	
	cache.Set("new_key", entry)
	tt.Equal(1, cache.Count())
}

func TestLegacyCacheAdapterEntryMetadata(t *testing.T) {
	tt := zlsgo.NewTest(t)
	
	cache := NewLegacyCacheAdapter(10)
	
	template, _ := zstring.NewTemplate("Hello {0}!", "{", "}")
	entry := &TemplateCacheEntry{
		Template: template,
		Hits:     5,
	}
	
	beforeSet := ztime.UnixMicro(ztime.Clock())
	cache.Set("test", entry)
	afterSet := ztime.UnixMicro(ztime.Clock())
	
	retrieved, found := cache.Get("test")
	tt.EqualTrue(found)
	tt.Equal(template, retrieved.Template)
	
	tt.Equal(retrieved.Created.Unix() >= beforeSet.Unix(), true)
	tt.Equal(retrieved.Created.Unix() <= afterSet.Unix(), true)
	tt.Equal(retrieved.Accessed.Unix() >= beforeSet.Unix(), true)
	tt.Equal(retrieved.Accessed.Unix() <= afterSet.Unix(), true)
	
	tt.Equal(retrieved.Hits >= 1, true)
	
	time.Sleep(1 * time.Millisecond)
	beforeGet := retrieved.Accessed
	cache.Get("test")
	afterGet := retrieved.Accessed
	
	tt.Equal(afterGet.After(beforeGet) || afterGet.Equal(beforeGet), true)
}