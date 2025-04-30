package zutil

import (
	"bytes"
	"sync"
)

var BuffSize = uint(32)

type BufferPool struct {
	shards map[int]*sync.Pool
	begin  int
	end    int
}

var bufPools = NewBufferPool(BuffSize, (1<<20)*100)

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

func GetBuff(size ...uint) *bytes.Buffer {
	return bufPools.Get(size...)
}

func PutBuff(buffer *bytes.Buffer, noreset ...bool) {
	bufPools.Put(buffer, noreset...)
}
