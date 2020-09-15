package zpool

import (
	"errors"
	"sync"
)

type (
	// Task Define function callbacks
	Task     func()
	workPool struct {
		workers        sync.Pool
		closed         bool
		mux            sync.RWMutex
		workerQueue    chan *worker
		workerSum      uint
		workesAliveNum uint
		maxWorkerSum   uint
	}
	worker struct {
		jobQueue  chan Task
		stop      chan struct{}
		Parameter chan []interface{}
	}
)

var (
	ErrPoolClosed = errors.New("pool has been closed")
)

func New(workerNum int, maxWorkerNum ...int) *workPool {
	workerSum := uint(workerNum)
	if workerSum <= 0 {
		workerSum = 1
	}
	maxWorkerSum := workerSum
	if len(maxWorkerNum) > 0 && maxWorkerNum[0] > 0 {
		max := uint(maxWorkerNum[0])
		if max > maxWorkerSum {
			maxWorkerSum = max
		}
	}

	w := &workPool{
		workerSum:    workerSum,
		maxWorkerSum: maxWorkerSum,
		workerQueue:  make(chan *worker, maxWorkerSum),
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
func (wp *workPool) Do(fn Task) error {
	return wp.do(fn, nil)
}

func (wp *workPool) do(fn Task, param []interface{}) error {
	if wp.IsClosed() {
		return ErrPoolClosed
	}
	wp.mux.Lock()
	run := func(w *worker) {
		if fn != nil {
			w.jobQueue <- fn
		}
	}
	select {
	case w := <-wp.workerQueue:
		wp.mux.Unlock()
		if w != nil {
			run(w)
		} else {
			return ErrPoolClosed
		}
	default:
		switch {
		case wp.workesAliveNum == wp.workerSum:
			wp.mux.Unlock()
			w := <-wp.workerQueue
			if w != nil {
				run(w)
			} else {
				return ErrPoolClosed
			}
		case wp.workesAliveNum < wp.workerSum:
			wp.workesAliveNum++
			wp.mux.Unlock()
			w := wp.workers.Get().(*worker)
			w.createGoroutines(wp.workerQueue)
			run(w)
		default:
			wp.mux.Unlock()
		}
	}
	return nil
}

// IsClosed Has it been closed
func (wp *workPool) IsClosed() bool {
	wp.mux.RLock()
	b := wp.closed
	wp.mux.RUnlock()
	return b
}

// Close close the pool
func (wp *workPool) Close() {
	if wp.IsClosed() {
		return
	}
	wp.mux.Lock()
	wp.closed = true
	for 0 < wp.workesAliveNum {
		wp.workesAliveNum--
		worker := <-wp.workerQueue
		worker.close()
	}
	wp.mux.Unlock()
}

// Pause pause
func (wp *workPool) Pause() {
	wp.AdjustSize(0)
}

// Continue continue
func (wp *workPool) Continue(workerNum ...int) {
	num := int(wp.maxWorkerSum)
	if len(workerNum) > 0 {
		num = workerNum[0]
	}
	wp.AdjustSize(num)
}

// Cap get the number of coroutines
func (wp *workPool) Cap() uint {
	wp.mux.RLock()
	defer wp.mux.RUnlock()
	return wp.workesAliveNum
}

// AdjustSize adjust the pool size
func (wp *workPool) AdjustSize(workSize int) {
	wp.mux.Lock()
	defer wp.mux.Unlock()
	if wp.closed {
		return
	}

	oldSize := wp.workerSum
	newSize := uint(workSize)
	if newSize > wp.maxWorkerSum {
		newSize = wp.maxWorkerSum
	}
	wp.workerSum = newSize

	if workSize > 0 && oldSize < wp.workerSum {
		for wp.workesAliveNum < wp.workerSum {
			wp.workesAliveNum++
			w := wp.workers.Get().(*worker)
			w.createGoroutines(wp.workerQueue)
			wp.workerQueue <- w
		}
	}
	for wp.workerSum < wp.workesAliveNum {
		wp.workesAliveNum--
		worker := <-wp.workerQueue
		worker.stop <- struct{}{}
		wp.workers.Put(worker)
	}
}

func (wp *workPool) PreInit() error {
	if wp.IsClosed() {
		return ErrPoolClosed
	}
	wp.mux.Lock()
	for wp.workesAliveNum < wp.workerSum {
		wp.workesAliveNum++
		w := wp.workers.Get().(*worker)
		w.createGoroutines(wp.workerQueue)
		wp.workerQueue <- w
	}
	wp.mux.Unlock()
	return nil
}

func (w *worker) createGoroutines(q chan<- *worker) {
	go func() {
		for {
			select {
			case job := <-w.jobQueue:
				job()
				q <- w
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
