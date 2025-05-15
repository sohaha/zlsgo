//go:build go1.18
// +build go1.18

package zarray

import (
	"sync/atomic"
	"unsafe"

	"github.com/sohaha/zlsgo/zutil"
)

// Element deletion status constants
const (
	// notDeleted indicates an active element
	notDeleted uint32 = iota
	// deleted indicates a logically deleted element
	deleted
)

// atomicPointer provides atomic operations for pointer types using generics
type atomicPointer[T any] struct {
	_   zutil.Nocmp
	ptr unsafe.Pointer
}

// Load atomically loads and returns the pointer value
func (p *atomicPointer[T]) Load() *T     { return (*T)(atomic.LoadPointer(&p.ptr)) }
// Store atomically stores the provided pointer value
func (p *atomicPointer[T]) Store(v *T)   { atomic.StorePointer(&p.ptr, unsafe.Pointer(v)) }
// Swap atomically stores the provided pointer value and returns the previous value
func (p *atomicPointer[T]) Swap(v *T) *T { return (*T)(atomic.SwapPointer(&p.ptr, unsafe.Pointer(v))) }
// CompareAndSwap atomically swaps the pointer value if the current value matches the old value
// Returns true if the swap was performed, false otherwise
func (p *atomicPointer[T]) CompareAndSwap(old, new *T) bool {
	return atomic.CompareAndSwapPointer(&p.ptr, unsafe.Pointer(old), unsafe.Pointer(new))
}

// newListHead creates a new sentinel element that serves as the head of a linked list
func newListHead[K hashable, V any]() *element[K, V] {
	e := &element[K, V]{keyHash: 0, key: *new(K)}
	e.nextPtr.Store(nil)
	e.value.Store(new(V))
	return e
}

// element represents a node in a concurrent linked list that stores key-value pairs
type element[K hashable, V any] struct {
	key     K
	nextPtr atomicPointer[element[K, V]]
	value   atomicPointer[V]
	keyHash uintptr
	deleted uint32
}

// next returns the next non-deleted element in the list
// It also performs cleanup by removing deleted elements from the list
func (self *element[K, V]) next() *element[K, V] {
	for nextElement := self.nextPtr.Load(); nextElement != nil; {
		if nextElement.isDeleted() {
			self.nextPtr.CompareAndSwap(nextElement, nextElement.next())
			nextElement = self.nextPtr.Load()
		} else {
			return nextElement
		}
	}
	return nil
}

// addBefore inserts the allocatedElement before the specified element
// Returns true if the insertion was successful, false otherwise
func (self *element[K, V]) addBefore(allocatedElement, before *element[K, V]) bool {
	if self.next() != before {
		return false
	}
	allocatedElement.nextPtr.Store(before)
	return self.nextPtr.CompareAndSwap(before, allocatedElement)
}

// inject inserts a new key-value pair into the list or updates an existing one
// Returns the element and a boolean indicating whether a new element was created
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

// search looks for an element with the specified hash and key
// Returns the element before the target position, the found element (or nil if not found),
// and the element after the target position
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
		if c < curr.keyHash {
			right = curr
			curr = nil
			return left, curr, right
		} else if c == curr.keyHash && key == curr.key {
			return left, curr, right
		}
		left = curr
		curr = left.next()
		right = nil
	}
}

// remove marks the element as deleted
// Returns true if the element was successfully marked as deleted, false if it was already deleted
func (self *element[K, V]) remove() bool {
	return atomic.CompareAndSwapUint32(&self.deleted, notDeleted, deleted)
}

// isDeleted checks if the element has been marked as deleted
// Returns true if the element is deleted, false otherwise
func (self *element[K, V]) isDeleted() bool {
	return atomic.LoadUint32(&self.deleted) == deleted
}
