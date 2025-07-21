//go:build go1.18
// +build go1.18

package zutil

import (
	"errors"
	"sync"
)

// Once creates a function that ensures the provided initialization function
// is executed only once, regardless of how many times the returned function is called.
// This implements the singleton pattern with built-in error recovery.
//
// If the initialization function panics, the Once state is reset after a delay,
// allowing for a retry on the next call.
func Once[T any](fn func() T) func() T {
	once := OnceWithError(func() (T, error) {
		return fn(), nil
	})

	return func() T {
		resp, _ := once()
		return resp
	}
}

// OnceWithError creates a function that ensures the provided initialization function
// is executed only once, regardless of how many times the returned function is called.
// This implements the singleton pattern with built-in error recovery.
func OnceWithError[T any](fn func() (T, error)) func() (T, error) {
	var (
		mu   sync.Mutex
		done = NewUint32(0)
		ivar T
		ierr error
	)

	return func() (T, error) {
		if done.Load() == 1 {
			return ivar, ierr
		}

		mu.Lock()
		defer mu.Unlock()

		if done.Load() == 0 {
			ierr = TryCatch(func() error {
				ivar, ierr = fn()
				return ierr
			})

			if ierr == nil {
				done.Store(1)
			}
		}

		return ivar, ierr
	}
}

// Guard creates a function that ensures mutually exclusive execution of the provided function.
// If the returned function is called while a previous call is still in progress,
// it will return an error instead of executing the function again.
//
// This is useful for preventing concurrent execution of functions that are not thread-safe
// or for rate-limiting access to resources.
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
