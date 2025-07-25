package ztype

import (
	"testing"
)

// TestPathParsingCorrectness tests path parsing correctness
func TestPathParsingCorrectness(t *testing.T) {
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
		path      string
		expected  interface{}
		exists    bool
		skipValue bool
	}{
		{"user.profile.name", "John", true, false},
		{"user.profile.age", 30, true, false},
		{"items.0.id", 1, true, false},
		{"items.1.name", "item2", true, false},
		{"simple", "value", true, false},
		{"user.nonexistent", nil, false, false},
		{"items.5.id", nil, false, false},
		{"", testData, true, true},
	}

	for _, test := range tests {
		result, exists := parsePath(test.path, testData)
		if exists != test.exists {
			t.Errorf("Path %s: expected exists %v, got %v", test.path, test.exists, exists)
		}
		if exists && !test.skipValue && result != test.expected {
			t.Errorf("Path %s: expected value %v, got %v", test.path, test.expected, result)
		}
	}
}

// TestPathParsingWithEscapeChars tests escape character handling
func TestPathParsingWithEscapeChars(t *testing.T) {
	testData := map[string]interface{}{
		"user.name":  "John",
		"user.email": "john@example.com",
		"data": map[string]interface{}{
			"key.with.dots": "value",
		},
	}

	tests := []struct {
		path     string
		expected interface{}
		exists   bool
	}{
		{"user\\.name", "John", true},
		{"user\\.email", "john@example.com", true},
		{"data.key\\.with\\.dots", "value", true},
	}

	for _, test := range tests {
		result, exists := parsePath(test.path, testData)
		t.Logf("Path: %s, Result: %v, Exists: %v", test.path, result, exists)
		if exists != test.exists {
			t.Errorf("Path %s: expected exists %v, got %v", test.path, test.exists, exists)
		}
		if exists && result != test.expected {
			t.Errorf("Path %s: expected value %v, got %v", test.path, test.expected, result)
		}
	}
}
