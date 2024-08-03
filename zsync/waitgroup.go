package zsync

import (
	"sync"

	"github.com/sohaha/zlsgo/zerror"
)

type WaitGroup struct {
	err error
	ch  chan struct{}
	wg  sync.WaitGroup
	mu  sync.RWMutex
}

func NewWaitGroup(max ...uint) *WaitGroup {
	wg := &WaitGroup{}

	if len(max) > 0 {
		wg.ch = make(chan struct{}, max[0])
	}

	return wg
}

func (h *WaitGroup) Add(delta int) {
	h.wg.Add(delta)
}

func (h *WaitGroup) Done() {
	h.wg.Done()
}

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

func (h *WaitGroup) Wait() error {
	h.wg.Wait()
	return h.err
}
