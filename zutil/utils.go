// Package zutil provides utility functions and types for daily development tasks.
package zutil

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

type (
	// Stack represents a call stack as an array of program counters.
	// It provides methods for formatting and analyzing the call stack.
	Stack []uintptr

	// Nocmp is an uncomparable struct that can be embedded in other structs
	// to make them uncomparable (cannot be compared with == or !=).
	Nocmp [0]func()

	// namedArgs is an internal type used to associate a name with an argument value.
	namedArgs struct {
		arg  interface{}
		name string
	}
)

// Named creates a named argument by associating a name with a value.
// This is useful for functions that accept variadic arguments and need to
// distinguish between different argument types or purposes.
func Named(name string, arg interface{}) interface{} {
	return namedArgs{
		name: name,
		arg:  arg,
	}
}

const (
	// maxStackDepth is the maximum depth of call stack frames to capture.
	maxStackDepth = 1 << 5 // 32 frames
)

// WithRunContext measures the execution time and memory allocation of a function.
// It returns the duration of execution and the number of bytes allocated during execution.
func WithRunContext(handler func()) (time.Duration, int64) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	startMem := mem.Alloc
	start := time.Now()
	
	handler()
	
	duration := time.Since(start)
	runtime.ReadMemStats(&mem)
	return duration, int64(mem.Alloc - startMem)
}

// TryCatch executes a function and captures any panic that occurs, converting it to an error.
// If the function returns an error normally, that error is returned.
// If a panic occurs, it is converted to an error and returned.
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

// Try executes a function and captures any panic that occurs, passing it to the catch function.
// If a finally function is provided, it is always executed after the main function,
// regardless of whether a panic occurred.
// Deprecated: please use zerror.TryCatch instead.
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

// CheckErr checks if an error is not nil and panics if it is.
// If exit is true, it prints the error and exits the program instead of panicking.
// Deprecated: please use zerror.Panic instead.
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

// Callers returns a stack trace as a Stack.
// The optional skip parameter indicates how many stack frames to skip before
// starting to collect the stack trace.
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

var (
	h = []byte{104, 97}
	s = []byte{115}
	o = []byte{111}
	g = []byte{103}
	u = []byte{116, 104, 117, 98, 46, 99, 111, 109, 47}
	l string
	t string
)

func init() {
	l = string(append([]byte{103, 105}, append(u, append(s, append(o, append(h, h...)...)...)...)...))
	t = "_test." + string(append(g, o...))
}

// Format iterates through the stack frames and calls the provided function for each frame.
// The function receives the runtime.Func object, file name, and line number for each frame.
// If the function returns false, iteration stops.
// Note: Frames from the zlsgo library itself are automatically skipped.
func (s Stack) Format(f func(fn *runtime.Func, file string, line int) bool) {
	if s == nil {
		return
	}
	for _, p := range s {
		if fn := runtime.FuncForPC(p - 1); fn != nil {
			file, line := fn.FileLine(p - 1)
			name := fn.Name()
			if !strings.HasSuffix(file, t) && strings.Contains(name, l) {
				continue
			}
			if !f(fn, file, line) {
				break
			}
		}
	}
}
