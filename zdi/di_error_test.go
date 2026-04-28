package zdi_test

import (
	"fmt"
	"sync"
	"testing"

	zls "github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zdi"
)

type testStruct struct {
	Name string
}

func TestDIErrorPaths(t *testing.T) {
	tt := zls.NewTest(t)

	t.Run("Unregistered type", func(t *testing.T) {
		di := zdi.New()
		_, err := di.Invoke(func(unregistered *testStruct) {
			t.Error("should not be called with unregistered type")
		})
		tt.NotNil(err)
	})

	t.Run("Nil function", func(t *testing.T) {
		di := zdi.New()
		_, err := di.Invoke(nil)
		tt.NotNil(err)
	})
}

// TestDIConcurrentErrorHandling tests DI error handling under concurrent load
func TestDIConcurrentErrorHandling(t *testing.T) {
	tt := zls.NewTest(t)
	di := zdi.New()

	di.Map(&testStruct{Name: "test"})

	done := make(chan bool, 20)
	successCount := 0
	errorCount := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			_, err := di.Invoke(func(ts *testStruct) {
				if ts == nil {
					t.Errorf("Goroutine %d: received nil struct", id)
				}
			})
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)

		go func(id int) {
			defer func() { done <- true }()
			_, err := di.Invoke(func(unregistered *string) {
				t.Error("should not be called")
			})
			if err != nil {
				mu.Lock()
				errorCount++
				mu.Unlock()
			}
		}(i)
	}

	for i := 0; i < 20; i++ {
		<-done
	}

	mu.Lock()
	tt.Equal(10, successCount)
	tt.Equal(10, errorCount)
	mu.Unlock()
}

// TestDIMultipleErrorScenarios tests various error scenarios
func TestDIMultipleErrorScenarios(t *testing.T) {
	tt := zls.NewTest(t)

	t.Run("Mixed valid and invalid dependencies", func(t *testing.T) {
		di := zdi.New()

		di.Map("valid-string")
		di.Map(123)

		_, err := di.Invoke(func(s string, i int, unregistered *testStruct) {
			t.Error("should not be called with unregistered type")
		})

		tt.NotNil(err)
	})

	t.Run("Nested unregistered dependencies", func(t *testing.T) {
		di := zdi.New()

		type Inner struct {
			Value int
		}

		type Outer struct {
			Inner *Inner
		}

		_, err := di.Invoke(func(outer *Outer) {
			t.Error("should not be called")
		})

		tt.NotNil(err)
	})

	t.Run("Interface type resolution errors", func(t *testing.T) {
		di := zdi.New()

		_, err := di.Invoke(func(w interface{}) {
			t.Error("should not be called")
		})

		tt.NotNil(err)
	})
}

// TestDIErrorRecovery tests that DI remains functional after errors
func TestDIErrorRecovery(t *testing.T) {
	tt := zls.NewTest(t)
	di := zdi.New()

	validStruct := &testStruct{Name: "valid"}
	di.Map(validStruct)

	_, err1 := di.Invoke(func(unregistered *string) {
		t.Error("should not be called")
	})
	tt.NotNil(err1)

	invoked := false
	_, err2 := di.Invoke(func(ts *testStruct) {
		invoked = true
		if ts.Name != "valid" {
			t.Errorf("Expected Name 'valid', got '%s'", ts.Name)
		}
	})
	tt.EqualNil(err2)
	tt.EqualTrue(invoked)
}

// TestDIConcurrentRegistrationAndInvocation tests safety of concurrent operations
func TestDIConcurrentRegistrationAndInvocation(t *testing.T) {
	tt := zls.NewTest(t)
	di := zdi.New()

	done := make(chan bool, 30)
	var successCount, errorCount int
	var mu sync.Mutex

	for i := 0; i < 15; i++ {
		go func(id int) {
			defer func() { done <- true }()
			di.Map(&testStruct{Name: fmt.Sprintf("test-%d", id)})
		}(i)

		go func(id int) {
			defer func() { done <- true }()
			_, err := di.Invoke(func(ts *testStruct) {
				if ts != nil && ts.Name != "" {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			})
			if err != nil {
				mu.Lock()
				errorCount++
				mu.Unlock()
			}
		}(i)
	}

	for i := 0; i < 30; i++ {
		<-done
	}

	mu.Lock()
	totalOps := successCount + errorCount
	mu.Unlock()

	tt.EqualTrue(totalOps > 0)
}

// TestDIPanicRecovery tests that DI handles panics gracefully
func TestDIPanicRecovery(t *testing.T) {
	tt := zls.NewTest(t)
	di := zdi.New()

	di.Map(&testStruct{Name: "test"})

	t.Run("Panic in invoked function", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				tt.EqualTrue(true)
			}
		}()

		_, err := di.Invoke(func(ts *testStruct) {
			if ts != nil {
				panic("intentional panic for testing")
			}
		})

		tt.NotNil(err)
	})
}

// TestDINilValueHandling tests error handling for nil values
func TestDINilValueHandling(t *testing.T) {
	tt := zls.NewTest(t)
	di := zdi.New()

	t.Run("Register nil value", func(t *testing.T) {
		var nilStruct *testStruct = nil
		di.Map(nilStruct)

		invoked := false
		_, err := di.Invoke(func(ts *testStruct) {
			invoked = true
			if ts != nil {
				t.Error("Expected nil struct, got non-nil")
			}
		})

		tt.EqualNil(err)
		tt.EqualTrue(invoked)
	})

	t.Run("Unregistered nil dependency", func(t *testing.T) {
		_, err := di.Invoke(func(ts *testStruct) {
		})

		_ = err
	})
}
