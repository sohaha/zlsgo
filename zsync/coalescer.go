package zsync

import (
	"sync/atomic"
)

// NewCoalescer returns a function that coalesces multiple calls into a single
// or a few executions: while one execution is running, further calls schedule
// at least one more run. Thread-safe.
func NewCoalescer(fn func()) func() {
	var running uint32
	var pending uint32

	return func() {
		if !atomic.CompareAndSwapUint32(&running, 0, 1) {
			atomic.AddUint32(&pending, 1)
			return
		}

		defer atomic.StoreUint32(&running, 0)

		for {
			if fn != nil {
				fn()
			}

			if atomic.SwapUint32(&pending, 0) > 0 {
				continue
			}

			break
		}
	}
}
