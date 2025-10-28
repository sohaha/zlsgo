//go:build go1.18
// +build go1.18

package zsync

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

// SeqLockT is a typed sequence lock that avoids interface conversions on the hot path.
// It provides the same semantics as the untyped SeqLock but returns/accepts T directly.
type SeqLock[T any] struct {
	ptr   unsafe.Pointer
	seq   uint64
	_pad  [56]byte
	_pad2 [56]byte
}

// NewSeqLock creates a typed sequence lock.
func NewSeqLock[T any]() *SeqLock[T] { return &SeqLock[T]{} }

// Write publishes a new value with seqlock semantics.
func (s *SeqLock[T]) Write(v T) {
	atomic.AddUint64(&s.seq, 1)
	nv := new(T)
	*nv = v
	atomic.StorePointer(&s.ptr, unsafe.Pointer(nv))
	atomic.AddUint64(&s.seq, 1)
}

// Read returns a consistent snapshot if the sequence was stable.
// It may spin briefly under write contention.
func (s *SeqLock[T]) Read() (T, bool) {
	var zero T
	for spin := 0; ; spin++ {
		seq1 := atomic.LoadUint64(&s.seq)
		if seq1&1 != 0 { // writer active
			if spin&15 == 15 {
				runtime.Gosched()
			}
			continue
		}
		p := atomic.LoadPointer(&s.ptr)
		if p == nil {
			if seq1 == atomic.LoadUint64(&s.seq) {
				return zero, false
			}
			continue
		}
		v := *(*T)(p)
		if seq1 == atomic.LoadUint64(&s.seq) {
			return v, true
		}
		if spin&15 == 15 {
			runtime.Gosched()
		}
	}
}
