//go:build go1.18
// +build go1.18

package ztype

func ToPointer[T any](value T) *T {
	return &value
}
