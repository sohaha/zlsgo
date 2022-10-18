//go:build go1.18
// +build go1.18

package zarray

import (
	"sync/atomic"
	"unsafe"

	"github.com/sohaha/zlsgo/zutil"
)

const marked = ^uintptr(0)

type atomicPointer[T any] struct {
	_   zutil.Nocmp
	ptr unsafe.Pointer
}

func (p *atomicPointer[T]) Load() *T     { return (*T)(atomic.LoadPointer(&p.ptr)) }
func (p *atomicPointer[T]) Store(v *T)   { atomic.StorePointer(&p.ptr, unsafe.Pointer(v)) }
func (p *atomicPointer[T]) Swap(v *T) *T { return (*T)(atomic.SwapPointer(&p.ptr, unsafe.Pointer(v))) }
func (p *atomicPointer[T]) CompareAndSwap(old, new *T) bool {
	return atomic.CompareAndSwapPointer(&p.ptr, unsafe.Pointer(old), unsafe.Pointer(new))
}

func newListHead[K hashable, V any]() *element[K, V] {
	e := &element[K, V]{keyHash: marked, key: *new(K)}
	e.nextPtr.Store(nil)
	e.value.Store(new(V))
	return e
}

type element[K hashable, V any] struct {
	key     K
	nextPtr atomicPointer[element[K, V]]
	value   atomicPointer[V]
	keyHash uintptr
}

func (self *element[K, V]) next() *element[K, V] {
	for nextElement := self.nextPtr.Load(); nextElement != nil; {
		if nextElement.keyHash == marked {
			return nextElement.next()
		}
		if nextElement.isDeleted() {
			self.nextPtr.CompareAndSwap(nextElement, nextElement.next())
			nextElement = self.nextPtr.Load()
		} else {
			return nextElement
		}
	}
	return nil
}

func (self *element[K, V]) addBefore(allocatedElement, before *element[K, V]) bool {
	if self.next() != before {
		return false
	}
	allocatedElement.nextPtr.Store(before)
	return self.nextPtr.CompareAndSwap(before, allocatedElement)
}

func (self *element[K, V]) inject(c uintptr, key K, value *V) (*element[K, V], bool) {
	var (
		alloc             *element[K, V]
		left, curr, right = self.search(c, key)
	)
	if curr != nil {
		curr.value.Store(value)
		return curr, false
	}
	if left != nil {
		alloc = &element[K, V]{keyHash: c, key: key}
		alloc.value.Store(value)
		if left.addBefore(alloc, right) {
			return alloc, true
		}
	}
	return nil, false
}

func (self *element[K, V]) search(c uintptr, key K) (*element[K, V], *element[K, V], *element[K, V]) {
	var (
		left, right *element[K, V]
		curr        = self
	)
	for {
		if curr == nil {
			return left, curr, right
		}
		right = curr.next()
		if curr.keyHash != marked {
			if c < curr.keyHash {
				right = curr
				curr = nil
				return left, curr, right
			} else if c == curr.keyHash && key == curr.key {
				return left, curr, right
			}
		}
		left = curr
		curr = left.next()
		right = nil
	}
}

func (self *element[K, V]) remove() {
	deletionNode := &element[K, V]{keyHash: marked}
	for !self.isDeleted() && !self.addBefore(deletionNode, self.next()) {
	}
}

func (self *element[K, V]) isDeleted() bool {
	if next := self.nextPtr.Load(); next != nil {
		return next.keyHash == marked
	}
	return false
}
