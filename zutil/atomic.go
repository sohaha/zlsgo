package zutil

import (
	"strconv"
	"sync/atomic"
)

type (
	Bool struct {
		_ Nocmp
		b int32
	}
	Int32 struct {
		_ Nocmp
		v int32
	}
)

func NewBool(b bool) *Bool {
	ret := &Bool{}
	if b {
		ret.b = 1
	}

	return ret
}

func (b *Bool) Store(val bool) bool {
	var newV int32
	if val {
		newV = 1
	}
	return atomic.SwapInt32(&b.b, newV) == 1
}

func (b *Bool) Load() bool {
	return atomic.LoadInt32(&b.b) == 1
}

func (b *Bool) Toggle() (old bool) {
	for {
		old := b.Load()
		if b.CAS(old, !old) {
			return old
		}
	}
}

func (b *Bool) CAS(old, new bool) bool {
	var o, n int32

	if old {
		o = 1
	}
	if new {
		n = 1
	}

	return atomic.CompareAndSwapInt32(&b.b, o, n)
}

func NewInt32(i int32) *Int32 {
	return &Int32{
		v: i,
	}
}

func (i3 *Int32) Add(i int32) int32 {
	return atomic.AddInt32(&i3.v, i)
}

func (i3 *Int32) Sub(i int32) int32 {
	return atomic.AddInt32(&i3.v, -i)
}

func (i3 *Int32) Swap(i int32) int32 {
	return atomic.SwapInt32(&i3.v, i)
}

func (i3 *Int32) Load() int32 {
	return atomic.LoadInt32(&i3.v)
}

func (i3 *Int32) Store(i int32) {
	atomic.StoreInt32(&i3.v, i)
}

func (i3 *Int32) CAS(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&i3.v, old, new)
}

func (i3 *Int32) String() string {
	v := i3.Load()
	return strconv.FormatInt(int64(v), 10)
}
