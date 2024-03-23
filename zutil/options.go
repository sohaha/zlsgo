//go:build go1.18
// +build go1.18

package zutil

// Optional Optional parameter
func Optional[T interface{}](o T, fn ...func(*T)) T {
	for _, f := range fn {
		f(&o)
	}
	return o
}
