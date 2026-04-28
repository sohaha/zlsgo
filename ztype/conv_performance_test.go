package ztype

import (
	"reflect"
	"testing"
	"time"
)

// Benchmark for nested struct field collection
func BenchmarkCollectStructFields_Simple(b *testing.B) {
	type SimpleStruct struct {
		Field1 string
		Field2 int
		Field3 bool
	}

	val := reflect.ValueOf(SimpleStruct{})
	conv := &Conver{Squash: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = conv.collectStructFields(val)
	}
}

func BenchmarkCollectStructFields_Nested(b *testing.B) {
	type Inner struct {
		Inner1 string
		Inner2 int
	}

	type Middle struct {
		Middle1 string
		Inner   Inner `z:"squash"`
	}

	type Outer struct {
		Outer1 string
		Middle Middle `z:"squash"`
	}

	val := reflect.ValueOf(Outer{})
	conv := &Conver{Squash: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = conv.collectStructFields(val)
	}
}

func BenchmarkCollectStructFields_DeepNested(b *testing.B) {
	type Level5 struct {
		F1, F2, F3 string
	}
	type Level4 struct {
		L4 Level5 `z:"squash"`
	}
	type Level3 struct {
		L3 Level4 `z:"squash"`
	}
	type Level2 struct {
		L2 Level3 `z:"squash"`
	}
	type Level1 struct {
		L1 Level2 `z:"squash"`
	}

	val := reflect.ValueOf(Level1{})
	conv := &Conver{Squash: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = conv.collectStructFields(val)
	}
}

func BenchmarkCollectStructFields_Wide(b *testing.B) {
	type WideStruct struct {
		F1, F2, F3, F4, F5, F6, F7, F8, F9, F10          string
		F11, F12, F13, F14, F15, F16, F17, F18, F19, F20 int
	}

	val := reflect.ValueOf(WideStruct{})
	conv := &Conver{Squash: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = conv.collectStructFields(val)
	}
}

// Test correctness after optimization
func TestCollectStructFields_Correctness(t *testing.T) {
	type Inner struct {
		InnerField string `z:"in"`
	}

	type Middle struct {
		MiddleField string `z:"mid"`
		Inner       `z:"squash"`
	}

	type TestStruct struct {
		TopField string `z:"top"`
		Middle   `z:"squash"`
	}

	val := reflect.ValueOf(TestStruct{
		TopField: "top",
		Middle: Middle{
			MiddleField: "middle",
			Inner: Inner{
				InnerField: "inner",
			},
		},
	})

	conv := &Conver{Squash: true}
	fields, remain, err := conv.collectStructFields(val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if remain != nil {
		t.Error("expected nil remain field")
	}

	// Should have 3 fields: TopField, MiddleField, InnerField
	if len(fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(fields))
	}

	// Check field names
	fieldNames := make(map[string]bool)
	for _, field := range fields {
		name := field.field.Name
		fieldNames[name] = true
	}

	expectedFields := []string{"TopField", "MiddleField", "InnerField"}
	for _, expected := range expectedFields {
		if !fieldNames[expected] {
			t.Errorf("missing expected field: %s", expected)
		}
	}
}

// Test that optimization doesn't break remain functionality
func TestCollectStructFields_Remain(t *testing.T) {
	type TestStruct struct {
		Field1 string `z:"f1"`
		Field2 string `z:"f2"`
	}

	val := reflect.ValueOf(TestStruct{
		Field1: "value1",
		Field2: "value2",
	})

	conv := &Conver{Squash: false}
	fields, remain, err := conv.collectStructFields(val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(fields))
	}

	if remain != nil {
		t.Error("expected nil remain field for simple struct")
	}
}

// Stress test with complex nested structure
func TestCollectStructFields_Complex(t *testing.T) {
	type Address struct {
		Street  string `z:"street"`
		City    string `z:"city"`
		Country string `z:"country"`
	}

	type Person struct {
		Name    string `z:"name"`
		Age     int    `z:"age"`
		Address `z:"squash"`
	}

	type Employee struct {
		ID         string `z:"id"`
		Department string `z:"dept"`
		Person     `z:"squash"`
	}

	val := reflect.ValueOf(Employee{
		ID:         "123",
		Department: "Engineering",
		Person: Person{
			Name: "John Doe",
			Age:  30,
			Address: Address{
				Street:  "123 Main St",
				City:    "San Francisco",
				Country: "USA",
			},
		},
	})

	conv := &Conver{Squash: true}
	fields, remain, err := conv.collectStructFields(val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if remain != nil {
		t.Error("expected nil remain field")
	}

	if len(fields) != 7 {
		t.Errorf("expected 7 fields, got %d", len(fields))
	}

	expectedFields := []string{"ID", "Department", "Name", "Age", "Street", "City", "Country"}
	fieldNames := make(map[string]bool)
	for _, field := range fields {
		fieldNames[field.field.Name] = true
	}

	for _, expected := range expectedFields {
		if !fieldNames[expected] {
			t.Errorf("missing expected field: %s", expected)
		}
	}
}

// Performance comparison test (not a benchmark, just timing)
func TestCollectStructFields_Performance(t *testing.T) {
	type Inner struct {
		F1, F2, F3, F4, F5 string
	}
	type Middle struct {
		M1, M2, M3, M4, M5 string
		Inner              Inner `z:"squash"`
	}
	type Outer struct {
		O1, O2, O3, O4, O5 string
		Middle             Middle `z:"squash"`
	}

	val := reflect.ValueOf(Outer{})
	conv := &Conver{Squash: true}

	for i := 0; i < 100; i++ {
		_, _, _ = conv.collectStructFields(val)
	}

	start := time.Now()
	iterations := 10000
	for i := 0; i < iterations; i++ {
		_, _, _ = conv.collectStructFields(val)
	}
	elapsed := time.Since(start)

	avgPerOp := elapsed / time.Duration(iterations)
	t.Logf("Average time per operation: %v", avgPerOp)

	if avgPerOp > time.Millisecond {
		t.Errorf("Performance regression: average operation took %v (expected < 1ms)", avgPerOp)
	}
}
