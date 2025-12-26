//go:build go1.19
// +build go1.19

package zutil_test

import (
	"runtime/debug"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zutil"
)

func TestNewMemoryLimiter(t *testing.T) {
	tt := zlsgo.NewTest(t)

	ml := zutil.NewMemoryLimiter()
	tt.NotNil(ml)
	defer ml.Stop()

	stats := ml.Stats()
	tt.Equal(uint64(0), stats.CurrentUsage)
	tt.Equal(uint64(0), stats.PeakUsage)
}

func TestNewMemoryLimiterWithOptions(t *testing.T) {
	tt := zlsgo.NewTest(t)

	ml := zutil.NewMemoryLimiter(func(cfg *zutil.MemoryStatsConfig) {
		cfg.Limit = 100 * 1024 * 1024
		cfg.PauseThreshold = 0.9
		cfg.MonitorInterval = 500 * time.Millisecond
		cfg.EnableGC = false
		cfg.SetRuntimeLimit = false
	})
	defer ml.Stop()

	tt.NotNil(ml)
	stats := ml.Stats()
	tt.NotNil(stats.LastGCTime)
}

func TestMemoryLimiter_StartStop(t *testing.T) {
	tt := zlsgo.NewTest(t)

	ml := zutil.NewMemoryLimiter()

	firstStart := ml.Start()
	tt.EqualTrue(firstStart)

	secondStart := ml.Start()
	tt.EqualFalse(secondStart)

	ml.Stop()

	afterStopStart := ml.Start()
	tt.EqualFalse(afterStopStart)
}

func TestMemoryLimiter_StopRestoresLimit(t *testing.T) {
	tt := zlsgo.NewTest(t)

	baseLimit := int64(64 * 1024 * 1024)
	prevLimit := debug.SetMemoryLimit(baseLimit)
	defer debug.SetMemoryLimit(prevLimit)

	limiterLimit := int64(128 * 1024 * 1024)
	ml := zutil.NewMemoryLimiter(func(cfg *zutil.MemoryStatsConfig) {
		cfg.Limit = uint64(limiterLimit)
		cfg.SetRuntimeLimit = true
	})

	appliedPrev := debug.SetMemoryLimit(limiterLimit)
	tt.Equal(limiterLimit, appliedPrev)

	ml.Stop()

	restoredPrev := debug.SetMemoryLimit(baseLimit)
	tt.Equal(baseLimit, restoredPrev)
}

func TestMemoryLimiter_Stats(t *testing.T) {
	tt := zlsgo.NewTest(t)

	ml := zutil.NewMemoryLimiter()
	defer ml.Stop()

	ml.Start()

	time.Sleep(200 * time.Millisecond)

	stats := ml.Stats()
	tt.NotNil(stats.CurrentUsage)
	if stats.CurrentUsage == 0 {
		ml.Refresh()
		stats = ml.Stats()
	}
	tt.NotNil(stats.CurrentUsage)
}

func TestMemoryLimiter_IsPaused(t *testing.T) {
	tt := zlsgo.NewTest(t)

	ml := zutil.NewMemoryLimiter(func(cfg *zutil.MemoryStatsConfig) {
		cfg.Limit = 1024 * 1024 * 1024
	})
	defer ml.Stop()

	tt.EqualFalse(ml.IsPaused())

	ml.Start()
	time.Sleep(100 * time.Millisecond)
	tt.EqualFalse(ml.IsPaused())
}

func TestMemoryLimiter_UpdateLimit(t *testing.T) {
	tt := zlsgo.NewTest(t)

	ml := zutil.NewMemoryLimiter()
	defer ml.Stop()

	ml.UpdateLimit(200 * 1024 * 1024)
	time.Sleep(100 * time.Millisecond)

	ml.UpdateLimit(0)
	tt.NotNil(ml)
}

func TestMemoryLimiter_OnPause(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var onPauseCalled atomic.Int32
	var receivedRatio float64

	ml := zutil.NewMemoryLimiter(func(cfg *zutil.MemoryStatsConfig) {
		cfg.Limit = 1024 * 1024
		cfg.PauseThreshold = 0.1
	})

	ml.OnPause(func(ratio float64) bool {
		onPauseCalled.Add(1)
		receivedRatio = ratio
		return false
	})

	ml.Start()
	time.Sleep(500 * time.Millisecond)

	if onPauseCalled.Load() > 0 {
		tt.EqualTrue(receivedRatio > 0)
	}

	ml.Stop()
}

func TestMemoryLimiter_OnStats(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var onStatsCalled atomic.Int32

	ml := zutil.NewMemoryLimiter()

	ml.OnStats(func(stats zutil.MemoryStats) {
		onStatsCalled.Add(1)
	})

	ml.Start()
	time.Sleep(200 * time.Millisecond)

	if onStatsCalled.Load() == 0 {
		ml.Refresh()
	}
	tt.EqualTrue(onStatsCalled.Load() > 0)

	ml.Stop()
}

func TestMemoryLimiter_Refresh(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var onStatsCalled atomic.Int32

	ml := zutil.NewMemoryLimiter()
	defer ml.Stop()

	ml.OnStats(func(stats zutil.MemoryStats) {
		onStatsCalled.Add(1)
	})

	ml.Refresh()
	tt.EqualTrue(onStatsCalled.Load() > 0)
}

func TestMemoryLimiter_Concurrent(t *testing.T) {
	tt := zlsgo.NewTest(t)

	ml := zutil.NewMemoryLimiter()
	defer ml.Stop()

	var wg sync.WaitGroup

	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			ml.Stats()
			ml.IsPaused()
		}()
	}

	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func(n int) {
			defer wg.Done()
			ml.UpdateLimit(uint64(n * 1024 * 1024))
		}(i)
	}

	wg.Wait()

	stats := ml.Stats()
	tt.NotNil(stats)
}

func TestMemoryLimiter_DefaultConfig(t *testing.T) {
	tt := zlsgo.NewTest(t)

	ml := zutil.NewMemoryLimiter(func(cfg *zutil.MemoryStatsConfig) {
		cfg.Limit = 0
		cfg.PauseThreshold = -1
		cfg.MonitorInterval = 1 * time.Millisecond
	})
	defer ml.Stop()

	tt.NotNil(ml)

	ml.Refresh()
	stats := ml.Stats()
	tt.NotNil(stats)
}

func TestMemoryLimiter_OnPauseContinue(t *testing.T) {
	tt := zlsgo.NewTest(t)

	ml := zutil.NewMemoryLimiter(func(cfg *zutil.MemoryStatsConfig) {
		cfg.Limit = 1024
		cfg.PauseThreshold = 0.01
	})

	continueCalled := atomic.Int32{}

	ml.OnPause(func(ratio float64) bool {
		continueCalled.Add(1)
		return true
	})

	ml.Start()
	time.Sleep(300 * time.Millisecond)

	if continueCalled.Load() > 0 {
		tt.EqualFalse(ml.IsPaused())
	}

	ml.Stop()
}

func TestMemoryLimiter_PauseRecovery(t *testing.T) {
	_ = zlsgo.NewTest(t)

	paused := atomic.Int32{}
	resumed := atomic.Int32{}

	ml := zutil.NewMemoryLimiter(func(cfg *zutil.MemoryStatsConfig) {
		cfg.Limit = 100 * 1024 * 1024
		cfg.PauseThreshold = 0.01
	})

	ml.OnPause(func(ratio float64) bool {
		paused.Add(1)
		return false
	})

	ml.Start()
	time.Sleep(200 * time.Millisecond)

	ml.UpdateLimit(10 * 1024 * 1024 * 1024)
	time.Sleep(300 * time.Millisecond)

	if paused.Load() > 0 {
		if !ml.IsPaused() {
			resumed.Add(1)
		}
	}

	ml.Stop()
}
