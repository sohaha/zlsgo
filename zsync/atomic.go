//go:build go1.18
// +build go1.18

package zsync

import (
	"sync/atomic"

	"github.com/sohaha/zlsgo/zutil"
)

// AtomicValue is the generic version of [atomic.Value].
type AtomicValue[T any] struct {
	_ zutil.Nocmp
	v atomic.Value
}

type wrappedValue[T any] struct{ v T }

func NewValue[T any](v T) *AtomicValue[T] {
	av := &AtomicValue[T]{}
	av.v.Store(wrappedValue[T]{v})
	return av
}

// Load returns the value set by the most recent Store.
// It returns the zero value for T if the value is empty.
func (v *AtomicValue[T]) Load() T {
	x := v.v.Load()
	if x != nil {
		return x.(wrappedValue[T]).v
	}
	var zero T
	return zero
}

// Store sets the value of the Value to x.
func (v *AtomicValue[T]) Store(x T) {
	v.v.Store(wrappedValue[T]{x})
}

// Swap stores new into Value and returns the previous value.
func (v *AtomicValue[T]) Swap(x T) (old T) {
	oldV := v.v.Swap(wrappedValue[T]{x})
	if oldV != nil {
		return oldV.(wrappedValue[T]).v
	}
	return old
}

// CAS executes the compare-and-swap operation for the Value.
func (v *AtomicValue[T]) CAS(oldV, newV T) (swapped bool) {
	return v.v.CompareAndSwap(wrappedValue[T]{oldV}, wrappedValue[T]{newV})
}
