package zpool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zutil"
)

type (
	// Task Define function callbacks
	Task     interface{}
	taskfn   func() error
	WorkPool struct {
		workers     sync.Pool
		closed      bool
		mu          sync.RWMutex
		queue       chan *worker
		minIdle     uint
		maxIdle     uint
		usedNum     *zutil.Int64
		injector    zdi.Injector
		panicFunc   PanicFunc
		releaseTime time.Duration
		New         func()
	}
	worker struct {
		jobQueue  chan taskfn
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
		minIdle:     minIdle,
		maxIdle:     maxIdle,
		injector:    zdi.New(),
		queue:       make(chan *worker, maxIdle),
		usedNum:     zutil.NewInt64(0),
		releaseTime: time.Second * 60,
		workers: sync.Pool{New: func() interface{} {
			return &worker{
				jobQueue:  make(chan taskfn),
				Parameter: make(chan []interface{}),
				stop:      make(chan struct{}),
			}
		}},
	}
	return w
}

// Do Add to the workpool and implement
func (wp *WorkPool) Do(fn Task) error {
	return wp.do(context.Background(), wp.handlerFunc(fn), nil)
}

func (wp *WorkPool) DoWithTimeout(fn Task, t time.Duration) error {
	ctx, canle := context.WithTimeout(context.Background(), t)
	defer canle()
	return wp.do(ctx, wp.handlerFunc(fn), nil)
}

// PanicFunc Do Add to the workpool and implement
func (wp *WorkPool) PanicFunc(handler PanicFunc) {
	wp.panicFunc = handler
}

func (wp *WorkPool) do(cxt context.Context, fn taskfn, param []interface{}) error {
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
		wp.usedNum.Add(1)
		wp.mu.Unlock()
		w := wp.workers.Get().(*worker)
		go w.createGoroutines(wp, wp.queue, wp.panicFunc)
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
		case uint(wp.usedNum.Load()) >= wp.minIdle:
			wp.mu.Unlock()
			// todo 超时处理
			select {
			case <-cxt.Done():
				wp.mu.Lock()
				if uint(wp.usedNum.Load()) >= wp.maxIdle {
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
		case uint(wp.usedNum.Load()) < wp.minIdle:
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

// Close  the pool
func (wp *WorkPool) Close() {
	if wp.IsClosed() {
		return
	}
	wp.mu.Lock()
	wp.closed = true
	for 0 < uint(wp.usedNum.Load()) {
		wp.usedNum.Sub(1)
		worker := <-wp.queue
		worker.close()
	}
	wp.mu.Unlock()
}

// Pause pause
func (wp *WorkPool) Pause() {
	wp.AdjustSize(0)
}

// Continue to work
func (wp *WorkPool) Continue(workerNum ...int) {
	num := int(wp.maxIdle)
	if len(workerNum) > 0 {
		num = workerNum[0]
	}
	wp.AdjustSize(num)
}

// Cap get the number of coroutines
func (wp *WorkPool) Cap() uint {
	return uint(wp.usedNum.Load())
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
		for uint(wp.usedNum.Load()) < wp.minIdle {
			wp.usedNum.Add(1)
			w := wp.workers.Get().(*worker)
			go w.createGoroutines(wp, wp.queue, wp.panicFunc)
			wp.queue <- w
		}
	}
	for wp.minIdle < uint(wp.usedNum.Load()) {
		wp.usedNum.Sub(1)
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
	for uint(wp.usedNum.Load()) < wp.minIdle {
		wp.usedNum.Add(1)
		w := wp.workers.Get().(*worker)
		go w.createGoroutines(wp, wp.queue, wp.panicFunc)
		wp.queue <- w
	}
	wp.mu.Unlock()
	return nil
}

func (w *worker) createGoroutines(wp *WorkPool, q chan<- *worker, handler PanicFunc) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			if err != nil && handler != nil {
				handler(err)
			}
			go w.createGoroutines(wp, q, handler)
			q <- w
		}
	}()
	if wp.releaseTime > 0 {
		timer := time.NewTimer(wp.releaseTime)
		defer timer.Stop()
		for {
			select {
			case job := <-w.jobQueue:
				timer.Stop()
				err := job()
				if err != nil && handler != nil {
					handler(err)
				}
				q <- w
				timer.Reset(wp.releaseTime)
			// case parameter := <-w.Parameter:
			// 	q <- w
			case <-timer.C:
				<-wp.queue
				wp.usedNum.Sub(1)
				return
			case <-w.stop:
				return
			}
		}
	} else {
		for {
			select {
			case job := <-w.jobQueue:
				err := job()
				if err != nil && handler != nil {
					handler(err)
				}
				q <- w
			case <-w.stop:
				return
			}
		}
	}
}

func (w *worker) close() {
	w.stop <- struct{}{}
	close(w.stop)
	close(w.jobQueue)
}
