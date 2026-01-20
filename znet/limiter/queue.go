package limiter

import (
	"errors"
	"time"

	"github.com/sohaha/zlsgo/zsync"
)

type circleQueue struct {
	slice   []int64
	maxSize int
	head    int
	tail    int
	mu      *zsync.RBMutex
}

// newCircleQueue Initialize ring queue
func newCircleQueue(size int) *circleQueue {
	var c circleQueue
	c.maxSize = size + 1
	c.slice = make([]int64, c.maxSize)
	c.mu = zsync.NewRBMutex()
	return &c
}

func (c *circleQueue) push(val int64) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if (c.tail+1)%c.maxSize == c.head {
		return errors.New("queue is full")
	}
	c.slice[c.tail] = val
	c.tail = (c.tail + 1) % c.maxSize
	return
}

func (c *circleQueue) pop() (val int64, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.tail == c.head {
		return 0, errors.New("queue is empty")
	}
	val = c.slice[c.head]
	c.head = (c.head + 1) % c.maxSize
	return
}

func (c *circleQueue) isFull() bool {
	t := c.mu.RLock()
	defer c.mu.RUnlock(t)
	return (c.tail+1)%c.maxSize == c.head
}

func (c *circleQueue) isEmpty() bool {
	t := c.mu.RLock()
	defer c.mu.RUnlock(t)
	return c.tail == c.head
}

func (c *circleQueue) usedSize() int {
	t := c.mu.RLock()
	defer c.mu.RUnlock(t)
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
	c.mu.Lock()
	defer c.mu.Unlock()
	for c.tail != c.head {
		if now > c.slice[c.head] {
			c.head = (c.head + 1) % c.maxSize
			continue
		}
		break
	}
}
