package zlsgo

import (
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const unableCallerInfo = "<Unable to get information>"

// testReporter defines the minimal reporting methods used by TestUtil.
type testReporter interface {
	Helper()
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Fatal(args ...interface{})
}

// TestUtil wraps testing.TB and provides lightweight assertions, error checks,
// logging helpers, and subtest helpers.
//
// Use NewTest to create a helper in unit tests:
//
//	tt := zlsgo.NewTest(t)
//	tt.Equal(1, 1)
//	tt.NoError(err)
//	tt.Len(3, []int{1, 2, 3})
//	tt.Run("sub", func(tt *zlsgo.TestUtil) {
//		tt.EqualTrue(true)
//	})
//
// Assertions cover equality, nil and error checks, string containment,
// length checks, and panic checks.
// Passing true in exit makes an assertion fail the current test immediately.
// Methods T, Run, and Parallel require the underlying value to be *testing.T.
type TestUtil struct {
	tb       testing.TB
	reporter testReporter
}

// NewTest creates a TestUtil from testing.TB.
//
// The returned helper uses the testing logger for assertion output and is
// intended for test assertions plus lightweight subtest orchestration.
// Methods T, Run, and Parallel require the underlying value to be *testing.T.
func NewTest(t testing.TB) *TestUtil {
	return newTestUtil(t, t)
}

// newTestUtil builds a TestUtil with a custom reporter.
func newTestUtil(tb testing.TB, reporter testReporter) *TestUtil {
	return &TestUtil{tb: tb, reporter: reporter}
}

// GetCallerInfo returns the file name and line number of the test caller.
func (u *TestUtil) GetCallerInfo() string {
	for depth := 1; depth < 20; depth++ {
		_, file, line, ok := runtime.Caller(depth)
		if !ok {
			break
		}
		if !strings.HasSuffix(file, "_test.go") {
			continue
		}
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	return unableCallerInfo
}

// Equal compares expected and actual values with reflect.DeepEqual.
// Passing true in exit will stop the current test immediately on failure.
func (u *TestUtil) Equal(expected, actual interface{}, exit ...bool) bool {
	if valuesEqual(expected, actual) {
		return true
	}
	return u.failAssertion(exit,
		"%s 期待:%v (type %s) - 结果:%v (type %s)",
		u.PrintMyName(), expected, valueType(expected), actual, valueType(actual),
	)
}

// NoEqual compares expected and actual values and asserts they are not equal.
func (u *TestUtil) NoEqual(expected, actual interface{}, exit ...bool) bool {
	if !valuesEqual(expected, actual) {
		return true
	}
	return u.failAssertion(exit,
		"%s 期待不等于:%v (type %s) - 结果:%v (type %s)",
		u.PrintMyName(), expected, valueType(expected), actual, valueType(actual),
	)
}

// EqualTrue asserts that the actual value is true.
func (u *TestUtil) EqualTrue(actual interface{}, exit ...bool) {
	u.Equal(true, actual, exit...)
}

// EqualFalse asserts that the actual value is false.
func (u *TestUtil) EqualFalse(actual interface{}, exit ...bool) {
	u.Equal(false, actual, exit...)
}

// True asserts that the actual boolean value is true.
func (u *TestUtil) True(actual bool, exit ...bool) bool {
	if actual {
		return true
	}
	return u.failAssertion(exit, "%s 期待:true - 结果:false", u.PrintMyName())
}

// False asserts that the actual boolean value is false.
func (u *TestUtil) False(actual bool, exit ...bool) bool {
	if !actual {
		return true
	}
	return u.failAssertion(exit, "%s 期待:false - 结果:true", u.PrintMyName())
}

// EqualNil asserts that the actual value is nil.
func (u *TestUtil) EqualNil(actual interface{}, exit ...bool) {
	u.IsNil(actual, exit...)
}

// NoError asserts that err is nil and reports failures through testing output.
func (u *TestUtil) NoError(err error, exit ...bool) bool {
	if err == nil {
		return true
	}
	return u.failAssertion(exit, "%s Error: %v", u.PrintMyName(), err)
}

// Error asserts that err is not nil.
func (u *TestUtil) Error(err error, exit ...bool) bool {
	if err != nil {
		return true
	}
	return u.failAssertion(exit, "%s 期待 error 不为 nil", u.PrintMyName())
}

// ErrorContains asserts that err is not nil and its message contains expected.
func (u *TestUtil) ErrorContains(expected string, err error, exit ...bool) bool {
	if err == nil {
		return u.failAssertion(exit, "%s 期待 error 包含:%q - 结果:nil", u.PrintMyName(), expected)
	}
	if strings.Contains(err.Error(), expected) {
		return true
	}
	return u.failAssertion(exit,
		"%s 期待 error 包含:%q - 结果:%q",
		u.PrintMyName(), expected, err.Error(),
	)
}

// EqualExit compares expected and actual values and immediately fails the test if not equal.
func (u *TestUtil) EqualExit(expected, actual interface{}) {
	u.Equal(expected, actual, true)
}

// Contains asserts that actual contains expected.
func (u *TestUtil) Contains(expected, actual string, exit ...bool) bool {
	if strings.Contains(actual, expected) {
		return true
	}
	return u.failAssertion(exit,
		"%s 期待包含:%q - 结果:%q",
		u.PrintMyName(), expected, actual,
	)
}

// NotContains asserts that actual does not contain expected.
func (u *TestUtil) NotContains(expected, actual string, exit ...bool) bool {
	if !strings.Contains(actual, expected) {
		return true
	}
	return u.failAssertion(exit,
		"%s 期待不包含:%q - 结果:%q",
		u.PrintMyName(), expected, actual,
	)
}

// Len asserts that the actual value length matches the expected length.
func (u *TestUtil) Len(expected int, actual interface{}, exit ...bool) bool {
	length, ok := lengthOf(actual)
	if !ok {
		return u.failAssertion(exit, "%s 无法获取长度 (type %s)", u.PrintMyName(), valueType(actual))
	}
	if length == expected {
		return true
	}
	return u.failAssertion(exit,
		"%s 期待长度:%d - 结果:%d (type %s)",
		u.PrintMyName(), expected, length, valueType(actual),
	)
}

// Panics asserts that fn panics.
func (u *TestUtil) Panics(fn func(), exit ...bool) bool {
	if fn == nil {
		return u.failAssertion(exit, "%s 期待 panic - 结果:函数为 nil", u.PrintMyName())
	}
	panicked, recovered := catchPanic(fn)
	if panicked {
		return true
	}
	return u.failAssertion(exit, "%s 期待 panic - 结果:未发生 panic (%v)", u.PrintMyName(), recovered)
}

// NotPanics asserts that fn does not panic.
func (u *TestUtil) NotPanics(fn func(), exit ...bool) bool {
	if fn == nil {
		return u.failAssertion(exit, "%s 期待不 panic - 结果:函数为 nil", u.PrintMyName())
	}
	panicked, recovered := catchPanic(fn)
	if !panicked {
		return true
	}
	return u.failAssertion(exit, "%s 期待不 panic - 结果:%v", u.PrintMyName(), recovered)
}

// Log logs the given values to the test output.
func (u *TestUtil) Log(v ...interface{}) {
	u.reporter.Helper()
	u.reporter.Log(v...)
}

// Logf logs the formatted string to the test output.
func (u *TestUtil) Logf(format string, args ...interface{}) {
	u.reporter.Helper()
	u.reporter.Logf(format, args...)
}

// Fatal logs the given values to the test output and immediately fails the test.
func (u *TestUtil) Fatal(v ...interface{}) {
	u.reporter.Helper()
	u.reporter.Fatal(v...)
}

// PrintMyName returns the caller information for the current test.
func (u *TestUtil) PrintMyName() string {
	return u.GetCallerInfo()
}

// Run runs a subtest with the given name and function.
// It is only available when the underlying testing object is *testing.T.
func (u *TestUtil) Run(name string, f func(tt *TestUtil)) {
	u.reporter.Helper()
	t := u.requireTestingT("Run")
	if t == nil {
		return
	}
	t.Run(name, func(t *testing.T) {
		f(NewTest(t))
	})
}

// T returns the underlying *testing.T object.
// It fails immediately if TestUtil was not created from *testing.T.
func (u *TestUtil) T() *testing.T {
	return u.requireTestingT("T")
}

// IsNil asserts that the actual value is nil.
func (u *TestUtil) IsNil(actual interface{}, exit ...bool) bool {
	if isNilValue(actual) {
		return true
	}
	return u.failAssertion(exit,
		"%s 期待:nil - 结果:%v (type %s)",
		u.PrintMyName(), actual, valueType(actual),
	)
}

// NotNil asserts that the actual value is not nil.
func (u *TestUtil) NotNil(actual interface{}, exit ...bool) bool {
	if !isNilValue(actual) {
		return true
	}
	return u.failAssertion(exit, "%s 期待非 nil (type %s)", u.PrintMyName(), valueType(actual))
}

// Parallel marks the test as a parallel test.
// It is only available when the underlying testing object is *testing.T.
func (u *TestUtil) Parallel() {
	u.reporter.Helper()
	t := u.requireTestingT("Parallel")
	if t == nil {
		return
	}
	t.Parallel()
}

// TestCase represents a test case with a name and arbitrary data.
type TestCase struct {
	Data interface{}
	Name string
}

// RunTests runs a series of named test cases through subtests.
func (u *TestUtil) RunTests(tests []TestCase, testFunc func(tt *TestUtil, tc TestCase)) {
	u.reporter.Helper()
	for _, tc := range tests {
		tc := tc
		u.Run(tc.Name, func(tt *TestUtil) {
			testFunc(tt, tc)
		})
	}
}

// ErrorTestCase represents a test case with error expectations.
type ErrorTestCase struct {
	Input    interface{}
	Expected interface{}
	Name     string
	WantErr  bool
}

// RunErrorTests runs test cases for functions returning (result, error).
func (u *TestUtil) RunErrorTests(
	tests []ErrorTestCase,
	testFunc func(input interface{}) (interface{}, error),
) {
	u.reporter.Helper()
	for _, tc := range tests {
		tc := tc
		u.Run(tc.Name, func(tt *TestUtil) {
			result, err := testFunc(tc.Input)
			if tc.WantErr {
				tt.Error(err, true)
				return
			}
			if !tt.NoError(err, true) {
				return
			}
			tt.Equal(tc.Expected, result)
		})
	}
}

// failAssertion reports an assertion failure and respects exit behavior.
func (u *TestUtil) failAssertion(exit []bool, format string, args ...interface{}) bool {
	u.reporter.Helper()
	if shouldExit(exit) {
		u.reporter.Fatalf(format, args...)
		return false
	}
	u.reporter.Errorf(format, args...)
	return false
}

// requireTestingT returns the underlying *testing.T when available.
func (u *TestUtil) requireTestingT(method string) *testing.T {
	t, ok := u.tb.(*testing.T)
	if ok {
		return t
	}
	u.reporter.Helper()
	u.reporter.Fatalf("%s 仅支持 *testing.T，当前类型:%T", method, u.tb)
	return nil
}

// shouldExit reports whether the assertion should stop the test.
func shouldExit(exit []bool) bool {
	return len(exit) > 0 && exit[0]
}

// valueType returns the reflected type name of v.
func valueType(v interface{}) string {
	t := reflect.TypeOf(v)
	if t == nil {
		return "<nil>"
	}
	return t.String()
}

// valuesEqual compares two values with reflect.DeepEqual.
func valuesEqual(expected, actual interface{}) bool {
	return reflect.DeepEqual(expected, actual)
}

// isNilValue reports whether v is nil or a typed nil value.
func isNilValue(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return rv.IsNil()
	default:
		return false
	}
}

// lengthOf returns the length of supported collection values.
func lengthOf(v interface{}) (int, bool) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return 0, false
	}
	switch rv.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return rv.Len(), true
	default:
		return 0, false
	}
}

// catchPanic runs fn and captures any panic value.
func catchPanic(fn func()) (panicked bool, recovered interface{}) {
	panicked = true

	defer func() {
		recovered = recover()
	}()

	fn()
	panicked = false
	return
}
