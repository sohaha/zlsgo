package zsync

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type (
	// RBMutex is a reader biased reader/writer mutual exclusion lock
	RBMutex struct {
		rslots       []rslot
		rmask        uint32
		rbias        int32
		inhibitUntil time.Time
		rw           sync.RWMutex
	}
	RToken struct {
		slot uint32
		pad  [cacheLineSize - 4]byte
	}
	rslot struct {
		mu  int32
		pad [cacheLineSize - 4]byte
	}
)

const nslowdown = 7

var rtokenPool sync.Pool

// NewRBMutex creates a new RBMutex instance.
func NewRBMutex() *RBMutex {
	nslots := nextPowOf2(parallelism())
	mu := RBMutex{
		rslots: make([]rslot, nslots),
		rmask:  nslots - 1,
		rbias:  1,
	}
	return &mu
}

func (mu *RBMutex) RLock() *RToken {
	if atomic.LoadInt32(&mu.rbias) == 1 {
		t, ok := rtokenPool.Get().(*RToken)
		if !ok {
			t = new(RToken)
			t.slot = fastrand()
		}
		for i := 0; i < len(mu.rslots); i++ {
			slot := t.slot + uint32(i)
			rslot := &mu.rslots[slot&mu.rmask]
			rslotmu := atomic.LoadInt32(&rslot.mu)
			if atomic.CompareAndSwapInt32(&rslot.mu, rslotmu, rslotmu+1) {
				if atomic.LoadInt32(&mu.rbias) == 1 {
					t.slot = slot
					return t
				}
				atomic.AddInt32(&rslot.mu, -1)
				rtokenPool.Put(t)
				break
			}
		}
	}

	mu.rw.RLock()
	if atomic.LoadInt32(&mu.rbias) == 0 && time.Now().After(mu.inhibitUntil) {
		atomic.StoreInt32(&mu.rbias, 1)
	}
	return nil
}

func (mu *RBMutex) RUnlock(t *RToken) {
	if t == nil {
		mu.rw.RUnlock()
		return
	}
	if atomic.AddInt32(&mu.rslots[t.slot&mu.rmask].mu, -1) < 0 {
		panic("invalid reader state detected")
	}
	rtokenPool.Put(t)
}

func (mu *RBMutex) Lock() {
	mu.rw.Lock()
	if atomic.LoadInt32(&mu.rbias) == 1 {
		atomic.StoreInt32(&mu.rbias, 0)
		start := time.Now()
		for i := 0; i < len(mu.rslots); i++ {
			for atomic.LoadInt32(&mu.rslots[i].mu) > 0 {
				runtime.Gosched()
			}
		}
		mu.inhibitUntil = time.Now().Add(time.Since(start) * nslowdown)
	}
}

func (mu *RBMutex) Unlock() {
	mu.rw.Unlock()
}
