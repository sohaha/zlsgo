package zutil

import (
	"bytes"
	"sync"
)

// BuffSize is the default initial buffer size for the buffer pool.
// It represents the smallest buffer size that will be allocated.
var BuffSize = uint(32)

// BufferPool implements a pool of bytes.Buffer objects with different capacities.
// It maintains multiple pools of buffers with power-of-two sizes between begin and end.
// This allows for efficient reuse of buffers while minimizing memory waste.
type BufferPool struct {
	// shards contains multiple sync.Pool instances, one for each buffer size
	shards map[int]*sync.Pool
	// begin is the minimum buffer capacity (power of 2)
	begin int
	// end is the maximum buffer capacity (power of 2)
	end int
}

// bufPools is the global buffer pool instance used by GetBuff and PutBuff.
// It provides buffers with sizes from 32 bytes up to 100MB (104857600 bytes).
var bufPools = NewBufferPool(BuffSize, (1<<20)*100)

// NewBufferPool creates a new BufferPool with buffer sizes ranging from
// left to right bytes (rounded up to powers of two).
func NewBufferPool(left, right uint) *BufferPool {
	begin, end := int(roundUpToPowerOfTwo(left)), int(roundUpToPowerOfTwo(right))
	p := &BufferPool{
		begin:  begin,
		end:    end,
		shards: map[int]*sync.Pool{},
	}
	for i := begin; i <= end; i *= 2 {
		capacity := i
		p.shards[i] = &sync.Pool{
			New: func() interface{} { return bytes.NewBuffer(make([]byte, 0, capacity)) },
		}
	}
	return p
}

// Put returns a buffer to the pool.
// If noreset is not provided or false, the buffer will be reset before being returned to the pool.
// If the buffer's capacity doesn't match any pool size, it will be discarded.
func (p *BufferPool) Put(b *bytes.Buffer, noreset ...bool) {
	if b != nil {
		if pool, ok := p.shards[b.Cap()]; ok {
			if len(noreset) == 0 || !noreset[0] {
				b.Reset()
			}

			pool.Put(b)
		}
	}
}

// Get retrieves a buffer from the pool with at least the requested capacity.
// If n is provided, it specifies the minimum capacity of the returned buffer.
// If no buffer of appropriate size is available, a new one will be created.
func (p *BufferPool) Get(n ...uint) *bytes.Buffer {
	size := int(max(uint(p.begin), n...))
	if pool, ok := p.shards[size]; ok {
		b := pool.Get().(*bytes.Buffer)
		if b.Cap() < size {
			b.Grow(size)
			b.Reset()
		}
		return b
	}
	return bytes.NewBuffer(make([]byte, 0, size))
}

// roundUpToPowerOfTwo rounds up a number to the next power of two.
// For example, 15 becomes 16, 17 becomes 32, etc.
func roundUpToPowerOfTwo(v uint) uint {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}

// max determines the maximum buffer size based on the provided parameters.
// It rounds up the requested size to the next power of two and ensures it's
// at least as large as the minimum size a.
func max(a uint, n ...uint) (size uint) {
	if len(n) > 0 && n[0] > 0 {
		size = n[0]
	} else {
		size = BuffSize
	}

	b := roundUpToPowerOfTwo(size)

	if a > b {
		return a
	}

	return b
}

// GetBuff retrieves a buffer from the global buffer pool with at least the requested capacity.
// This is a convenience function that uses the global bufPools instance.
func GetBuff(size ...uint) *bytes.Buffer {
	return bufPools.Get(size...)
}

// PutBuff returns a buffer to the global buffer pool.
// This is a convenience function that uses the global bufPools instance.
func PutBuff(buffer *bytes.Buffer, noreset ...bool) {
	bufPools.Put(buffer, noreset...)
}
