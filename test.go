// Package zlsgo is a collection of commonly used functions for golang daily development.
package zlsgo

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// TestUtil provides testing utilities to simplify test assertions and validations
type TestUtil struct {
	t testing.TB
}

// NewTest creates a new TestUtil instance with the given testing.TB implementation
func NewTest(t testing.TB) *TestUtil {
	return &TestUtil{
		t: t,
	}
}

// GetCallerInfo returns the file name and line number of the test function caller
func (u *TestUtil) GetCallerInfo() string {
	var info string

	for i := 0; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		basename := file
		if !strings.HasSuffix(basename, "_test.go") {
			continue
		}

		funcName := runtime.FuncForPC(pc).Name()
		index := strings.LastIndex(funcName, ".Test")
		if index == -1 {
			index = strings.LastIndex(funcName, ".Benchmark")
			if index == -1 {
				continue
			}
		}
		funcName = funcName[index+1:]

		if index := strings.IndexByte(funcName, '.'); index > -1 {
			// funcName = funcName[:index]
			// info = funcName + "(" + basename + ":" + strconv.Itoa(line) + ")"
			info = basename + ":" + strconv.Itoa(line)
			continue
		}

		info = basename + ":" + strconv.Itoa(line)
		break
	}

	if info == "" {
		info = "<Unable to get information>"
	}
	return info
}

// Equal compares expected and actual values using deep equality check
// Returns true if values are equal, false otherwise
// If exit is true and values are not equal, test will immediately fail
func (u *TestUtil) Equal(expected, actual interface{}, exit ...bool) bool {
	if !reflect.DeepEqual(expected, actual) {
		u.t.Helper()
		fmt.Printf("        %s 期待:%v (type %v) - 结果:%v (type %v)\n", u.PrintMyName(), expected, reflect.TypeOf(expected), actual, reflect.TypeOf(actual))
		if len(exit) > 0 && exit[0] {
			u.t.FailNow()
		} else {
			u.t.Fail()
		}
		return false
	}
	return true
}

func (u *TestUtil) NoEqual(expected, actual interface{}, exit ...bool) bool {
	if reflect.DeepEqual(expected, actual) {
		u.t.Helper()
		fmt.Printf("        %s 期待不等于:%v (type %v) - 结果:%v (type %v)\n", u.PrintMyName(), expected, reflect.TypeOf(expected), actual, reflect.TypeOf(actual))
		if len(exit) > 0 && exit[0] {
			u.t.FailNow()
		} else {
			u.t.Fail()
		}
		return false
	}
	return true
}

// EqualTrue asserts that the actual value is true
// If exit is true and actual is not true, test will immediately fail
func (u *TestUtil) EqualTrue(actual interface{}, exit ...bool) {
	u.Equal(true, actual, exit...)
}

// EqualFalse asserts that the actual value is false
func (u *TestUtil) EqualFalse(actual interface{}, exit ...bool) {
	u.Equal(false, actual, exit...)
}

// EqualNil asserts that the actual value is nil
// If exit is true and actual is not nil, test will immediately fail
func (u *TestUtil) EqualNil(actual interface{}, exit ...bool) {
	u.Equal(nil, actual, exit...)
}

// NoError asserts that the error is nil
// Returns true if error is nil, false otherwise
// If exit is true and error is not nil, test will immediately fail
func (u *TestUtil) NoError(err error, exit ...bool) bool {
	if err == nil {
		return true
	}

	fmt.Printf("    %s Error: %s\n", u.PrintMyName(), err)

	if len(exit) > 0 && exit[0] {
		u.t.FailNow()
	} else {
		u.t.Fail()
	}
	return false
}

// EqualExit compares expected and actual values and immediately fails the test if not equal
func (u *TestUtil) EqualExit(expected, actual interface{}) {
	u.Equal(expected, actual, true)
}

// Log logs the given values to the test output
func (u *TestUtil) Log(v ...interface{}) {
	u.t.Helper()
	u.t.Log(v...)
}

// Logf logs the formatted string to the test output
func (u *TestUtil) Logf(format string, args ...interface{}) {
	u.t.Helper()
	u.t.Logf(format, args...)
}

// Fatal logs the given values to the test output and immediately fails the test
func (u *TestUtil) Fatal(v ...interface{}) {
	u.t.Helper()
	u.t.Fatal(v...)
}

// PrintMyName returns the caller information for the current test
func (u *TestUtil) PrintMyName() string {
	return u.GetCallerInfo()
}

// Run runs a subtest with the given name and function
func (u *TestUtil) Run(name string, f func(tt *TestUtil)) {
	u.t.Helper()
	u.t.(*testing.T).Run(name, func(t *testing.T) {
		f(NewTest(t))
	})
}

// T returns the underlying *testing.T object
func (u *TestUtil) T() *testing.T {
	return u.t.(*testing.T)
}

// IsNil asserts that the actual value is nil
// Returns true if actual is nil, false otherwise
// If exit is true and actual is not nil, test will immediately fail
func (u *TestUtil) IsNil(actual interface{}, exit ...bool) bool {
	return u.Equal(nil, actual, exit...)
}

// NotNil asserts that the actual value is not nil
// Returns true if actual is not nil, false otherwise
// If exit is true and actual is nil, test will immediately fail
func (u *TestUtil) NotNil(actual interface{}, exit ...bool) bool {
	return u.Equal(true, actual != nil, exit...)
}

// Parallel marks the test as a parallel test
// It should be called before the test starts
func (u *TestUtil) Parallel() {
	u.t.Helper()
	u.t.(*testing.T).Parallel()
}

// TestCase represents a test case with a name and arbitrary data
type TestCase struct {
    Name string
    Data interface{}
}

// RunTests runs a series of test cases with a test function
func (u *TestUtil) RunTests(tests []TestCase, testFunc func(tt *TestUtil, tc TestCase)) {
    u.t.Helper()
    for _, tc := range tests {
        u.Run(tc.Name, func(tt *TestUtil) {
            testFunc(tt, tc)
        })
    }
}


// ErrorTestCase represents a test case with error expectations
type ErrorTestCase struct {
    Name     string
    Input    interface{}
    Expected interface{}
    WantErr  bool
}

// RunErrorTests runs test cases that test functions returning (result, error)
func (u *TestUtil) RunErrorTests(
    tests []ErrorTestCase,
    testFunc func(input interface{}) (interface{}, error),
) {
    u.t.Helper()
    for _, tc := range tests {
        u.Run(tc.Name, func(tt *TestUtil) {
            result, err := testFunc(tc.Input)

            if tc.WantErr {
                if err == nil {
                    tt.Fatal("Expected error but got none")
                }
                return
            }

            if err != nil {
                tt.Fatal("Unexpected error:", err)
            }

            tt.Equal(tc.Expected, result)
        })
    }
}

