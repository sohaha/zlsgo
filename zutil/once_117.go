//go:build !go1.18
// +build !go1.18

package zutil

import (
	"sync"
	"time"
)

// Once initialize the singleton
func Once(fn func() interface{}) func() interface{} {
	var (
		once sync.Once
		ivar interface{}
	)
	return func() interface{} {
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
