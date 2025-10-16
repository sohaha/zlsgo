//go:build go1.18
// +build go1.18

package zsync

import (
	"runtime"
	"sync/atomic"
)

// SeqLockT is a typed sequence lock that avoids interface conversions on the hot path.
// It provides the same semantics as the untyped SeqLock but returns/accepts T directly.
type SeqLockT[T any] struct {
	seq  uint64
	_pad [56]byte

	data  T
	_pad2 [56]byte
}

// NewSeqLockT creates a typed sequence lock.
func NewSeqLock[T any]() *SeqLockT[T] { return &SeqLockT[T]{} }

// Write publishes a new value with seqlock semantics.
func (s *SeqLockT[T]) Write(v T) {
	atomic.AddUint64(&s.seq, 1)
	s.data = v
	atomic.AddUint64(&s.seq, 1)
}

// Read returns a consistent snapshot if the sequence was stable.
// It may spin briefly under write contention.
func (s *SeqLockT[T]) Read() (T, bool) {
	for spin := 0; ; spin++ {
		seq1 := atomic.LoadUint64(&s.seq)
		if seq1&1 != 0 {
			if spin&15 == 15 {
				runtime.Gosched()
			}
			continue
		}
		v := s.data
		if seq1 == atomic.LoadUint64(&s.seq) {
			return v, true
		}
	}
}
