// Package zutil daily development helper functions
package zutil

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

type (
	// Stack uintptr array
	Stack []uintptr
	// Nocmp is an uncomparable struct
	Nocmp     [0]func()
	namedArgs struct {
		arg  interface{}
		name string
	}
)

// Named creates a named argument
func Named(name string, arg interface{}) interface{} {
	return namedArgs{
		name: name,
		arg:  arg,
	}
}

const (
	maxStackDepth = 1 << 5
)

// WithRunContext function execution time and memory
func WithRunContext(handler func()) (time.Duration, uint64) {
	start, mem := time.Now(), runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem := mem.Alloc
	handler()
	runtime.ReadMemStats(&mem)
	return time.Since(start), mem.Alloc - curMem
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

// Deprecated: please use zerror.TryCatch
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

// Deprecated: please use zerror.Panic
// CheckErr Check Err
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

func Callers(skip ...int) Stack {
	var (
		pcs [maxStackDepth]uintptr
		n   = 0
	)
	if len(skip) > 0 {
		n += skip[0]
	}
	return pcs[:runtime.Callers(n, pcs[:])]
}

func (s Stack) Format(f func(fn *runtime.Func, file string, line int) bool) {
	if s == nil {
		return
	}
	for _, p := range s {
		if fn := runtime.FuncForPC(p - 1); fn != nil {
			file, line := fn.FileLine(p - 1)
			name := fn.Name()
			if !strings.HasSuffix(file, "_test.go") && strings.Contains(name, "github.com/sohaha") {
				continue
			}
			if !f(fn, file, line) {
				break
			}
		}
	}
}
