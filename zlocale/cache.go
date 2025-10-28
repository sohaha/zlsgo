package zlocale

import "time"

// TemplateCache defines the interface for template caching functionality
// This abstraction allows for different cache implementations while maintaining
// a consistent API for the i18n module
type TemplateCache interface {
	// Get retrieves a cached template by its key
	Get(key string) (*TemplateCacheEntry, bool)

	// Set stores a template with optional expiration
	Set(key string, template *TemplateCacheEntry)

	// Delete removes a template from the cache
	Delete(key string)

	// Clear removes all templates from the cache
	Clear()

	// Count returns the number of cached templates
	Count() int

	// Stats returns cache statistics for monitoring
	Stats() CacheStats

	// Close cleans up cache resources
	Close()
}

// TemplateCacheEntry represents a cached template with metadata
type TemplateCacheEntry struct {
	Template interface{} // The actual template (e.g., *zstring.Template)
	Created  time.Time   // Creation timestamp
	Accessed time.Time   // Last access timestamp
	Hits     int64       // Access count
}

// CacheStats provides statistics about cache performance
type CacheStats struct {
	// TotalItems is the number of items currently in the cache
	TotalItems int

	// HitCount is the total number of successful cache hits
	HitCount int64

	// MissCount is the total number of cache misses
	MissCount int64

	// HitRate is the cache hit ratio (0-1)
	HitRate float64

	// MemoryUsage is the estimated memory usage in bytes
	MemoryUsage int64

	// EvictionCount is the number of items evicted due to expiration or capacity limits
	EvictionCount int64

	// IdleDuration shows how long the cache has been idle
	IdleDuration time.Duration

	// CleanerLevel indicates the current cleaning intensity level
	CleanerLevel int32

	// IsCleanerRunning indicates if background cleaner is active
	IsCleanerRunning bool
}
