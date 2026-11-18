package ztype

import (
	"testing"
)

// The pooled scratch buffers that used to back SliceType.Value/String/Int and
// parseTagOptions were removed: every caller owns the returned slice, so the
// pool always cost an extra allocation plus a copy on top of the unavoidable
// result allocation, and recycling the header outside the pool left stale
// references in the backing array. These tests pin down the replacement
// contract: exactly one allocation, and a result independent from any other.

func TestSliceTypeConversionsSingleAlloc(t *testing.T) {
	s := Slice([]interface{}{1, "2", 3.0, true})

	cases := []struct {
		name string
		fn   func()
	}{
		{"Value", func() { _ = s.Value() }},
		{"String", func() { _ = s.String() }},
		{"Int", func() { _ = s.Int() }},
	}
	for _, c := range cases {
		allocs := testing.AllocsPerRun(200, c.fn)
		if allocs != 1 {
			t.Fatalf("%s allocs = %v, want 1 (result slice only)", c.name, allocs)
		}
	}
}

func TestSliceTypeResultsIndependent(t *testing.T) {
	s := Slice([]interface{}{"a", "b"})

	v1, v2 := s.Value(), s.Value()
	v1[0] = "mutated"
	if v2[0] == "mutated" {
		t.Fatal("Value() results share a backing array")
	}

	s1, s2 := s.String(), s.String()
	s1[0] = "mutated"
	if s2[0] == "mutated" {
		t.Fatal("String() results share a backing array")
	}

	i1, i2 := Slice([]interface{}{1, 2}).Int(), Slice([]interface{}{1, 2}).Int()
	i1[0] = 99
	if i2[0] == 99 {
		t.Fatal("Int() results share a backing array")
	}
}

func TestSliceTypeEmptyResults(t *testing.T) {
	var s SliceType
	if got := s.Value(); got == nil || len(got) != 0 {
		t.Fatalf("Value() on empty slice = %#v, want empty non-nil", got)
	}
	if got := s.String(); got == nil || len(got) != 0 {
		t.Fatalf("String() on empty slice = %#v, want empty non-nil", got)
	}
	if got := s.Int(); got == nil || len(got) != 0 {
		t.Fatalf("Int() on empty slice = %#v, want empty non-nil", got)
	}
}

func TestParseTagOptionsSingleAlloc(t *testing.T) {
	allocs := testing.AllocsPerRun(200, func() {
		_ = parseTagOptions("omitempty,string,min=1")
	})
	if allocs != 1 {
		t.Fatalf("parseTagOptions allocs = %v, want 1 (strings.Split only)", allocs)
	}
}

func TestParseTagOptionsContract(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{" ", nil},
		{" , , ", nil},
		{"omitempty", []string{"omitempty"}},
		{"omitempty,string,min=1", []string{"omitempty", "string", "min=1"}},
		{" a , b , c ", []string{"a", "b", "c"}},
		{"a,,b", []string{"a", "b"}},
	}
	for _, c := range cases {
		got := parseTagOptions(c.in)
		if len(got) != len(c.want) {
			t.Fatalf("parseTagOptions(%q) = %#v, want %#v", c.in, got, c.want)
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Fatalf("parseTagOptions(%q)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}

type poolProbeItem struct {
	A int    `json:"a"`
	B string `json:"b"`
}

type poolProbeHolder struct {
	Items []poolProbeItem `json:"items"`
}

func TestToMapStructSliceAllocs(t *testing.T) {
	h := poolProbeHolder{Items: []poolProbeItem{{A: 1, B: "x"}, {A: 2, B: "y"}}}

	got := ToMap(h)
	items, ok := got["items"].([]map[string]interface{})
	if !ok || len(items) != 2 {
		t.Fatalf("ToMap struct slice = %#v", got["items"])
	}
	if items[0]["a"] != 1 || items[1]["b"] != "y" {
		t.Fatalf("unexpected struct slice mapping: %#v", items)
	}

	// Two element maps plus their contents; the scratch slice is no longer
	// allocated from a pool, so this stays flat between runs.
	allocs := testing.AllocsPerRun(50, func() {
		_ = ToMap(h)
	})
	if allocs > 12 {
		t.Fatalf("ToMap struct-slice allocs = %v, want <= 12", allocs)
	}
}

func BenchmarkSliceTypeValue(b *testing.B) {
	s := Slice([]interface{}{1, "2", 3.0, true})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = s.Value()
	}
}

func BenchmarkParseTagOptions(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = parseTagOptions("omitempty,string,min=1")
	}
}
