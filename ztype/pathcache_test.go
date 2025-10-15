package ztype

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

// TestPathParsingCorrectness tests path parsing correctness
func TestPathParsingCorrectness(t *testing.T) {
	tt := zlsgo.NewTest(t)
	testData := map[string]interface{}{
		"user": map[string]interface{}{
			"profile": map[string]interface{}{
				"name": "John",
				"age":  30,
			},
		},
		"items": []interface{}{
			map[string]interface{}{"id": 1, "name": "item1"},
			map[string]interface{}{"id": 2, "name": "item2"},
		},
		"simple": "value",
	}

	tests := []struct {
		expected  interface{}
		path      string
		exists    bool
		skipValue bool
	}{
		{path: "user.profile.name", expected: "John", exists: true, skipValue: false},
		{path: "user.profile.age", expected: 30, exists: true, skipValue: false},
		{path: "items.0.id", expected: 1, exists: true, skipValue: false},
		{path: "items.1.name", expected: "item2", exists: true, skipValue: false},
		{path: "simple", expected: "value", exists: true, skipValue: false},
		{path: "user.nonexistent", expected: nil, exists: false, skipValue: false},
		{path: "items.5.id", expected: nil, exists: false, skipValue: false},
		{path: "", expected: testData, exists: true, skipValue: true},
	}

	for _, test := range tests {
		result, exists := parsePath(test.path, testData)
		tt.Equal(test.exists, exists)
		if exists && !test.skipValue {
			tt.Equal(test.expected, result)
		}
	}
}

// TestPathParsingWithEscapeChars tests escape character handling
func TestPathParsingWithEscapeChars(t *testing.T) {
	tt := zlsgo.NewTest(t)
	testData := map[string]interface{}{
		"user.name":  "John",
		"user.email": "john@example.com",
		"data": map[string]interface{}{
			"key.with.dots": "value",
		},
	}

	tests := []struct {
		expected interface{}
		path     string
		exists   bool
	}{
		{path: "user\\.name", expected: "John", exists: true},
		{path: "user\\.email", expected: "john@example.com", exists: true},
		{path: "data.key\\.with\\.dots", expected: "value", exists: true},
	}

	for _, test := range tests {
		result, exists := parsePath(test.path, testData)
		tt.Logf("Path: %s, Result: %v, Exists: %v", test.path, result, exists)
		tt.Equal(test.exists, exists)
		if exists {
			tt.Equal(test.expected, result)
		}
	}
}
