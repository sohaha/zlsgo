//go:build go1.18
// +build go1.18

package zutil

// Optional applies a series of option functions to a value of any type.
// This is a generic implementation of the functional options pattern that works with any type.
// It's useful for configuring structs or other values with optional parameters.
// Optional applies configuration functions to a value and returns the modified value.
func Optional[T interface{}](o T, fn ...func(*T)) T {
	for _, f := range fn {
		f(&o)
	}
	return o
}
