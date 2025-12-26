//go:build go1.19
// +build go1.19

package zutil

import (
	"context"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/sohaha/zlsgo/zfile"
)

// MemoryStats holds memory usage statistics.
type MemoryStats struct {
	CurrentUsage uint64    // Current heap allocation in bytes
	PeakUsage    uint64    // Peak heap allocation in bytes
	LastGCTime   time.Time // Time of the last garbage collection
	NumGC        uint32    // Number of garbage collections
	HeapInuse    uint64    // Bytes in in-use spans
	HeapSys      uint64    // Bytes obtained from system
	PauseTotalNs uint64    // Cumulative GC pause time in nanoseconds
}

// MemoryStatsConfig configures the memory limiter behavior.
type MemoryStatsConfig struct {
	Limit           uint64        // Memory hard limit in bytes
	PauseThreshold  float64       // Pause threshold ratio [0, 1], default 0.85
	MonitorInterval time.Duration // Monitoring interval, default 1s
	EnableGC        bool          // Trigger GC when exceeding limit, default true
	SetRuntimeLimit bool          // Call debug.SetMemoryLimit, default true
}

// MemoryLimiter monitors and controls memory usage.
// It tracks heap allocation, triggers GC when needed, and can pause
// operations when memory exceeds configured thresholds.
type MemoryLimiter struct {
	mu        sync.Mutex
	config    MemoryStatsConfig
	stats     MemoryStats
	paused    bool
	onPause   func(ratio float64) bool
	onStats   func(stats MemoryStats)
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	started   bool
	lastNumGC uint32
	prevLimit int64
	setLimit  bool
}

// NewMemoryLimiter creates a new memory limiter with optional configuration.
func NewMemoryLimiter(opt ...func(cfg *MemoryStatsConfig)) *MemoryLimiter {
	limit := uint64(50 * zfile.MB)
	pauseThreshold := 0.85
	monitorInterval := 1 * time.Second
	cfg := Optional(MemoryStatsConfig{
		Limit:           limit,
		PauseThreshold:  pauseThreshold,
		MonitorInterval: monitorInterval,
		EnableGC:        true,
		SetRuntimeLimit: true,
	}, opt...)

	if cfg.Limit == 0 {
		cfg.Limit = limit
	}
	if cfg.PauseThreshold <= 0 || cfg.PauseThreshold > 1 {
		cfg.PauseThreshold = pauseThreshold
	}
	if cfg.MonitorInterval < 10*time.Millisecond {
		cfg.MonitorInterval = monitorInterval
	}

	ctx, cancel := context.WithCancel(context.Background())
	ml := &MemoryLimiter{
		config:  cfg,
		ctx:     ctx,
		cancel:  cancel,
		paused:  false,
		started: false,
	}

	if cfg.SetRuntimeLimit {
		ml.prevLimit = debug.SetMemoryLimit(int64(cfg.Limit))
		ml.setLimit = true
	}

	return ml
}

// Start begins the monitoring goroutine.
// Returns false if already started or if the limiter was stopped.
func (ml *MemoryLimiter) Start() bool {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	if ml.started {
		return false
	}
	if ml.ctx.Err() != nil {
		return false
	}

	ml.started = true
	ml.wg.Add(1)
	go ml.monitor()
	return true
}

// Stop stops the monitoring goroutine and waits for it to exit.
func (ml *MemoryLimiter) Stop() {
	ml.cancel()
	ml.wg.Wait()
	if ml.setLimit {
		debug.SetMemoryLimit(ml.prevLimit)
	}
}

// Refresh manually updates statistics and performs a memory check.
func (ml *MemoryLimiter) Refresh() {
	ml.updateStats()
	ml.checkMemoryUsage()
}

// Stats returns a copy of the current memory statistics.
func (ml *MemoryLimiter) Stats() MemoryStats {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	return ml.stats
}

// IsPaused returns whether the limiter is currently in paused state.
func (ml *MemoryLimiter) IsPaused() bool {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	return ml.paused
}

// UpdateLimit dynamically updates the memory limit.
// A limit of 0 is ignored.
func (ml *MemoryLimiter) UpdateLimit(limit uint64) {
	if limit == 0 {
		return
	}
	ml.mu.Lock()
	ml.config.Limit = limit
	setRuntimeLimit := ml.config.SetRuntimeLimit
	ml.mu.Unlock()

	if setRuntimeLimit {
		debug.SetMemoryLimit(int64(limit))
	}
}

// OnPause sets the callback invoked when memory exceeds PauseThreshold.
// The callback receives the current usage ratio and returns true to continue
// processing, or false to pause.
func (ml *MemoryLimiter) OnPause(fn func(ratio float64) bool) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.onPause = fn
}

// OnStats sets the callback invoked with updated memory statistics.
func (ml *MemoryLimiter) OnStats(fn func(stats MemoryStats)) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.onStats = fn
}

func (ml *MemoryLimiter) monitor() {
	defer ml.wg.Done()

	ticker := time.NewTicker(ml.config.MonitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ml.ctx.Done():
			return
		case <-ticker.C:
			ml.Refresh()
		}
	}
}

func (ml *MemoryLimiter) updateStats() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	ml.mu.Lock()
	defer ml.mu.Unlock()

	ml.stats.CurrentUsage = memStats.Alloc
	if memStats.Alloc > ml.stats.PeakUsage {
		ml.stats.PeakUsage = memStats.Alloc
	}
	ml.stats.NumGC = memStats.NumGC
	ml.stats.HeapInuse = memStats.HeapInuse
	ml.stats.HeapSys = memStats.HeapSys
	ml.stats.PauseTotalNs = memStats.PauseTotalNs

	if ml.lastNumGC != memStats.NumGC {
		ml.stats.LastGCTime = time.Unix(0, int64(memStats.LastGC))
		ml.lastNumGC = memStats.NumGC
	}
}

func (ml *MemoryLimiter) checkMemoryUsage() {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	current := ml.stats.CurrentUsage
	ratio := float64(current) / float64(ml.config.Limit)

	if ratio > ml.config.PauseThreshold {
		if ml.config.EnableGC {
			ml.runGCUnsafe()
			current = ml.stats.CurrentUsage
			ratio = float64(current) / float64(ml.config.Limit)
		}

		if ml.onPause != nil {
			shouldPause := !ml.onPause(ratio)
			ml.paused = shouldPause
		} else {
			ml.paused = true
		}
	} else {
		if ml.paused && ratio < ml.config.PauseThreshold*0.9 {
			ml.paused = false
		}
	}

	if ml.onStats != nil {
		statsCopy := ml.stats
		ml.mu.Unlock()
		ml.onStats(statsCopy)
		ml.mu.Lock()
	}
}

func (ml *MemoryLimiter) runGCUnsafe() {
	runtime.GC()
	ml.updateStatsUnsafe()
}

func (ml *MemoryLimiter) updateStatsUnsafe() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	ml.stats.CurrentUsage = memStats.Alloc
	if memStats.Alloc > ml.stats.PeakUsage {
		ml.stats.PeakUsage = memStats.Alloc
	}
	ml.stats.NumGC = memStats.NumGC
	ml.stats.HeapInuse = memStats.HeapInuse
	ml.stats.HeapSys = memStats.HeapSys
	ml.stats.PauseTotalNs = memStats.PauseTotalNs

	if ml.lastNumGC != memStats.NumGC {
		ml.stats.LastGCTime = time.Unix(0, int64(memStats.LastGC))
		ml.lastNumGC = memStats.NumGC
	}
}
