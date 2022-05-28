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
	Int64 struct {
		_ Nocmp
		v int64
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

func (i32 *Int32) Add(i int32) int32 {
	return atomic.AddInt32(&i32.v, i)
}

func (i32 *Int32) Sub(i int32) int32 {
	return atomic.AddInt32(&i32.v, -i)
}

func (i32 *Int32) Swap(i int32) int32 {
	return atomic.SwapInt32(&i32.v, i)
}

func (i32 *Int32) Load() int32 {
	return atomic.LoadInt32(&i32.v)
}

func (i32 *Int32) Store(i int32) {
	atomic.StoreInt32(&i32.v, i)
}

func (i32 *Int32) CAS(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&i32.v, old, new)
}

func (i32 *Int32) String() string {
	v := i32.Load()
	return strconv.FormatInt(int64(v), 10)
}

func NewInt64(i int64) *Int64 {
	return &Int64{
		v: i,
	}
}

func (i64 *Int64) Add(i int64) int64 {
	return atomic.AddInt64(&i64.v, i)
}

func (i64 *Int64) Sub(i int64) int64 {
	return atomic.AddInt64(&i64.v, -i)
}

func (i64 *Int64) Swap(i int64) int64 {
	return atomic.SwapInt64(&i64.v, i)
}

func (i64 *Int64) Load() int64 {
	return atomic.LoadInt64(&i64.v)
}

func (i64 *Int64) Store(i int64) {
	atomic.StoreInt64(&i64.v, i)
}

func (i64 *Int64) CAS(old, new int64) bool {
	return atomic.CompareAndSwapInt64(&i64.v, old, new)
}

func (i64 *Int64) String() string {
	v := i64.Load()
	return strconv.FormatInt(v, 10)
}
