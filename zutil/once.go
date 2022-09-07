//go:build go1.18
// +build go1.18

package zutil

import (
	"sync"
)

// Once initialize the singleton
func Once[T any](fn func() T) func() T {
	var (
		once sync.Once
		ivar T
	)
	return func() T {
		once.Do(func() {
			ivar = fn()
		})
		return ivar
	}
}
