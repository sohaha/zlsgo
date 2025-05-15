package zutil

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"unsafe"
)

type (
	// Bool is an atomic boolean type that can be safely accessed concurrently
	// from multiple goroutines without additional synchronization.
	Bool struct {
		_ Nocmp // Makes the struct uncomparable
		b int32 // 0 means false, 1 means true
	}

	// Int32 is an atomic int32 type that can be safely accessed concurrently
	// from multiple goroutines without additional synchronization.
	Int32 struct {
		_ Nocmp // Makes the struct uncomparable
		v int32 // The actual value
	}

	// Uint32 is an atomic uint32 type that can be safely accessed concurrently
	// from multiple goroutines without additional synchronization.
	Uint32 struct {
		_ Nocmp  // Makes the struct uncomparable
		v uint32 // The actual value
	}

	// Uint64 is an atomic uint64 type that can be safely accessed concurrently
	// from multiple goroutines without additional synchronization.
	Uint64 struct {
		_ Nocmp  // Makes the struct uncomparable
		v uint64 // The actual value
	}

	// Int64 is an atomic int64 type that can be safely accessed concurrently
	// from multiple goroutines without additional synchronization.
	Int64 struct {
		_ Nocmp // Makes the struct uncomparable
		v int64 // The actual value
	}

	// Uintptr is an atomic uintptr type that can be safely accessed concurrently
	// from multiple goroutines without additional synchronization.
	Uintptr struct {
		_ Nocmp   // Makes the struct uncomparable
		v uintptr // The actual value
	}

	// Pointer is an atomic unsafe.Pointer type that can be safely accessed concurrently
	// from multiple goroutines without additional synchronization.
	Pointer struct {
		_ Nocmp          // Makes the struct uncomparable
		v unsafe.Pointer // The actual pointer value
	}
)

// NewBool creates a new atomic Bool with the given initial value.
func NewBool(b bool) *Bool {
	ret := &Bool{}
	if b {
		ret.b = 1
	}

	return ret
}

// Store atomically stores the given value and returns the previous value.
func (b *Bool) Store(val bool) bool {
	var newV int32
	if val {
		newV = 1
	}
	return atomic.SwapInt32(&b.b, newV) == 1
}

// Load atomically loads and returns the current value.
func (b *Bool) Load() bool {
	return atomic.LoadInt32(&b.b) == 1
}

// Toggle atomically negates the boolean value and returns the previous value.
// This is done in a loop to ensure atomicity even under contention.
func (b *Bool) Toggle() (old bool) {
	for {
		old := b.Load()
		if b.CAS(old, !old) {
			return old
		}
	}
}

// CAS (Compare-And-Swap) atomically compares the current value with 'old'
// and, if they match, sets the value to 'new'.
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

// NewInt32 creates a new atomic Int32 with the given initial value.
func NewInt32(i int32) *Int32 {
	return &Int32{
		v: i,
	}
}

// Add atomically adds the given delta to the current value and returns the new value.
func (i32 *Int32) Add(i int32) int32 {
	return atomic.AddInt32(&i32.v, i)
}

// Sub atomically subtracts the given delta from the current value and returns the new value.
func (i32 *Int32) Sub(i int32) int32 {
	return atomic.AddInt32(&i32.v, -i)
}

// Swap atomically stores the given value and returns the previous value.
func (i32 *Int32) Swap(i int32) int32 {
	return atomic.SwapInt32(&i32.v, i)
}

// Load atomically loads and returns the current value.
func (i32 *Int32) Load() int32 {
	return atomic.LoadInt32(&i32.v)
}

// Store atomically stores the given value.
func (i32 *Int32) Store(i int32) {
	atomic.StoreInt32(&i32.v, i)
}

// CAS (Compare-And-Swap) atomically compares the current value with 'old'
// and, if they match, sets the value to 'new'.
func (i32 *Int32) CAS(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&i32.v, old, new)
}

// String returns the string representation of the current value.
func (i32 *Int32) String() string {
	v := i32.Load()
	return strconv.FormatInt(int64(v), 10)
}

// NewUint32 creates a new atomic Uint32 with the given initial value.
func NewUint32(i uint32) *Uint32 {
	return &Uint32{
		v: i,
	}
}

// Add atomically adds the given delta to the current value and returns the new value.
func (u32 *Uint32) Add(i uint32) uint32 {
	return atomic.AddUint32(&u32.v, i)
}

// Sub atomically subtracts the given delta from the current value and returns the new value.
func (u32 *Uint32) Sub(i uint32) uint32 {
	return atomic.AddUint32(&u32.v, ^(i - 1))
}

// Swap atomically stores the given value and returns the previous value.
func (u32 *Uint32) Swap(i uint32) uint32 {
	return atomic.SwapUint32(&u32.v, i)
}

// Load atomically loads and returns the current value.
func (u32 *Uint32) Load() uint32 {
	return atomic.LoadUint32(&u32.v)
}

// Store atomically stores the given value.
func (u32 *Uint32) Store(i uint32) {
	atomic.StoreUint32(&u32.v, i)
}

// CAS (Compare-And-Swap) atomically compares the current value with 'old'
// and, if they match, sets the value to 'new'.
func (u32 *Uint32) CAS(old, new uint32) bool {
	return atomic.CompareAndSwapUint32(&u32.v, old, new)
}

// String returns the string representation of the current value.
func (u32 *Uint32) String() string {
	v := u32.Load()
	return strconv.FormatInt(int64(v), 10)
}

// NewUint64 creates a new atomic Uint64 with the given initial value.
func NewUint64(i uint64) *Uint64 {
	return &Uint64{
		v: i,
	}
}

// Add atomically adds the given delta to the current value and returns the new value.
func (u64 *Uint64) Add(i uint64) uint64 {
	return atomic.AddUint64(&u64.v, i)
}

// Sub atomically subtracts the given delta from the current value and returns the new value.
func (u64 *Uint64) Sub(i uint64) uint64 {
	return atomic.AddUint64(&u64.v, ^(i - 1))
}

// Swap atomically stores the given value and returns the previous value.
func (u64 *Uint64) Swap(i uint64) uint64 {
	return atomic.SwapUint64(&u64.v, i)
}

// Load atomically loads and returns the current value.
func (u64 *Uint64) Load() uint64 {
	return atomic.LoadUint64(&u64.v)
}

// Store atomically stores the given value.
func (u64 *Uint64) Store(i uint64) {
	atomic.StoreUint64(&u64.v, i)
}

// CAS (Compare-And-Swap) atomically compares the current value with 'old'
// and, if they match, sets the value to 'new'.
func (u64 *Uint64) CAS(old, new uint64) bool {
	return atomic.CompareAndSwapUint64(&u64.v, old, new)
}

// String returns the string representation of the current value.
func (u64 *Uint64) String() string {
	v := u64.Load()
	return strconv.FormatInt(int64(v), 10)
}

// NewInt64 creates a new atomic Int64 with the given initial value.
func NewInt64(i int64) *Int64 {
	return &Int64{
		v: i,
	}
}

// Add atomically adds the given delta to the current value and returns the new value.
func (i64 *Int64) Add(i int64) int64 {
	return atomic.AddInt64(&i64.v, i)
}

// Sub atomically subtracts the given delta from the current value and returns the new value.
func (i64 *Int64) Sub(i int64) int64 {
	return atomic.AddInt64(&i64.v, -i)
}

// Swap atomically stores the given value and returns the previous value.
func (i64 *Int64) Swap(i int64) int64 {
	return atomic.SwapInt64(&i64.v, i)
}

// Load atomically loads and returns the current value.
func (i64 *Int64) Load() int64 {
	return atomic.LoadInt64(&i64.v)
}

// Store atomically stores the given value.
func (i64 *Int64) Store(i int64) {
	atomic.StoreInt64(&i64.v, i)
}

// CAS (Compare-And-Swap) atomically compares the current value with 'old'
// and, if they match, sets the value to 'new'.
func (i64 *Int64) CAS(old, new int64) bool {
	return atomic.CompareAndSwapInt64(&i64.v, old, new)
}

// String returns the string representation of the current value.
func (i64 *Int64) String() string {
	v := i64.Load()
	return strconv.FormatInt(v, 10)
}

// NewUintptr creates a new atomic Uintptr with the given initial value.
func NewUintptr(i uintptr) *Uintptr {
	return &Uintptr{
		v: i,
	}
}

// Add atomically adds the given delta to the current value and returns the new value.
func (ptr *Uintptr) Add(i uintptr) uintptr {
	return atomic.AddUintptr(&ptr.v, i)
}

// Sub atomically subtracts the given delta from the current value and returns the new value.
func (ptr *Uintptr) Sub(i uintptr) uintptr {
	return atomic.AddUintptr(&ptr.v, -i)
}

// Swap atomically stores the given value and returns the previous value.
func (ptr *Uintptr) Swap(i uintptr) uintptr {
	return atomic.SwapUintptr(&ptr.v, i)
}

// Load atomically loads and returns the current value.
func (ptr *Uintptr) Load() uintptr {
	return atomic.LoadUintptr(&ptr.v)
}

// Store atomically stores the given value.
func (ptr *Uintptr) Store(i uintptr) {
	atomic.StoreUintptr(&ptr.v, i)
}

// CAS (Compare-And-Swap) atomically compares the current value with 'old'
// and, if they match, sets the value to 'new'.
func (ptr *Uintptr) CAS(old, new uintptr) bool {
	return atomic.CompareAndSwapUintptr(&ptr.v, old, new)
}

// String returns the string representation of the current value.
func (ptr *Uintptr) String() string {
	v := ptr.Load()
	return fmt.Sprintf("%+v", v)
}

// NewPointer creates a new atomic Pointer with the given initial value.
func NewPointer(p unsafe.Pointer) *Pointer {
	return &Pointer{
		v: p,
	}
}

// Load atomically loads and returns the current value.
func (ptr *Pointer) Load() unsafe.Pointer {
	return atomic.LoadPointer(&ptr.v)
}

// Store atomically stores the given value.
func (ptr *Pointer) Store(p unsafe.Pointer) {
	atomic.StorePointer(&ptr.v, p)
}

// CAS (Compare-And-Swap) atomically compares the current value with 'old'
// and, if they match, sets the value to 'new'.
func (ptr *Pointer) CAS(old, new unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&ptr.v, old, new)
}

// String returns the string representation of the current value.
func (ptr *Pointer) String() string {
	v := ptr.Load()
	return fmt.Sprintf("%+v", v)
}
