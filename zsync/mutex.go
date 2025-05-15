package zsync

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sohaha/zlsgo/zstring"
)

type (
	// RBMutex is a reader biased reader/writer mutual exclusion lock.
	// It is optimized for read-heavy workloads by minimizing contention between readers.
	// Writers will still block all readers, similar to sync.RWMutex.
	RBMutex struct {
		inhibitUntil time.Time      // Time until which read bias is inhibited
		rslots       []rslot        // Array of reader slots to distribute contention
		rw           sync.RWMutex   // Underlying RWMutex for write locking
		rmask        uint32         // Mask for selecting reader slots
		rbias        int32          // Flag indicating if read bias is active
	}

	// RToken represents a read lock token that must be passed to RUnlock.
	// It contains the slot information needed to release the correct read lock.
	RToken struct {
		slot uint32                  // The reader slot index
		pad  [cacheLineSize - 4]byte // Padding to prevent false sharing
	}

	// rslot is an internal structure representing a single reader slot.
	// Each slot has its own mutex to minimize contention between readers.
	rslot struct {
		mu  int32                   // Mutex for this slot (0=unlocked, 1=locked)
		pad [cacheLineSize - 4]byte // Padding to prevent false sharing
	}
)

const nslowdown = 7

var rtokenPool sync.Pool

// NewRBMutex creates a new RBMutex instance.
// It initializes the mutex with a number of reader slots based on the system's parallelism level,
// which helps distribute reader lock contention across multiple memory locations.
func NewRBMutex() *RBMutex {
	nslots := nextPowOf2(parallelism())
	mu := RBMutex{
		rslots: make([]rslot, nslots),
		rmask:  nslots - 1,
		rbias:  1,
	}
	return &mu
}

// RLock acquires a read lock and returns a token that must be used to release the lock.
// Unlike sync.RWMutex, this method returns a token that must be passed to RUnlock.
// Multiple readers can hold the lock simultaneously while no writers are holding it.
func (mu *RBMutex) RLock() *RToken {
	if atomic.LoadInt32(&mu.rbias) == 1 {
		t, ok := rtokenPool.Get().(*RToken)
		if !ok {
			t = &RToken{}
			t.slot = zstring.RandUint32()
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

// RUnlock releases a read lock using the token obtained from RLock.
// The token must match the one returned by the corresponding RLock call.
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

// Lock acquires a write lock, blocking all readers and other writers.
// This behaves similarly to sync.RWMutex.Lock() but also temporarily
// disables the read bias optimization after the lock is released.
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

// Unlock releases a write lock, allowing readers and other writers to proceed.
// After a write lock is released, read bias is temporarily inhibited to
// give waiting writers a fair chance to acquire the lock.
func (mu *RBMutex) Unlock() {
	mu.rw.Unlock()
}
