//go:build go1.18
// +build go1.18

package zutil

// IfVal Simulate ternary calculations, pay attention to handling no variables or indexing problems
func IfVal[T interface{}](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}
