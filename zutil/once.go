//go:build go1.18
// +build go1.18

package zutil

import (
	"sync"
	"time"
)

// Once initialize the singleton
func Once[T any](fn func() T) func() T {
	var (
		once sync.Once
		ivar T
	)
	return func() T {
		once.Do(func() {
			err := TryCatch(func() error {
				ivar = fn()
				return nil
			})
			if err != nil {
				go func() {
					time.Sleep(time.Second)
					once = sync.Once{}
				}()
			}
		})

		return ivar
	}
}
