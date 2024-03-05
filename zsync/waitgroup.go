package zsync

import (
	"sync"

	"github.com/sohaha/zlsgo/zerror"
)

type WaitGroup struct {
	err error
	wg  sync.WaitGroup
	mu  sync.RWMutex
}

func (h *WaitGroup) Add(delta int) {
	h.wg.Add(delta)
}

func (h *WaitGroup) Done() {
	h.wg.Done()
}

func (h *WaitGroup) Go(f func()) {
	h.Add(1)
	go func() {
		defer h.Done()
		f()
	}()
}

func (h *WaitGroup) GoTry(f func()) {
	h.Add(1)
	go func() {
		defer h.Done()
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

func (h *WaitGroup) Wait() error {
	h.wg.Wait()
	return h.err
}
