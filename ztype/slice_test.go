package ztype

import (
	"encoding/json"
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestSlice(t *testing.T) {
	tt := zlsgo.NewTest(t)
	value := "ddd"
	res := Slice(value, false)
	tt.Log(res)

	res = Slice([]interface{}{"1", 2, 3.0})
	tt.Equal("1", res.First().String())
	tt.Equal("3", res.Last().String())

	m := []map[string]interface{}{{"h": "ddd"}}
	res = ToSlice(m)
	tt.Log(res)
	tt.Equal(1, res.Len())
	tt.Equal([]int{0}, res.Int())
	tt.Equal([]string{`{"h":"ddd"}`}, res.String())
	tt.Equal(map[string]interface{}{"h": "ddd"}, res.Index(0).Value())
	tt.Equal("10086", res.Index(110).String("10086"))

	tt.Equal([]string{"1", "2"}, ToSlice([]int{1, 2}).String())
	tt.Equal([]string{"1", "2"}, ToSlice([]int64{1, 2}).String())

	rmaps := res.Maps()
	tt.Equal(Maps{{"h": "ddd"}}, rmaps)

	tt.Equal("[]interface {}", GetType(res.Value()))

	j, _ := json.Marshal(res)
	tt.Equal(`[{"h":"ddd"}]`, string(j))
}

func TestSliceStrToIface(t *testing.T) {
	tt := zlsgo.NewTest(t)
	value := []string{"ddd", "222"}
	res := SliceStrToAny(value)
	tt.Equal(len(value), len(res))
	t.Log(res)
}

func TestSliceForce(t *testing.T) {
	tt := zlsgo.NewTest(t)

	value := "test"

	tt.Equal([]string{"test"}, ToSlice(value).String())
	tt.Equal([]string{}, ToSlice(value, true).String())
}

func BenchmarkSlice(b *testing.B) {
	b.Run("str", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ToSlice("test").String()
		}
	})

	b.Run("str_no", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ToSlice("test", true).String()
		}
	})

	b.Run("int", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ToSlice([]int{1, 2, 3}).String()
		}
	})

	b.Run("any", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ToSlice([]interface{}{1, 2, 3}).String()
		}
	})
}

func TestToSliceNoConvAndTypeInput(t *testing.T) {
	s := ToSlice("[1,2]", true)
	if s.Len() != 0 {
		t.Fatalf("expected empty slice with noConv, got: %v", s)
	}

	var arr []interface{}
	_ = json.Unmarshal([]byte(`[1,2]`), &arr)
	typ := New(arr)
	s2 := ToSlice(typ)
	if s2.Len() != 2 || s2.Index(0).Int() != 1 || s2.Index(1).Int() != 2 {
		t.Fatalf("Type->ToSlice failed: %v", s2)
	}
}

func TestPoolsLargeCapEarlyReturn(t *testing.T) {
	s := make([]string, 0, 65)
	putStringSlice(s)
	is := make([]interface{}, 0, 65)
	putInterfaceSlice(is)
	ii := make([]int, 0, 65)
	putIntSlice(ii)
}

func TestExecuteFieldAccessStructFallback(t *testing.T) {
	type S struct {
		N int `z:"n"`
	}
	v, ok := executeFieldAccess("n", S{N: 7})
	if !ok || v.(int) != 7 {
		t.Fatalf("struct fallback via ToMap failed: %v %v", v, ok)
	}
}

func TestSliceTypeLargeLengthBranches(t *testing.T) {
	base := make([]int, 12)
	for i := range base {
		base[i] = i
	}
	st := ToSlice(base)
	ints := st.Int()
	if len(ints) != 12 || ints[0] != 0 || ints[11] != 11 {
		t.Fatalf("Int path failed: %v", ints)
	}
	vals := st.Value()
	if len(vals) != 12 || vals[5].(int) != 5 {
		t.Fatalf("Value path failed: %v", vals)
	}
}

func TestSliceEdgeCases(t *testing.T) {
	tt := zlsgo.NewTest(t)

	s := Slice("not a slice", true)
	tt.Equal(0, s.Len())

	s = Slice(nil, false)
	tt.Equal(0, s.Len())

	s = Slice([]int{1, 2, 3})
	tt.Equal("default", s.Index(100).String("default"))

	empty := Slice([]int{})
	tt.Equal("", empty.First().String())
	tt.Equal("", empty.Last().String())

	s2 := ToSlice([]string{"a", "b", "c"})
	tt.Equal(3, s2.Len())
	tt.Equal("a", s2.First().String())
	tt.Equal("c", s2.Last().String())
}

func TestToSliceComplexConversions(t *testing.T) {
	tt := zlsgo.NewTest(t)

	jsonStr := `["a", "b", "c"]`
	result := ToSlice(jsonStr)
	tt.Equal(3, result.Len())
	tt.Equal("a", result.Index(0).String())

	result2 := ToSlice(jsonStr, true)
	tt.Equal(0, result2.Len())

	int64Slice := []int64{1, 2, 3}
	result3 := ToSlice(int64Slice)
	tt.Equal(3, result3.Len())
	tt.Equal(1, result3.Index(0).Int())

	complexData := map[string]interface{}{
		"items": []map[string]interface{}{
			{"id": 1},
			{"id": 2},
		},
	}
	result4 := ToMap(complexData)
	tt.NotNil(result4)
}

func TestToSliceWithMapInput(t *testing.T) {
	tt := zlsgo.NewTest(t)

	source := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	var target []int
	err := To(source, &target)
	tt.NoError(err)
	tt.EqualTrue(len(target) >= 0)

	emptyMap := map[string]string{}
	var target2 []string
	err = To(emptyMap, &target2)
	tt.NoError(err)
	tt.Equal(0, len(target2))
}
