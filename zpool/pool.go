// Package zpool provides a thread-safe work pool implementation.
// It allows for concurrent execution of tasks with configurable limits.
package zpool

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
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
		injector    zdi.Injector
		queue       *workerQueue
		closeCh     chan struct{}
		usedNum     *zutil.Int64
		activeNum   *zutil.Int64
		panicFunc   PanicFunc
		New         func()
		minIdle     uint
		maxIdle     uint
		releaseTime time.Duration
		mu          sync.RWMutex
		closed      bool
	}
	worker struct {
		jobQueue  chan taskfn
		stop      chan struct{}
		Parameter chan []interface{}
		state     int32
		queueElem *list.Element
	}
	PanicFunc func(err error)
)

var (
	ErrPoolClosed  = errors.New("pool has been closed")
	ErrWaitTimeout = errors.New("pool wait timeout")
)

const (
	workerStateIdle int32 = iota
	workerStateBusy
	workerStateClosing
)

type workerQueue struct {
	mu    sync.Mutex
	list  *list.List
	ready chan struct{}
}

func newWorkerQueue() *workerQueue {
	return &workerQueue{
		list:  list.New(),
		ready: make(chan struct{}),
	}
}

func (q *workerQueue) push(w *worker) {
	q.mu.Lock()
	if w.queueElem != nil {
		q.mu.Unlock()
		return
	}
	wasEmpty := q.list.Len() == 0
	w.queueElem = q.list.PushBack(w)
	if wasEmpty {
		ready := q.ready
		q.ready = make(chan struct{})
		close(ready)
	}
	q.mu.Unlock()
}

func (q *workerQueue) tryPop() *worker {
	q.mu.Lock()
	elem := q.list.Front()
	if elem == nil {
		q.mu.Unlock()
		return nil
	}
	w := elem.Value.(*worker)
	q.list.Remove(elem)
	w.queueElem = nil
	atomic.StoreInt32(&w.state, workerStateBusy)
	q.mu.Unlock()
	return w
}

func (q *workerQueue) pop(ctx context.Context, stop <-chan struct{}) (*worker, error) {
	for {
		if w := q.tryPop(); w != nil {
			return w, nil
		}
		q.mu.Lock()
		if q.list.Len() != 0 {
			q.mu.Unlock()
			continue
		}
		wait := q.ready
		q.mu.Unlock()
		if ctx == nil {
			if stop == nil {
				<-wait
				continue
			}
			select {
			case <-stop:
				return nil, ErrPoolClosed
			case <-wait:
				continue
			}
		}
		select {
		case <-ctx.Done():
			return nil, ErrWaitTimeout
		case <-stop:
			return nil, ErrPoolClosed
		case <-wait:
		}
	}
}

func (q *workerQueue) remove(w *worker) bool {
	q.mu.Lock()
	if w.queueElem == nil {
		q.mu.Unlock()
		return false
	}
	q.list.Remove(w.queueElem)
	w.queueElem = nil
	q.mu.Unlock()
	return true
}

// type Options func(*WorkPool)
// // func WithReleaseTime
// func NewCustom(min int, opt Options) *WorkPool {
// 	w := New(min)
// 	if opt != nil {
// 		opt(w)
// 	}
// 	return w
// }

func New(size int, max ...int) *WorkPool {
	minIdle := uint(size)
	if minIdle <= 0 {
		minIdle = 1
	}
	maxIdle := minIdle
	if len(max) > 0 && max[0] > 0 {
		m := uint(max[0])
		if m > maxIdle {
			maxIdle = m
		}
	}

	w := &WorkPool{
		minIdle:     minIdle,
		maxIdle:     maxIdle,
		injector:    zdi.New(),
		queue:       newWorkerQueue(),
		closeCh:     make(chan struct{}),
		usedNum:     zutil.NewInt64(0),
		activeNum:   zutil.NewInt64(0),
		releaseTime: time.Second * 60,
		workers: sync.Pool{New: func() interface{} {
			return &worker{
				jobQueue:  make(chan taskfn),
				Parameter: make(chan []interface{}),
				stop:      make(chan struct{}),
				state:     workerStateIdle,
			}
		}},
	}
	// TODO: periodically write queue to chan

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
	wp.activeNum.Add(1)
	run := func(w *worker) {
		if fn != nil {
			w.jobQueue <- fn
		}
	}
	wp.mu.Lock()
	if wp.closed {
		wp.mu.Unlock()
		wp.activeNum.Sub(1)
		return ErrPoolClosed
	}
	w := wp.queue.tryPop()
	if w != nil {
		wp.mu.Unlock()
		run(w)
		return nil
	}
	if uint(wp.usedNum.Load()) < wp.maxIdle {
		wp.usedNum.Add(1)
		wp.mu.Unlock()
		w := wp.workers.Get().(*worker)
		w.queueElem = nil
		atomic.StoreInt32(&w.state, workerStateBusy)
		go w.createGoroutines(wp, wp.queue, wp.panicFunc)
		run(w)
		return nil
	}
	wp.mu.Unlock()
	w, err := wp.queue.pop(cxt, wp.closeCh)
	if err != nil {
		wp.activeNum.Sub(1)
		return err
	}
	wp.mu.RLock()
	closed := wp.closed
	wp.mu.RUnlock()
	if closed {
		wp.queue.push(w)
		wp.activeNum.Sub(1)
		return ErrPoolClosed
	}
	run(w)
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
	close(wp.closeCh)
	for 0 < uint(wp.usedNum.Load()) {
		wp.usedNum.Sub(1)
		worker, _ := wp.queue.pop(nil, nil)
		worker.close()
	}
	wp.mu.Unlock()
}

// Wait for the task to finish
func (wp *WorkPool) Wait() {
	for 0 < uint(wp.activeNum.Load()) {
		time.Sleep(100 * time.Millisecond)
	}
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
			w.queueElem = nil
			atomic.StoreInt32(&w.state, workerStateIdle)
			go w.createGoroutines(wp, wp.queue, wp.panicFunc)
			wp.queue.push(w)
		}
	}
	for wp.minIdle < uint(wp.usedNum.Load()) {
		wp.usedNum.Sub(1)
		worker, _ := wp.queue.pop(nil, nil)
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
		w.queueElem = nil
		atomic.StoreInt32(&w.state, workerStateIdle)
		go w.createGoroutines(wp, wp.queue, wp.panicFunc)
		wp.queue.push(w)
	}
	wp.mu.Unlock()
	return nil
}

func (w *worker) createGoroutines(wp *WorkPool, q *workerQueue, handler PanicFunc) {
	defer func() {
		if r := recover(); r != nil {
			wp.activeNum.Sub(1)
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			if err != nil && handler != nil {
				handler(err)
			}
			atomic.StoreInt32(&w.state, workerStateIdle)
			go w.createGoroutines(wp, q, handler)
			q.push(w)
		}
	}()
	if wp.releaseTime > 0 {
		timer := time.NewTimer(wp.releaseTime)
		defer timer.Stop()
		for {
			select {
			case job := <-w.jobQueue:
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				err := job()
				if err != nil && handler != nil {
					handler(err)
				}
				wp.activeNum.Sub(1)
				atomic.StoreInt32(&w.state, workerStateIdle)
				q.push(w)
				timer.Reset(wp.releaseTime)
			// case parameter := <-w.Parameter:
			// 	q <- w
			case <-timer.C:
				if q.remove(w) {
					atomic.StoreInt32(&w.state, workerStateClosing)
					wp.usedNum.Sub(1)
					return
				}
				timer.Reset(wp.releaseTime)
			case <-w.stop:
				atomic.StoreInt32(&w.state, workerStateClosing)
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
				wp.activeNum.Sub(1)
				atomic.StoreInt32(&w.state, workerStateIdle)
				q.push(w)
			case <-w.stop:
				atomic.StoreInt32(&w.state, workerStateClosing)
				return
			}
		}
	}
}

func (w *worker) close() {
	atomic.StoreInt32(&w.state, workerStateClosing)
	w.stop <- struct{}{}
}
