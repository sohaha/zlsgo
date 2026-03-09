package zlsgo

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// fakeReporter records assertion output for internal tests.
type fakeReporter struct {
	failed   bool
	fatal    bool
	messages []string
}

// Helper marks fakeReporter as a helper reporter.
func (f *fakeReporter) Helper() {}

// Errorf records a non-fatal assertion failure.
func (f *fakeReporter) Errorf(format string, args ...interface{}) {
	f.failed = true
	f.messages = append(f.messages, fmt.Sprintf(format, args...))
}

// Fatalf records a fatal assertion failure.
func (f *fakeReporter) Fatalf(format string, args ...interface{}) {
	f.failed = true
	f.fatal = true
	f.messages = append(f.messages, fmt.Sprintf(format, args...))
}

// Log is a no-op for fakeReporter.
func (f *fakeReporter) Log(args ...interface{}) {}

// Logf is a no-op for fakeReporter.
func (f *fakeReporter) Logf(format string, args ...interface{}) {}

// Fatal records a fatal message.
func (f *fakeReporter) Fatal(args ...interface{}) {
	f.failed = true
	f.fatal = true
	f.messages = append(f.messages, fmt.Sprint(args...))
}

// TestAssertionFailures verifies failure output for assertions.
func TestAssertionFailures(t *testing.T) {
	tests := []struct {
		name    string
		run     func(tt *TestUtil) bool
		want    string
		wantHit bool
	}{
		{name: "Equal", run: func(tt *TestUtil) bool { return tt.Equal(1, 2) }, want: "期待:1", wantHit: false},
		{name: "NoEqual", run: func(tt *TestUtil) bool { return tt.NoEqual(1, 1) }, want: "期待不等于:1", wantHit: false},
		{name: "True", run: func(tt *TestUtil) bool { return tt.True(false) }, want: "期待:true", wantHit: false},
		{name: "False", run: func(tt *TestUtil) bool { return tt.False(true) }, want: "期待:false", wantHit: false},
		{name: "NoError", run: func(tt *TestUtil) bool { return tt.NoError(errors.New("boom")) }, want: "Error: boom", wantHit: false},
		{name: "Error", run: func(tt *TestUtil) bool { return tt.Error(nil) }, want: "期待 error 不为 nil", wantHit: false},
		{name: "ErrorContainsNil", run: func(tt *TestUtil) bool { return tt.ErrorContains("boom", nil) }, want: "期待 error 包含", wantHit: false},
		{name: "ErrorContainsValue", run: func(tt *TestUtil) bool { return tt.ErrorContains("want", errors.New("boom")) }, want: "期待 error 包含", wantHit: false},
		{name: "IsNil", run: func(tt *TestUtil) bool { return tt.IsNil(1) }, want: "期待:nil", wantHit: false},
		{name: "NotNil", run: func(tt *TestUtil) bool { return tt.NotNil([]int(nil)) }, want: "期待非 nil", wantHit: false},
		{name: "Contains", run: func(tt *TestUtil) bool { return tt.Contains("xyz", "hello") }, want: "期待包含", wantHit: false},
		{name: "NotContains", run: func(tt *TestUtil) bool { return tt.NotContains("ell", "hello") }, want: "期待不包含", wantHit: false},
		{name: "LenMismatch", run: func(tt *TestUtil) bool { return tt.Len(1, "ab") }, want: "期待长度:1 - 结果:2", wantHit: false},
		{name: "LenUnsupported", run: func(tt *TestUtil) bool { return tt.Len(1, 1) }, want: "无法获取长度", wantHit: false},
		{name: "Panics", run: func(tt *TestUtil) bool { return tt.Panics(func() {}) }, want: "期待 panic", wantHit: false},
		{name: "PanicsNil", run: func(tt *TestUtil) bool { return tt.Panics(nil) }, want: "函数为 nil", wantHit: false},
		{name: "NotPanics", run: func(tt *TestUtil) bool { return tt.NotPanics(func() { panic("boom") }) }, want: "期待不 panic", wantHit: false},
		{name: "NotPanicsNil", run: func(tt *TestUtil) bool { return tt.NotPanics(nil) }, want: "函数为 nil", wantHit: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reporter := &fakeReporter{}
			result := tc.run(newTestUtil(nil, reporter))
			if result != tc.wantHit {
				t.Fatalf("result = %v, want %v", result, tc.wantHit)
			}
			if !reporter.failed || reporter.fatal {
				t.Fatalf("failed = %v, fatal = %v", reporter.failed, reporter.fatal)
			}
			if len(reporter.messages) == 0 || !strings.Contains(reporter.messages[0], tc.want) {
				t.Fatalf("messages = %v, want contains %q", reporter.messages, tc.want)
			}
		})
	}
}

// TestAssertionExitFailures verifies fatal assertion behavior.
func TestAssertionExitFailures(t *testing.T) {
	reporter := &fakeReporter{}
	newTestUtil(nil, reporter).Equal(1, 2, true)

	if !reporter.failed || !reporter.fatal {
		t.Fatalf("failed = %v, fatal = %v", reporter.failed, reporter.fatal)
	}
}

// TestCatchPanic verifies panic capture behavior.
func TestCatchPanic(t *testing.T) {
	panicked, recovered := catchPanic(func() { panic("boom") })
	if !panicked || recovered != "boom" {
		t.Fatalf("panicked = %v, recovered = %v", panicked, recovered)
	}

	panicked, recovered = catchPanic(func() { panic(nil) })
	if !panicked || recovered != nil {
		t.Fatalf("panicked = %v, recovered = %v", panicked, recovered)
	}

	panicked, recovered = catchPanic(func() {})
	if panicked || recovered != nil {
		t.Fatalf("panicked = %v, recovered = %v", panicked, recovered)
	}
}

// TestUnsupportedTestingTMethods verifies *testing.T-only helpers.
func TestUnsupportedTestingTMethods(t *testing.T) {
	tests := []struct {
		name string
		run  func(tt *TestUtil)
		want string
	}{
		{name: "T", run: func(tt *TestUtil) { _ = tt.T() }, want: "T 仅支持 *testing.T"},
		{name: "Run", run: func(tt *TestUtil) { tt.Run("sub", func(tt *TestUtil) {}) }, want: "Run 仅支持 *testing.T"},
		{name: "Parallel", run: func(tt *TestUtil) { tt.Parallel() }, want: "Parallel 仅支持 *testing.T"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reporter := &fakeReporter{}
			tc.run(newTestUtil(nil, reporter))
			if !reporter.failed || !reporter.fatal {
				t.Fatalf("failed = %v, fatal = %v", reporter.failed, reporter.fatal)
			}
			if len(reporter.messages) == 0 || !strings.Contains(reporter.messages[0], tc.want) {
				t.Fatalf("messages = %v, want contains %q", reporter.messages, tc.want)
			}
		})
	}
}

// TestTypedNilComparisonDoesNotRegress verifies typed nil comparison safety.
func TestTypedNilComparisonDoesNotRegress(t *testing.T) {
	reporter := &fakeReporter{}
	tt := newTestUtil(nil, reporter)

	var nilPtr *int
	var nilSlice []int

	if !tt.NoEqual(nilPtr, nilSlice) {
		t.Fatal("typed nil values with different types should not be Equal")
	}
	if tt.Equal(nilPtr, nilSlice) {
		t.Fatal("typed nil values with different types should not compare equal")
	}
	if !reporter.failed || len(reporter.messages) == 0 || !strings.Contains(reporter.messages[0], "期待:") {
		t.Fatalf("messages = %v", reporter.messages)
	}
}
