package zpool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sohaha/zlsgo/zutil"
)

type (
	// Task Define function callbacks
	Task     func()
	WorkPool struct {
		workers   sync.Pool
		closed    bool
		mu        sync.RWMutex
		queue     chan *worker
		minIdle   uint
		usedNum   uint
		maxIdle   uint
		panicFunc PanicFunc
	}
	worker struct {
		jobQueue  chan Task
		stop      chan struct{}
		Parameter chan []interface{}
	}
	PanicFunc func(err error)
)

var (
	ErrPoolClosed  = errors.New("pool has been closed")
	ErrWaitTimeout = errors.New("pool wait timeout")
)

func New(min int, max ...int) *WorkPool {
	minIdle := uint(min)
	if minIdle <= 0 {
		minIdle = 1
	}
	maxIdle := minIdle
	if len(max) > 0 && max[0] > 0 {
		max := uint(max[0])
		if max > maxIdle {
			maxIdle = max
		}
	}

	w := &WorkPool{
		minIdle: minIdle,
		maxIdle: maxIdle,
		queue:   make(chan *worker, maxIdle),
		workers: sync.Pool{New: func() interface{} {
			return &worker{
				jobQueue:  make(chan Task),
				Parameter: make(chan []interface{}),
				stop:      make(chan struct{}),
			}
		}},
	}
	return w
}

// Do Add to the workpool and implement
func (wp *WorkPool) Do(fn Task) error {
	return wp.do(context.Background(), fn, nil)
}

func (wp *WorkPool) DoWithTimeout(fn Task, t time.Duration) error {
	ctx, canle := context.WithTimeout(context.Background(), t)
	defer canle()
	return wp.do(ctx, fn, nil)
}

// Do Add to the workpool and implement
func (wp *WorkPool) PanicFunc(handler PanicFunc) {
	wp.panicFunc = handler
}

func (wp *WorkPool) do(cxt context.Context, fn Task, param []interface{}) error {
	if wp.IsClosed() {
		return ErrPoolClosed
	}
	wp.mu.Lock()
	run := func(w *worker) {
		if fn != nil {
			w.jobQueue <- fn
		}
	}
	add := func() *worker {
		wp.usedNum++
		wp.mu.Unlock()
		w := wp.workers.Get().(*worker)
		w.createGoroutines(wp.queue, wp.panicFunc)
		return w
	}
	select {
	case w := <-wp.queue:
		wp.mu.Unlock()
		if w != nil {
			run(w)
		} else {
			return ErrPoolClosed
		}
	default:
		switch {
		case wp.usedNum >= wp.minIdle:
			wp.mu.Unlock()
			// todo 超时处理
			select {
			case <-cxt.Done():
				wp.mu.Lock()
				if wp.usedNum >= wp.maxIdle {
					wp.mu.Unlock()
					return ErrWaitTimeout
				}
				w := add()
				run(w)
				return nil
				// 尝试扩大容量？
			case w := <-wp.queue:
				if w != nil {
					run(w)
				} else {
					return ErrPoolClosed
				}

			}
		case wp.usedNum < wp.minIdle:
			w := add()
			run(w)
		default:
			wp.mu.Unlock()
		}
	}
	return nil
}

// IsClosed Has it been closed
func (wp *WorkPool) IsClosed() bool {
	wp.mu.RLock()
	b := wp.closed
	wp.mu.RUnlock()
	return b
}

// Close close the pool
func (wp *WorkPool) Close() {
	if wp.IsClosed() {
		return
	}
	wp.mu.Lock()
	wp.closed = true
	for 0 < wp.usedNum {
		wp.usedNum--
		worker := <-wp.queue
		worker.close()
	}
	wp.mu.Unlock()
}

// Pause pause
func (wp *WorkPool) Pause() {
	wp.AdjustSize(0)
}

// Continue continue
func (wp *WorkPool) Continue(workerNum ...int) {
	num := int(wp.maxIdle)
	if len(workerNum) > 0 {
		num = workerNum[0]
	}
	wp.AdjustSize(num)
}

// Cap get the number of coroutines
func (wp *WorkPool) Cap() uint {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.usedNum
}

// AdjustSize adjust the pool size
func (wp *WorkPool) AdjustSize(workSize int) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	if wp.closed {
		return
	}

	oldSize := wp.minIdle
	newSize := uint(workSize)
	if newSize > wp.maxIdle {
		newSize = wp.maxIdle
	}
	wp.minIdle = newSize

	if workSize > 0 && oldSize < wp.minIdle {
		for wp.usedNum < wp.minIdle {
			wp.usedNum++
			w := wp.workers.Get().(*worker)
			w.createGoroutines(wp.queue, wp.panicFunc)
			wp.queue <- w
		}
	}
	for wp.minIdle < wp.usedNum {
		wp.usedNum--
		worker := <-wp.queue
		worker.stop <- struct{}{}
		wp.workers.Put(worker)
	}
}

func (wp *WorkPool) PreInit() error {
	if wp.IsClosed() {
		return ErrPoolClosed
	}
	wp.mu.Lock()
	for wp.usedNum < wp.minIdle {
		wp.usedNum++
		w := wp.workers.Get().(*worker)
		w.createGoroutines(wp.queue, wp.panicFunc)
		wp.queue <- w
	}
	wp.mu.Unlock()
	return nil
}

func (w *worker) createGoroutines(q chan<- *worker, handler PanicFunc) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				w.createGoroutines(q, handler)
				q <- w
			}
		}()
		for {
			select {
			case job := <-w.jobQueue:
				zutil.Try(job, func(err interface{}) {
					if handler != nil {
						errMsg, ok := err.(error)
						if !ok {
							errMsg = errors.New(fmt.Sprint(err))
						}
						handler(errMsg)
					}
				}, func() {
					q <- w
				})
			// case parameter := <-w.Parameter:
			// 	q <- w
			case <-w.stop:
				return
			}
		}
	}()
}

func (w *worker) close() {
	w.stop <- struct{}{}
	close(w.stop)
	close(w.jobQueue)
}
