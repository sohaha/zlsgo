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
	rslots         []optRSlot
	state          uint64
	rw             sync.RWMutex
	rmask          uint32
	writerMomentum uint32
	_              [64]byte
	_              [56]byte
	_              [4]byte
}

type optRSlot struct {
	counter uint64
	_       [56]byte
}

type RBToken struct {
	p *uint64
}

const (
	rbiasShift           = 63
	rbiasMask            = uint64(1) << rbiasShift
	writerShift          = 32
	writerMask           = uint64(0x7FFFFFFF) << writerShift
	defaultBiasLimit     = 4
	writerMomentumMedium = 6
	writerMomentumHigh   = 12
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

//go:nosplit
func unlikely(b bool) bool {
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
		return RBToken{p: cptr}
	}

	if unlikely(state&rbiasMask != 0) {
		momentum := atomic.LoadUint32(&mu.writerMomentum)
		limit := biasLimit(momentum)

		if (state & writerMask) < uint64(limit)<<writerShift {
			if len(mu.rslots) == 0 {
				mu.rw.RLock()
				return RBToken{p: nil}
			}
			slot := getProcID() & mu.rmask
			cptr := &mu.rslots[slot].counter
			atomic.AddUint64(cptr, 1)

			s2 := atomic.LoadUint64(&mu.state)
			if likely(s2&rbiasMask != 0 && (s2&writerMask) < uint64(limit)<<writerShift) {
				return RBToken{p: cptr}
			}
			atomic.AddUint64(cptr, ^uint64(0))
		}
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
	for {
		state := atomic.LoadUint64(&mu.state)
		newState := (state &^ rbiasMask) + (1 << writerShift)
		if atomic.CompareAndSwapUint64(&mu.state, state, newState) {
			break
		}
	}

	atomic.AddUint32(&mu.writerMomentum, 1)

	mu.rw.Lock()

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
		// Adaptive backoff to reduce potential writer starvation
		switch {
		case tries < 64:
			if tries&7 == 0 {
				runtime.Gosched()
			}
		case tries < 256:
			runtime.Gosched()
		default:
			// As contention persists, yield more aggressively
			time.Sleep(time.Microsecond)
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

	for {
		momentum := atomic.LoadUint32(&mu.writerMomentum)
		if momentum == 0 {
			break
		}
		dec := (momentum + 1) >> 1
		if dec == 0 {
			dec = 1
		}
		if atomic.CompareAndSwapUint32(&mu.writerMomentum, momentum, momentum-dec) {
			break
		}
	}
}

//go:nosplit
//go:inline
func biasLimit(momentum uint32) uint32 {
	if momentum > writerMomentumHigh {
		return 1
	}
	if momentum > writerMomentumMedium {
		return 2
	}
	return defaultBiasLimit
}
