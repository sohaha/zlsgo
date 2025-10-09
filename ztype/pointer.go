//go:build go1.18
// +build go1.18

package ztype

// ToPointer returns a pointer to the given value.
// This is a generic function that works with any type T.
func ToPointer[T any](value T) *T {
	return &value
}
