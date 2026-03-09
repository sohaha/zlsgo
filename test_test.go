package zlsgo_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/sohaha/zlsgo"
)

// TestNewTest covers the basic TestUtil API surface.
func TestNewTest(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Equal(1, 1)
	tt.EqualExit(1, 1)
	tt.True(true)
	tt.False(false)
	tt.EqualTrue(true)
	tt.EqualFalse(false)
	tt.EqualNil(nil)
	tt.NoError(nil, true)
	tt.Error(errors.New("ok"), true)
	tt.ErrorContains("ok", errors.New("ok"), true)
	tt.NoEqual(nil, true)
	tt.Contains("ell", "hello", true)
	tt.NotContains("xyz", "hello", true)
	tt.IsNil(nil, true)
	tt.NotNil(true, true)
	tt.Len(2, []int{1, 2}, true)
	tt.Panics(func() { panic("ok") }, true)
	tt.NotPanics(func() {}, true)
	tt.Log("ok")
	tt.T().Log("ok")
	tt.Run("Logf", func(tt *zlsgo.TestUtil) {
		tt.Parallel()
		tt.Logf("name: %s\n", tt.PrintMyName())
	})
}

// TestGetCallerInfo verifies caller lookup behavior.
func TestGetCallerInfo(t *testing.T) {
	tt := zlsgo.NewTest(t)
	info := tt.GetCallerInfo()

	tt.NotNil(info)
	if info != "<Unable to get information>" {
		tt.EqualTrue(strings.HasPrefix(info, "test_test.go:"))
	}
}

// TestEqual verifies equality assertions.
func TestEqual(t *testing.T) {
	tt := zlsgo.NewTest(t)
	var nilSlice []int

	tt.EqualTrue(tt.Equal(1, 1))
	tt.EqualTrue(tt.Equal("hello", "hello"))
	tt.EqualTrue(tt.Equal([]int{1, 2, 3}, []int{1, 2, 3}))
	tt.EqualTrue(tt.NoEqual(nil, nilSlice))
}

// TestNoEqual verifies inequality assertions.
func TestNoEqual(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.EqualTrue(tt.NoEqual(1, 2))
	tt.EqualTrue(tt.NoEqual("hello", "world"))
}

// TestNilAssertions verifies nil-related assertions.
func TestNilAssertions(t *testing.T) {
	tt := zlsgo.NewTest(t)

	var nilPtr *int
	var nilSlice []int
	var nilMap map[string]int
	var nilFunc func()
	var nilChan chan int

	tt.EqualNil(nilPtr)
	tt.EqualNil(nilSlice)
	tt.EqualNil(nilMap)
	tt.EqualNil(nilFunc)
	tt.EqualNil(nilChan)
	tt.EqualTrue(tt.IsNil(nilSlice))
	tt.EqualTrue(tt.NotNil(123))
}

// TestNoErrorAndError verifies error assertions.
func TestNoErrorAndError(t *testing.T) {
	tt := zlsgo.NewTest(t)
	err := errors.New("test error")

	tt.EqualTrue(tt.NoError(nil))
	tt.EqualTrue(tt.Error(err))
	tt.EqualTrue(tt.ErrorContains("test", err))
}

// TestLen verifies length assertions.
func TestLen(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.EqualTrue(tt.Len(5, "hello"))
	tt.EqualTrue(tt.Len(3, []int{1, 2, 3}))
	tt.EqualTrue(tt.Len(2, map[string]int{"a": 1, "b": 2}))
}

// TestBoolAssertions verifies boolean assertions.
func TestBoolAssertions(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.EqualTrue(tt.True(true))
	tt.EqualTrue(tt.False(false))
}

// TestContainsAssertions verifies string containment assertions.
func TestContainsAssertions(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.EqualTrue(tt.Contains("ell", "hello"))
	tt.EqualTrue(tt.NotContains("xyz", "hello"))
}

// TestPanicAssertions verifies panic assertions.
func TestPanicAssertions(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.EqualTrue(tt.Panics(func() { panic("boom") }))
	tt.EqualTrue(tt.NotPanics(func() {}))
}

// TestLogMethods verifies logging helpers.
func TestLogMethods(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.Log("This is a test log")
	tt.Logf("This is a formatted log: %s", "test")
}

// TestRun verifies subtest execution.
func TestRun(t *testing.T) {
	tt := zlsgo.NewTest(t)
	executed := false

	tt.Run("SubTest", func(subTt *zlsgo.TestUtil) {
		executed = true
		subTt.Equal(1, 1)
	})

	if !executed {
		t.Fatal("SubTest was not executed")
	}
}

// TestTMethod verifies access to the underlying testing.T.
func TestTMethod(t *testing.T) {
	tt := zlsgo.NewTest(t)
	underlyingT := tt.T()

	if underlyingT == nil {
		t.Fatal("T() should return the underlying *testing.T")
	}
	if underlyingT != t {
		t.Fatal("T() should return the same *testing.T instance")
	}
}

// TestRunTests verifies batch test execution helpers.
func TestRunTests(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tests := []zlsgo.TestCase{
		{Name: "Test1", Data: 1},
		{Name: "Test2", Data: 2},
		{Name: "Test3", Data: 3},
	}

	executedCount := 0
	tt.RunTests(tests, func(subTt *zlsgo.TestUtil, tc zlsgo.TestCase) {
		executedCount++
		subTt.NotNil(tc.Name)
		subTt.NotNil(tc.Data)
	})

	tt.Equal(3, executedCount)
}

// TestRunErrorTests verifies error-oriented batch execution.
func TestRunErrorTests(t *testing.T) {
	tt := zlsgo.NewTest(t)
	testFunc := func(input interface{}) (interface{}, error) {
		v := input.(int)
		if v < 0 {
			return "", errors.New("negative number")
		}
		return "positive", nil
	}

	tests := []zlsgo.ErrorTestCase{
		{Name: "PositiveNumber", Input: 5, Expected: "positive"},
		{Name: "NegativeNumber", Input: -1, WantErr: true},
	}

	tt.RunErrorTests(tests, testFunc)
}

// TestNewAPIStyle verifies mixed helper usage in subtests.
func TestNewAPIStyle(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tests := []zlsgo.TestCase{
		{Name: "NewAPITest1", Data: "data1"},
		{Name: "NewAPITest2", Data: "data2"},
	}

	executedCount := 0
	tt.RunTests(tests, func(subTt *zlsgo.TestUtil, tc zlsgo.TestCase) {
		executedCount++
		subTt.NotNil(tc.Data)
		s := tc.Data.(string)
		subTt.EqualTrue(len(s) > 0)
		subTt.Len(len(s), s)
	})

	tt.Equal(2, executedCount)

	testFunc := func(input interface{}) (interface{}, error) {
		s := input.(string)
		if s == "error" {
			return 0, errors.New("test error")
		}
		return len(s), nil
	}

	errorTests := []zlsgo.ErrorTestCase{
		{Name: "ValidString", Input: "hello", Expected: 5},
		{Name: "ErrorString", Input: "error", WantErr: true},
	}

	tt.RunErrorTests(errorTests, testFunc)
}
