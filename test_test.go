package zlsgo_test

import (
	"errors"
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestNewTest(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Equal(1, 1)
	tt.EqualExit(1, 1)
	tt.EqualTrue(true)
	tt.EqualFalse(false)
	tt.EqualNil(nil)
	tt.NoError(nil, true)
	tt.NoEqual(nil, true)
	tt.IsNil(nil, true)
	tt.NotNil(true, true)
	tt.Log("ok")
	tt.T().Log("ok")
	tt.Run("Logf", func(tt *zlsgo.TestUtil) {
		tt.Parallel()
		tt.Logf("name: %s\n", tt.PrintMyName())
	})
}

func TestGetCallerInfo(t *testing.T) {
	tt := zlsgo.NewTest(t)

	info := tt.GetCallerInfo()
	tt.NotNil(info)

	// 检查是否包含文件名和行号信息
	if info != "<Unable to get information>" {
		tt.EqualTrue(info != "")
	}
}

func TestEqual(t *testing.T) {
	tt := zlsgo.NewTest(t)

	result := tt.Equal(1, 1)
	tt.EqualTrue(result)

	result = tt.Equal("hello", "hello")
	tt.EqualTrue(result)

	result = tt.Equal([]int{1, 2, 3}, []int{1, 2, 3})
	tt.EqualTrue(result)

	mockT := &testing.T{}
	mockTt := zlsgo.NewTest(mockT)

	result = mockTt.Equal(1, 2)
	tt.EqualFalse(result)

	result = mockTt.Equal("hello", "world")
	tt.EqualFalse(result)
}

func TestNoEqual(t *testing.T) {
	tt := zlsgo.NewTest(t)

	result := tt.NoEqual(1, 2)
	tt.EqualTrue(result)

	result = tt.NoEqual("hello", "world")
	tt.EqualTrue(result)

	mockT := &testing.T{}
	mockTt := zlsgo.NewTest(mockT)

	result = mockTt.NoEqual(1, 1)
	tt.EqualFalse(result)
}

func TestEqualTrue(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.EqualTrue(true)
}

func TestEqualFalse(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.EqualFalse(false)
}

func TestEqualNil(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.EqualNil(nil)

	var nilPtr *int
	tt.Equal((*int)(nil), nilPtr)
}

func TestNoError(t *testing.T) {
	tt := zlsgo.NewTest(t)

	result := tt.NoError(nil)
	tt.EqualTrue(result)

	mockT := &testing.T{}
	mockTt := zlsgo.NewTest(mockT)

	result = mockTt.NoError(errors.New("test error"))
	tt.EqualFalse(result)
}

func TestIsNil(t *testing.T) {
	tt := zlsgo.NewTest(t)

	result := tt.IsNil(nil)
	tt.EqualTrue(result)

	mockT := &testing.T{}
	mockTt := zlsgo.NewTest(mockT)

	var nilPtr *int
	mockTt.IsNil(nilPtr)

	result = mockTt.IsNil(123)
	tt.EqualFalse(result)
}

func TestNotNil(t *testing.T) {
	tt := zlsgo.NewTest(t)

	result := tt.NotNil(123)
	tt.EqualTrue(result)

	result = tt.NotNil("hello")
	tt.EqualTrue(result)

	mockT := &testing.T{}
	mockTt := zlsgo.NewTest(mockT)

	result = mockTt.NotNil(nil)
	tt.EqualFalse(result)
}

func TestLogMethods(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Log("This is a test log")
	tt.Logf("This is a formatted log: %s", "test")
}

func TestRun(t *testing.T) {
	tt := zlsgo.NewTest(t)

	executed := false
	tt.Run("SubTest", func(subTt *zlsgo.TestUtil) {
		executed = true
		subTt.Equal(1, 1)
	})

	if !executed {
		t.Error("SubTest was not executed")
	}
}

func TestTMethod(t *testing.T) {
	tt := zlsgo.NewTest(t)

	underlyingT := tt.T()
	if underlyingT == nil {
		t.Error("T() should return the underlying *testing.T")
	}

	if underlyingT != t {
		t.Error("T() should return the same *testing.T instance")
	}
}

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
		{
			Name:     "PositiveNumber",
			Input:    5,
			Expected: "positive",
			WantErr:  false,
		},
		{
			Name:     "NegativeNumber",
			Input:    -1,
			Expected: "",
			WantErr:  true,
		},
	}

	tt.RunErrorTests(tests, testFunc)
}

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
		{
			Name:     "ValidString",
			Input:    "hello",
			Expected: 5,
			WantErr:  false,
		},
		{
			Name:     "ErrorString",
			Input:    "error",
			Expected: 0,
			WantErr:  true,
		},
	}

	tt.RunErrorTests(errorTests, testFunc)
}
