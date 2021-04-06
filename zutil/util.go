// Package zutil daily development helper functions
package zutil

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

func WithLockContext(mu *sync.Mutex, fn func()) {
	mu.Lock()
	defer mu.Unlock()
	fn()
}

func WithRunTimeContext(handler func()) time.Duration {
	start := time.Now()
	handler()
	return time.Since(start)
}

func WithRunMemContext(handler func()) uint64 {
	var mem = runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem := mem.TotalAlloc
	handler()
	runtime.ReadMemStats(&mem)
	return mem.TotalAlloc - curMem
}

// IfVal Simulate ternary calculations, pay attention to handling no variables or indexing problems
func IfVal(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

// TryCatch exception capture
func TryCatch(fn func() error) (err error) {
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			if e, ok := recoverErr.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", recoverErr)
			}
		}
	}()
	err = fn()
	return
}

// Deprecated: please use TryCatch
// Try exception capture
func Try(fn func(), catch func(e interface{}), finally ...func()) {
	if len(finally) > 0 {
		defer func() {
			finally[0]()
		}()
	}
	defer func() {
		if err := recover(); err != nil {
			if catch != nil {
				catch(err)
			} else {
				panic(err)
			}
		}
	}()
	fn()
}

// CheckErr CheckErr
func CheckErr(err error, exit ...bool) {
	if err != nil {
		if len(exit) > 0 && exit[0] {
			fmt.Println(err)
			os.Exit(1)
			return
		}
		panic(err)
	}
}

