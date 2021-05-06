package limiter

import (
	"errors"
	"sync"
	"time"
)

type circleQueue struct {
	maxSize int
	slice   []int64
	head    int
	tail    int
	sync.RWMutex
}

// newCircleQueue Initialize ring queue
func newCircleQueue(size int) *circleQueue {
	var c circleQueue
	c.maxSize = size + 1
	c.slice = make([]int64, c.maxSize)
	return &c
}

func (c *circleQueue) push(val int64) (err error) {
	if c.isFull() {
		return errors.New("queue is full")
	}
	c.slice[c.tail] = val
	c.tail = (c.tail + 1) % c.maxSize
	return
}

func (c *circleQueue) pop() (val int64, err error) {
	if c.isEmpty() {
		return 0, errors.New("queue is empty")
	}
	c.Lock()
	defer c.Unlock()
	val = c.slice[c.head]
	c.head = (c.head + 1) % c.maxSize
	return
}

func (c *circleQueue) isFull() bool {
	return (c.tail+1)%c.maxSize == c.head
}

func (c *circleQueue) isEmpty() bool {
	return c.tail == c.head
}

func (c *circleQueue) usedSize() int {
	c.RLock()
	defer c.RUnlock()
	return (c.tail + c.maxSize - c.head) % c.maxSize
}

func (c *circleQueue) unUsedSize() int {
	return c.maxSize - 1 - c.usedSize()
}

func (c *circleQueue) size() int {
	return c.maxSize - 1
}

func (c *circleQueue) deleteExpired() {
	now := time.Now().UnixNano()
	size := c.usedSize()
	if size == 0 {
		return
	}
	for i := 0; i < size; i++ {
		if now > c.slice[c.head] {
			_, _ = c.pop()
		} else {
			return
		}
	}
}
