package zsync

import (
	"sync"

	"github.com/sohaha/zlsgo/zerror"
)

// WaitGroup extends sync.WaitGroup with additional features including:
// - Concurrency limiting (max number of concurrent goroutines)
// - Error propagation from goroutines to the Wait call
// - Panic recovery in goroutines
type WaitGroup struct {
	err error           // Stores the first error encountered in any goroutine
	ch  chan struct{}   // Channel for limiting concurrency
	wg  sync.WaitGroup  // Underlying wait group for synchronization
	mu  sync.RWMutex    // Mutex for protecting err field
}

// NewWaitGroup creates a new WaitGroup instance.
// If max is provided, it limits the number of concurrent goroutines to that value.
// Without max, there is no limit on concurrency.
func NewWaitGroup(max ...uint) *WaitGroup {
	wg := &WaitGroup{}

	if len(max) > 0 {
		wg.ch = make(chan struct{}, max[0])
	}

	return wg
}

// Add adds delta to the WaitGroup counter.
// If the counter becomes zero, all goroutines blocked on Wait are released.
// If the counter goes negative, Add panics.
func (h *WaitGroup) Add(delta int) {
	h.wg.Add(delta)
}

// Done decrements the WaitGroup counter by one.
// It is equivalent to calling Add(-1).
func (h *WaitGroup) Done() {
	h.wg.Done()
}

// Go runs a function in a new goroutine and tracks it with the WaitGroup.
// If a concurrency limit was set when creating the WaitGroup, this method
// will block until the number of concurrent goroutines is below the limit.
func (h *WaitGroup) Go(f func()) {
	if h.ch != nil {
		h.ch <- struct{}{}
	}
	h.Add(1)
	go func() {
		defer func() {
			if h.ch != nil {
				<-h.ch
			}
			h.Done()
		}()
		f()
	}()
}

func (h *WaitGroup) GoTry(f func()) {
	if h.ch != nil {
		h.ch <- struct{}{}
	}
	h.Add(1)
	go func() {
		defer func() {
			if h.ch != nil {
				<-h.ch
			}
			h.Done()
		}()
		err := zerror.TryCatch(func() error {
			f()
			return nil
		})
		if err != nil {
			h.mu.Lock()
			if h.err == nil {
				h.err = err
			}
			h.mu.Unlock()
		}
	}()
}

// Wait blocks until the WaitGroup counter is zero.
// It returns the first error that was encountered in any of the goroutines
// launched with GoTry, or nil if no errors occurred.
func (h *WaitGroup) Wait() error {
	h.wg.Wait()
	return h.err
}
