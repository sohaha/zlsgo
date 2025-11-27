package fast

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/sohaha/zlsgo/ztime"
	"golang.org/x/sync/singleflight"
)

// FastCache implements a high-performance concurrent LRU (Least Recently Used) cache
// with support for expiration, callbacks, and multiple buckets for reduced lock contention.
type FastCache struct {
	gsf                 singleflight.Group
	ticker              *time.Ticker
	callback            handler
	stopCh              chan struct{}
	insts               [][2]*lruCache
	locks               []sync.Mutex
	longIdleThreshold   time.Duration // Threshold for level 3 (default: 5m)
	mediumIdleThreshold time.Duration // Threshold for level 2 (default: 2m)
	shortIdleThreshold  time.Duration // Threshold for level 1 (default: 30s)
	cleanInterval       time.Duration
	expiration          time.Duration
	idleAfter           time.Duration
	lastActiveMs        int64
	accessCount         int64 // Access counter for activity tracking (atomic access)
	lastAccessMs        int64 // Last access timestamp for Set/Get/Delete operations (atomic access)
	cleanerMu           sync.Mutex
	cleanerLevel        int32 // Cleaner level: 0=normal, 1=reduced_freq, 2=light, 3=stopped (atomic access)
	cleanIdx            uint16
	mask                uint16
	lazyCleaner         bool
	autoCleaner         bool
	cleanerOn           bool
}

// NewFast creates a new FastCache instance with the specified options.
// If no options are provided, default values are used.
func NewFast(opt ...func(o *Options)) *FastCache {
	o := Options{
		Cap:                 1 << 10,
		Bucket:              4,
		AutoCleaner:         false,
		LazyCleaner:         true,
		IdleAfter:           30 * time.Second,
		ShortIdleThreshold:  30 * time.Second,
		MediumIdleThreshold: 1 * time.Minute,
		LongIdleThreshold:   2 * time.Minute,
	}

	for _, f := range opt {
		f(&o)
	}

	var mask uint16
	if o.Bucket > 0 && o.Bucket&(o.Bucket-1) == 0 {
		mask = o.Bucket - 1
	} else {
		o.Bucket |= o.Bucket >> 1
		o.Bucket |= o.Bucket >> 2
		o.Bucket |= o.Bucket >> 4
		mask = o.Bucket | (o.Bucket >> 8)
	}
	c := &FastCache{
		locks:               make([]sync.Mutex, mask+1),
		insts:               make([][2]*lruCache, mask+1),
		expiration:          o.Expiration,
		mask:                mask,
		callback:            o.Callback,
		autoCleaner:         o.AutoCleaner,
		lazyCleaner:         o.LazyCleaner,
		idleAfter:           o.IdleAfter,
		shortIdleThreshold:  o.ShortIdleThreshold,
		mediumIdleThreshold: o.MediumIdleThreshold,
		longIdleThreshold:   o.LongIdleThreshold,
	}
	for i := range c.insts {
		c.insts[i][0] = &lruCache{dlList: make([][2]uint16, uint32(o.Cap)+1), nodes: make([]node, o.Cap), hashmap: make(map[string]uint16, o.Cap), last: 0}
		if o.LRU2Cap > 0 {
			c.insts[i][1] = &lruCache{dlList: make([][2]uint16, uint32(o.LRU2Cap)+1), nodes: make([]node, o.LRU2Cap), hashmap: make(map[string]uint16, o.LRU2Cap), last: 0}
		}
	}

	if c.expiration > 0 && c.autoCleaner {
		c.cleanInterval = c.expiration
		// Initialize timestamps to current time to avoid zero-value issues
		now := ztime.Clock() * 1000
		atomic.StoreInt64(&c.lastAccessMs, now)
		atomic.StoreInt64(&c.lastActiveMs, now)

		if !c.lazyCleaner {
			c.startCleaner()
		}
		// Set finalizer as a safety net to prevent memory leaks
		// if user forgets to call Close()
		runtime.SetFinalizer(c, (*FastCache).finalize)
	}
	return c
}

// Options defines configuration parameters for creating a new FastCache instance
type Options struct {
	// Callback is called when items are accessed or modified in the cache
	Callback func(ActionKind, string, uintptr)
	// Expiration is the default expiration time for cache items
	Expiration time.Duration
	// Bucket is the number of shards to divide the cache into for better concurrency
	Bucket uint16
	// Cap is the maximum capacity of the primary LRU cache per bucket
	Cap uint16
	// LRU2Cap is the capacity of the secondary LRU cache per bucket (for multi-level LRU)
	LRU2Cap uint16
	// AutoCleaner enables background cleaner when Expiration>0 (default: false)
	AutoCleaner bool
	// LazyCleaner delays starting the cleaner until first activity (default: true)
	LazyCleaner bool
	// IdleAfter >0 enables idle self-stop when cache remains empty and inactive for this duration
	IdleAfter time.Duration
	// Configurable intelligent cleaning thresholds (optional, defaults provided)
	// ShortIdleThreshold sets when to enter reduced frequency cleaning (default: 30s)
	ShortIdleThreshold time.Duration
	// MediumIdleThreshold sets when to enter light cleaning mode (default: 2m)
	MediumIdleThreshold time.Duration
	// LongIdleThreshold sets when to stop cleaning entirely (default: 5m)
	LongIdleThreshold time.Duration
}

// set is an internal method that adds or updates an item in the cache.
// It supports storing either an interface{} value or a byte slice.
func (l *FastCache) set(k string, v *interface{}, b []byte, expiration ...time.Duration) {
	if l.callback != nil {
		if v != nil {
			l.callback(SET, k, uintptr(unsafe.Pointer(v)))
		} else {
			l.callback(SET, k, uintptr(unsafe.Pointer(&b)))
		}
	}
	idx := hasher(k) & l.mask
	var expireAt int64
	if len(expiration) > 0 {
		if expiration[0] == -1 {
		} else if expiration[0] > 0 {
			expireAt = ztime.Clock()*1000 + int64(expiration[0])
		} else if l.expiration > 0 {
			expireAt = ztime.Clock()*1000 + int64(l.expiration)
		}
	} else if l.expiration > 0 {
		expireAt = ztime.Clock()*1000 + int64(l.expiration)
	}
	l.locks[idx].Lock()
	l.insts[idx][0].put(k, v, b, expireAt)
	l.locks[idx].Unlock()
	l.markActive()
}

// Set adds or updates an item in the cache with the specified key, value, and optional expiration.
// If no expiration is provided, the default expiration time is used (if configured).
func (l *FastCache) Set(key string, val interface{}, expiration ...time.Duration) {
	l.set(key, &val, nil, expiration...)
}

// SetBytes adds or updates a byte slice in the cache with the specified key.
// The default expiration time is used (if configured).
func (l *FastCache) SetBytes(key string, b []byte) {
	l.set(key, nil, b)
}

// Get retrieves an item from the cache by its key.
// Returns the item's value and a boolean indicating whether the item was found.
func (l *FastCache) Get(key string) (interface{}, bool) {
	if i, b, ok := l.get(key); ok {
		if i != nil {
			return *i, true
		}
		return b, true
	}
	return nil, false
}

// GetBytes retrieves a byte slice from the cache by its key.
// Returns the byte slice and a boolean indicating whether the item was found and is a byte slice.
func (l *FastCache) GetBytes(key string) ([]byte, bool) {
	if i, b, ok := l.get(key); ok {
		if b != nil {
			return b, true
		}
		b, ok = (*i).([]byte)
		return b, ok
	}
	return nil, false
}

// provideResult is returned through singleflight to propagate provider status.
type provideResult struct {
	value interface{}
	ok    bool
}

// ProvideGet retrieves an item from the cache, or computes and stores it if not present.
// If the item doesn't exist, the provide function is called to generate the value.
// Returns the item's value and a boolean indicating whether the item was found or created.
func (l *FastCache) ProvideGet(key string, provide func() (interface{}, bool), expiration ...time.Duration) (interface{}, bool) {
	if i, b, ok := l.get(key); ok {
		switch {
		case i != nil:
			return *i, true
		case b != nil:
			return b, true
		default:
			return nil, true
		}
	}

	v, err, _ := l.gsf.Do(key, func() (interface{}, error) {
		value, ok := provide()
		if ok {
			l.Set(key, value, expiration...)
		}
		return provideResult{value: value, ok: ok}, nil
	})
	if err != nil {
		return nil, false
	}

	if res, ok := v.(provideResult); ok {
		return res.value, res.ok
	}
	return v, true
}

// getValue is an internal method that retrieves a node from a specific cache level.
// It also handles expiration checking and marking expired items as deleted.
func (l *FastCache) getValue(key string, idx, level uint16) (*node, int) {
	n, s := l.insts[idx][level].get(key)
	if s > 0 {
		if !n.isDelete && (n.expireAt == 0 || (ztime.Clock()*1000 <= n.expireAt)) {
			return n, s
		}
		n.isDelete, n.value.value, n.value.byteValue = true, nil, nil
	}
	return nil, 0
}

// get is an internal method that retrieves an item from the cache.
// It handles the multi-level LRU logic and callback invocation.
func (l *FastCache) get(key string) (i *interface{}, b []byte, loaded bool) {
	idx := hasher(key) & l.mask
	l.locks[idx].Lock()
	n, s := (*node)(nil), 0
	if l.insts[idx][1] == nil {
		n, s = l.getValue(key, idx, 0)
	} else {
		e := int64(0)
		if n, s, e = l.insts[idx][0].delete(key); s <= 0 {
			n, s = l.getValue(key, idx, 1)
		} else {
			l.insts[idx][1].put(key, n.value.value, n.value.byteValue, e)
		}
	}
	if s <= 0 {
		l.locks[idx].Unlock()
		if l.callback != nil {
			l.callback(GET, key, uintptr(0))
		}
		return
	}
	i, b = n.value.value, n.value.byteValue
	l.locks[idx].Unlock()
	if l.callback != nil {
		if i != nil {
			l.callback(GET, key, uintptr(unsafe.Pointer(i)))
		} else {
			var b interface{} = b
			l.callback(GET, key, uintptr(unsafe.Pointer(&b)))
		}
	}
	return i, b, true
}

// Delete removes an item with the specified key from the cache.
// If the item doesn't exist, this operation is a no-op.
func (l *FastCache) Delete(key string) {
	idx := hasher(key) & l.mask
	l.locks[idx].Lock()
	n, s, e := l.insts[idx][0].delete(key)
	if l.insts[idx][1] != nil {
		if n2, s2, e2 := l.insts[idx][1].delete(key); n2 != nil && (n == nil || e < e2) {
			n, s = n2, s2
		}
	}
	if s > 0 {
		if l.callback != nil {
			if n.value.value != nil {
				l.callback(DELETE, key, uintptr(unsafe.Pointer(n.value.value)))
			} else {
				l.callback(DELETE, key, uintptr(unsafe.Pointer(&n.value.byteValue)))
			}
		}
		n.value.value, n.value.byteValue = nil, nil
	} else if l.callback != nil {
		l.callback(DELETE, key, uintptr(0))
	}

	l.locks[idx].Unlock()
}

// ForEach iterates through all items in the cache and applies the provided function to each key-value pair.
// The iteration continues as long as the function returns true, and stops when it returns false.
func (l *FastCache) ForEach(walker func(key string, iface interface{}) bool) {
	for i := range l.insts {
		l.locks[i].Lock()
		if l.insts[i][0].forEach(walker); l.insts[i][1] != nil {
			l.insts[i][1].forEach(walker)
		}
		l.locks[i].Unlock()
	}
}

func (l *FastCache) clean() {
	if len(l.insts) == 0 {
		return
	}

	now := ztime.Clock() * 1000

	// Optimized level calculation - avoid unnecessary computation
	currentLevel := atomic.LoadInt32(&l.cleanerLevel)

	// For performance, only recalculate level periodically or when activity changes
	// Check if we need to recalculate based on idle duration
	lastAccess := atomic.LoadInt64(&l.lastAccessMs)
	idleDuration := time.Duration(now-lastAccess) * time.Millisecond

	// Calculate target level using configurable thresholds
	var targetLevel int32
	switch {
	case idleDuration < l.shortIdleThreshold:
		targetLevel = 0 // Normal cleaning
	case idleDuration < l.mediumIdleThreshold:
		targetLevel = 1 // Reduced frequency (skip every other clean)
	case idleDuration < l.longIdleThreshold:
		targetLevel = 2 // Light cleaning (skip 4 out of 5 cleans)
	default:
		targetLevel = 3 // Stop cleaning
	}

	// Only update level if it actually changed - avoid unnecessary atomic write
	if currentLevel != targetLevel {
		atomic.StoreInt32(&l.cleanerLevel, targetLevel)
		currentLevel = targetLevel // Update local copy for following logic
	}

	// Stop cleaner if idle too long
	if currentLevel == 3 {
		l.stopCleaner()
		return
	}

	// Apply frequency reduction based on current level
	cleanCycle := uint16(1)
	switch currentLevel {
	case 1: // Reduced frequency: clean every 2nd cycle
		cleanCycle = 2
	case 2: // Light cleaning: clean every 5th cycle
		cleanCycle = 5
	}

	// Skip cleaning cycles based on level
	if currentLevel > 0 && (l.cleanIdx%cleanCycle) != 0 {
		l.cleanIdx++
		return
	}

	idx := l.cleanIdx & l.mask
	l.cleanIdx++
	l.locks[idx].Lock()
	if l.insts[idx][0] != nil {
		l.insts[idx][0].cleanExpired(now)
	}
	if l.insts[idx][1] != nil {
		l.insts[idx][1].cleanExpired(now)
	}
	l.locks[idx].Unlock()

	// Enhanced idle detection with additional safety checks
	if l.idleAfter > 0 && (l.cleanIdx&l.mask) == 0 {
		allEmpty := true
		for i := range l.insts {
			l.locks[i].Lock()
			if (l.insts[i][0] != nil && !l.insts[i][0].isEmpty()) || (l.insts[i][1] != nil && !l.insts[i][1].isEmpty()) {
				allEmpty = false
				l.locks[i].Unlock()
				break
			}
			l.locks[i].Unlock()
		}
		if allEmpty {
			lastActive := atomic.LoadInt64(&l.lastActiveMs)
			if lastActive > 0 && time.Duration(now-lastActive)*time.Millisecond >= l.idleAfter {
				l.stopCleaner()
			}
		}
	}
}

// Close stops the background cleaner if it is running.
// Enhanced with finalizer cleanup to optimize GC performance.
func (l *FastCache) Close() {
	// Mark as closed to prevent finalize from running
	// Use SwapInt32 to avoid race condition in CompareAndSwap
	oldLevel := atomic.SwapInt32(&l.cleanerLevel, -1)
	if oldLevel >= 0 {
		// Clear finalizer since we're properly closing manually
		// This reduces GC pressure and prevents unnecessary finalize calls
		runtime.SetFinalizer(l, nil)

		// Stop the cleaner
		l.stopCleaner()
	}

	// Ensure channel is closed for backward compatibility
	if l.stopCh != nil {
		select {
		case <-l.stopCh:
		default:
			close(l.stopCh)
		}
	}
}

// markActive records recent activity and triggers lazy cleaner start if needed.
// Enhanced with more precise activity tracking and optimized atomic operations.
func (l *FastCache) markActive() {
	if l.expiration <= 0 || !l.autoCleaner {
		return
	}
	now := ztime.Clock() * 1000

	// Batch update timestamps (most frequent operations)
	atomic.StoreInt64(&l.lastActiveMs, now)
	atomic.StoreInt64(&l.lastAccessMs, now)
	atomic.AddInt64(&l.accessCount, 1)

	// Optimized cleaner level reset - avoid unnecessary atomic operations
	// Only reset if level is elevated (most common case is level already 0)
	if atomic.LoadInt32(&l.cleanerLevel) > 0 {
		atomic.StoreInt32(&l.cleanerLevel, 0)
	}

	if l.lazyCleaner {
		l.startCleaner()
	}
}

// startCleaner starts the background cleaner if not already running.
func (l *FastCache) startCleaner() {
	l.cleanerMu.Lock()
	defer l.cleanerMu.Unlock()
	if l.cleanerOn || l.expiration <= 0 || !l.autoCleaner {
		return
	}
	if l.cleanInterval == 0 {
		l.cleanInterval = l.expiration
	}
	if l.stopCh == nil {
		l.stopCh = make(chan struct{})
	}
	l.ticker = time.NewTicker(l.cleanInterval)
	l.cleanerOn = true
	t := l.ticker
	ch := l.stopCh
	go func(tk *time.Ticker, stop <-chan struct{}) {
		for {
			select {
			case <-tk.C:
				l.clean()
			case <-stop:
				tk.Stop()
				return
			}
		}
	}(t, ch)
}

// stopCleaner stops the background cleaner if running.
func (l *FastCache) stopCleaner() {
	l.cleanerMu.Lock()
	defer l.cleanerMu.Unlock()
	if !l.cleanerOn {
		return
	}
	if l.stopCh != nil {
		select {
		case <-l.stopCh:
		default:
			close(l.stopCh)
		}
	}
	l.cleanerOn = false
	l.stopCh = nil
	l.ticker = nil
}

// finalize is called by the garbage collector as a safety net to ensure
// that background goroutines are properly cleaned up even if Close() wasn't called.
// This prevents memory leaks in cases where users forget to call Close().
func (l *FastCache) finalize() {
	// Use SwapInt32 to atomically mark as finalized and get previous state
	// This avoids race condition and eliminates need to check cleanerOn
	oldLevel := atomic.SwapInt32(&l.cleanerLevel, -1)
	if oldLevel >= 0 {
		// Previous level was valid (not already closed), so cleanup is needed
		l.stopCleaner()
	}
}

// Stats represents cache performance and status statistics
type Stats struct {
	// AccessCount shows total number of cache accesses since creation
	AccessCount int64
	// IdleDuration shows how long cache has been idle
	IdleDuration time.Duration
	// TotalItems shows approximate total items across all buckets
	TotalItems int
	// CleanerLevel indicates current cleaning intensity (0=normal, 1=reduced, 2=light, 3=stopped)
	CleanerLevel int32
	// IsCleanerRunning indicates if background cleaner is active
	IsCleanerRunning bool
}

// GetStats returns current cache statistics for monitoring and debugging
func (l *FastCache) GetStats() Stats {
	now := ztime.Clock() * 1000
	lastAccess := atomic.LoadInt64(&l.lastAccessMs)
	idleDuration := time.Duration(now-lastAccess) * time.Millisecond

	// Count items across all buckets (approximation to avoid locking)
	totalItems := 0
	for i := range l.insts {
		if l.insts[i][0] != nil {
			totalItems += l.insts[i][0].size
		}
		if l.insts[i][1] != nil {
			totalItems += l.insts[i][1].size
		}
	}

	return Stats{
		CleanerLevel:     atomic.LoadInt32(&l.cleanerLevel),
		AccessCount:      atomic.LoadInt64(&l.accessCount),
		IdleDuration:     idleDuration,
		IsCleanerRunning: l.cleanerOn, // Note: this may have slight race condition but for monitoring it's acceptable
		TotalItems:       totalItems,
	}
}

// GetCleanerLevel returns the current cleaning intensity level
// 0=normal, 1=reduced frequency, 2=light cleaning, 3=stopped
func (l *FastCache) GetCleanerLevel() int32 {
	return atomic.LoadInt32(&l.cleanerLevel)
}

// GetAccessCount returns total number of accesses since cache creation
func (l *FastCache) GetAccessCount() int64 {
	return atomic.LoadInt64(&l.accessCount)
}

// GetIdleDuration returns how long the cache has been idle
func (l *FastCache) GetIdleDuration() time.Duration {
	now := ztime.Clock() * 1000
	lastAccess := atomic.LoadInt64(&l.lastAccessMs)
	return time.Duration(now-lastAccess) * time.Millisecond
}
