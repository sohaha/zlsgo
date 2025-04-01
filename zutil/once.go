//go:build go1.18
// +build go1.18

package zutil

import (
	"errors"
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

// Guard ensures mutually exclusive execution
func Guard[T any](fn func() T) func() (T, error) {
	status := NewBool(false)
	return func() (resp T, err error) {
		if !status.CAS(false, true) {
			return resp, errors.New("already running")
		}
		defer status.Store(false)

		err = TryCatch(func() error {
			resp = fn()
			return nil
		})
		return resp, err
	}
}
