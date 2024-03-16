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
