//go:build amd64 || arm64 || ppc64 || ppc64le || mips64 || mips64le || s390x || riscv64 || loong64

package zsync

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	_ "unsafe" // required for //go:linkname
)

type RBMutex struct {
	// state is read by every biased reader and written only by writers, so it
	// gets its own cache line: sharing a line with rw would make each writer
	// invalidate the word the read path depends on.
	state uint64
	_     [56]byte

	rw     sync.RWMutex
	rslots []optRSlot
	rmask  uint32
}

type optRSlot struct {
	counter uint64
	_       [56]byte
}

type RBToken struct {
	p *uint64
}

const (
	rbiasShift  = 63
	rbiasMask   = uint64(1) << rbiasShift
	writerShift = 32
	writerMask  = uint64(0x7FFFFFFF) << writerShift
)

// Use runtime's proc pin/unpin to obtain a stable per-P identifier.
// These are internal runtime functions accessed via linkname.
// See: src/runtime/proc.go (procPin/procUnpin)
//
//go:linkname runtime_procPin runtime.procPin
func runtime_procPin() int

//go:linkname runtime_procUnpin runtime.procUnpin
func runtime_procUnpin()

//go:nosplit
func getProcID() uint32 {
	// Pin to current P to retrieve its id, then unpin immediately.
	// We only need the id for slot selection; we do not keep the P pinned
	// across the critical section to avoid excessive overhead.
	id := runtime_procPin()
	runtime_procUnpin()
	return uint32(id)
}

//go:nosplit
func likely(b bool) bool {
	return b
}

// NewRBMutex Extreme optimized version of read bias lock (read more and write less scene)
func NewRBMutex() *RBMutex {
	p := parallelism()
	if p == 0 {
		p = 1
	}
	nslots := nextPowOf2(p * 4)
	if nslots == 0 {
		nslots = 1
	}
	mu := &RBMutex{
		state:  rbiasMask,
		rslots: make([]optRSlot, nslots),
		rmask:  nslots - 1,
	}
	return mu
}

//go:nosplit
func (mu *RBMutex) RLock() RBToken {
	state := atomic.LoadUint64(&mu.state)
	if likely(state&rbiasMask != 0 && (state&writerMask) == 0) {
		// If slots are not initialized, fall back to RWMutex to keep semantics safe.
		if len(mu.rslots) == 0 {
			mu.rw.RLock()
			return RBToken{p: nil}
		}
		slot := getProcID() & mu.rmask
		cptr := &mu.rslots[slot].counter
		atomic.AddUint64(cptr, 1)

		// Re-check after registering the slot. If a writer already cleared
		// the bias bit, back off to rw.RLock. If the writer arrives later,
		// its slot scan observes this registration and waits for RUnlock.
		state = atomic.LoadUint64(&mu.state)
		if likely(state&rbiasMask != 0 && (state&writerMask) == 0) {
			return RBToken{p: cptr}
		}
		atomic.AddUint64(cptr, ^uint64(0))
	}

	mu.rw.RLock()
	return RBToken{p: nil}
}

//go:nosplit
func (mu *RBMutex) RUnlock(token RBToken) {
	if token.p == nil {
		mu.rw.RUnlock()
		return
	}

	atomic.AddUint64(token.p, ^uint64(0))
}

func (mu *RBMutex) Lock() {
	// If another writer has already cleared the bias bit, no biased reader can
	// have registered since: the fast read path requires that bit to be set.
	// That writer also drained the slots before entering its own critical
	// section, and this writer only reaches its section after that one, so the
	// slot scan can be skipped entirely.
	drain := true
	for {
		state := atomic.LoadUint64(&mu.state)
		newState := (state &^ rbiasMask) + (1 << writerShift)
		if atomic.CompareAndSwapUint64(&mu.state, state, newState) {
			drain = state&rbiasMask != 0
			break
		}
	}

	mu.rw.Lock()
	if !drain {
		return
	}

	tries := 0
	for {
		allZero := true
		for i := range mu.rslots {
			if atomic.LoadUint64(&mu.rslots[i].counter) > 0 {
				allZero = false
				break
			}
		}
		if allZero {
			break
		}
		tries++
		// Bias readers hold no rw lock, so the writer must poll for them to
		// drain. Yield instead of sleeping: time.Sleep rounds up to the timer
		// granularity (tens of microseconds) and stalls every reader behind
		// rw.Lock for that long. Only fall back to sleeping when a reader has
		// held its section for an unusually long time, so a stuck reader does
		// not burn a core.
		switch {
		case tries < 16:
		case tries < 1024:
			runtime.Gosched()
		default:
			time.Sleep(50 * time.Microsecond)
		}
	}
}

func (mu *RBMutex) Unlock() {
	mu.rw.Unlock()

	for {
		state := atomic.LoadUint64(&mu.state)
		writerCount := (state & writerMask) >> writerShift
		newState := state - (1 << writerShift)

		if writerCount == 1 {
			// Only enable read-bias if slots are initialized.
			if len(mu.rslots) > 0 {
				newState |= rbiasMask
			}
		}

		if atomic.CompareAndSwapUint64(&mu.state, state, newState) {
			break
		}
	}
}
