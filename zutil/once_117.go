//go:build !go1.18
// +build !go1.18

package zutil

import (
	"sync"
)

// Once initialize the singleton
func Once(fn func() interface{}) func() interface{} {
	var (
		once sync.Once
		ivar interface{}
	)
	return func() interface{} {
		once.Do(func() {
			ivar = fn()
		})
		return ivar
	}
}
