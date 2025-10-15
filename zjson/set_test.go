package zjson

import (
	"strconv"
	"testing"

	"github.com/sohaha/zlsgo/zstring"

	"github.com/sohaha/zlsgo"
)

func TestSet(t *testing.T) {
	demoData := "{}"
	var err error
	var str string
	tt := zlsgo.NewTest(t)

	str, err = Set(demoData, "set", "new set")
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = Set(str, "set.b", true)
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = Set(str, "arr1", []string{"one"})
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = SetRaw("", "arr2", "[1,2,3]")
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = Set("", "arr2", "[1,2,3]")
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = SetRaw(str, "arr3.:3", "[1,2,3]")
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = SetRaw(str, "obj", `{"name":"pp"}`)
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = Set(str, "set\\.ff", 1.66)
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	str, err = SetRaw(str, "newSet", `"haha"`)
	tt.EqualExit(true, err == nil)
	t.Log(str, err)

	_, err = SetRaw(str, "", `haha`)
	tt.EqualExit(true, err != nil)
	t.Log(err)

	newSet := Get(str, "newSet").String()
	t.Log(newSet)
	newStr, err := Delete(str, "newSet")
	tt.EqualExit(false, Get(newStr, "newSet").Exists())
	t.Log(newStr, str, err)

	strBytes := []byte("{}")

	strBytes, err = SetBytes(strBytes, "setBytes", "new set")
	tt.EqualExit(true, err == nil)
	t.Log(string(strBytes), err)

	strBytes, err = SetRawBytes(strBytes, "setRawBytes", []byte("Raw"))
	tt.EqualExit(true, err == nil)
	t.Log(string(strBytes), err)

	strBytes, err = SetRawBytes(strBytes, "set\\.other", []byte("Raw"))
	tt.EqualExit(true, err == nil)
	t.Log(string(strBytes), err)

	strBytes, err = SetRawBytes(strBytes, "set.other", []byte("Raw"))
	tt.EqualExit(true, err == nil)
	t.Log(string(strBytes), err)

	_, _ = SetBytes(strBytes, "setBytes2.s", "new set")

	strBytes, err = SetBytesOptions(strBytes, "setBytes3.op", "new set", &Options{Optimistic: true, ReplaceInPlace: true})

	t.Log(string(strBytes), err)
	_, _ = DeleteBytes(strBytes, "setRawBytes")

	j := struct {
		Name string `json:"n"`
	}{"isName"}
	jj, err := Marshal(j)
	t.Log(string(jj), err)
	t.Log(Stringify(j))
}

func TestSetSt(tt *testing.T) {
	j := struct {
		Name string `json:"n"`
	}{"isName"}
	t := zlsgo.NewTest(tt)
	json, _ := Set("{}", "code", 200)
	json, _ = Set(json, "code2", uint(200))
	json, _ = Set(json, "int8", int8(8))
	json, _ = Set(json, "int32", int32(200))
	json, _ = Set(json, "int64", int64(200))
	json, _ = Set(json, "uint8", uint8(8))
	json, _ = Set(json, "uint32", uint32(200))
	json, _ = Set(json, "uint64", uint64(200))
	json, _ = SetOptions(json, "code2", 200.01, &Options{
		Optimistic: true,
	})
	tt.Log(json)
	json, _ = Set(json, "data", j)
	tt.Log(json)
	t.Equal("isName", Get(json, "data.n").String())
}

func TestSetRawArrayExpansion(t *testing.T) {
	json := "[]"
	expanded, err := SetRaw(json, "2", "3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := Get(expanded, "1").String(); got != "3" {
		t.Fatalf("expanded result %s", expanded)
	}
	if Get(expanded, "0").Raw() != "null" {
		t.Fatalf("expected filler null at index 0, got %s", Get(expanded, "0").Raw())
	}
	json = `{"arr":[0]}`
	updated, err := SetRaw(json, "arr.3", "4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != `{"arr":[0,null,null,4]}` {
		t.Fatalf("unexpected array result %s", updated)
	}
}

func BenchmarkSet(b *testing.B) {
	s := zstring.Rand(100)
	json := "{}"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Set(json, strconv.Itoa(i), s)
	}
}

func BenchmarkSetBytes(b *testing.B) {
	s := []byte(zstring.Rand(100))
	json := []byte("{}")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SetBytes(json, strconv.Itoa(i), s)
	}
}

func TestSetOptionsEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		path     string
		value    interface{}
		expected string
	}{
		{
			name:     "set in nested object",
			json:     `{"a":{"b":{"c":1}}}`,
			path:     "a.b.c",
			value:    2,
			expected: `{"a":{"b":{"c":2}}}`,
		},
		{
			name:     "set new key",
			json:     `{"a":1}`,
			path:     "b",
			value:    2,
			expected: `{"b":2,"a":1}`,
		},
		{
			name:     "set in array",
			json:     `{"arr":[1,2,3]}`,
			path:     "arr.1",
			value:    99,
			expected: `{"arr":[1,99,3]}`,
		},
		{
			name:     "set deep nested",
			json:     `{}`,
			path:     "a.b.c.d",
			value:    "deep",
			expected: `{"a":{"b":{"c":{"d":"deep"}}}}`,
		},
		{
			name:     "set with escaped key",
			json:     `{}`,
			path:     `key\.with\.dots`,
			value:    "value",
			expected: `{"key.with.dots":"value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SetOptions(tt.json, tt.path, tt.value, &Options{Optimistic: true})
			if err != nil {
				t.Fatalf("SetOptions failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("SetOptions() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSetRawBytesOptions(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		path     string
		value    string
		expected string
	}{
		{
			name:     "set raw object",
			json:     `{"a":1}`,
			path:     "b",
			value:    `{"nested":true}`,
			expected: `{"b":{"nested":true},"a":1}`,
		},
		{
			name:     "set raw array",
			json:     `{"a":1}`,
			path:     "arr",
			value:    `[1,2,3]`,
			expected: `{"arr":[1,2,3],"a":1}`,
		},
		{
			name:     "replace with raw",
			json:     `{"a":"old"}`,
			path:     "a",
			value:    `"new"`,
			expected: `{"a":"new"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SetRawBytesOptions([]byte(tt.json), tt.path, []byte(tt.value), &Options{})
			if err != nil {
				t.Fatalf("SetRawBytesOptions failed: %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("SetRawBytesOptions() = %q, want %q", string(result), tt.expected)
			}
		})
	}
}

func TestSetRawOptions(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		path     string
		value    string
		opts     *Options
		expected string
	}{
		{
			name:     "set raw with optimistic",
			json:     `{"a":1}`,
			path:     "b.c",
			value:    `"value"`,
			opts:     &Options{Optimistic: true},
			expected: `{"b":{"c":"value"},"a":1}`,
		},
		{
			name:     "set raw replace",
			json:     `{"a":{"b":1}}`,
			path:     "a.b",
			value:    `2`,
			opts:     &Options{},
			expected: `{"a":{"b":2}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SetRawOptions(tt.json, tt.path, tt.value, tt.opts)
			if err != nil {
				t.Fatalf("SetRawOptions failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("SetRawOptions() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDeleteEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		path     string
		expected string
	}{
		{
			name:     "delete key",
			json:     `{"a":1,"b":2,"c":3}`,
			path:     "b",
			expected: `{"a":1,"c":3}`,
		},
		{
			name:     "delete nested key",
			json:     `{"a":{"b":{"c":1}}}`,
			path:     "a.b.c",
			expected: `{"a":{"b":{}}}`,
		},
		{
			name:     "delete array element",
			json:     `{"arr":[1,2,3]}`,
			path:     "arr.1",
			expected: `{"arr":[1,3]}`,
		},
		{
			name:     "delete non-existent",
			json:     `{"a":1}`,
			path:     "b",
			expected: `{"a":1}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Delete(tt.json, tt.path)
			if err != nil {
				t.Fatalf("Delete failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Delete() = %q, want %q", result, tt.expected)
			}
		})
	}
}
