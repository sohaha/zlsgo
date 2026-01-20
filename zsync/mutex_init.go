package zsync

import (
	"sync/atomic"
	"unsafe"
)

func EnsureRBMutex(mu **RBMutex) *RBMutex {
	if mu == nil {
		return NewRBMutex()
	}
	ptr := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(mu)))
	if ptr != nil {
		return (*RBMutex)(ptr)
	}
	next := NewRBMutex()
	if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(mu)), nil, unsafe.Pointer(next)) {
		return next
	}
	ptr = atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(mu)))
	if ptr != nil {
		return (*RBMutex)(ptr)
	}
	return next
}
