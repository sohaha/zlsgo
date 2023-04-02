package ztype

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestSlice(t *testing.T) {
	tt := zlsgo.NewTest(t)
	value := "ddd"
	res := Slice(value)
	tt.Log(res)
	m := []map[string]interface{}{{"h": "ddd"}}
	res = ToSlice(m)
	tt.Log(res)

	tt.Equal(1, res.Len())
	tt.Equal([]int{0}, res.Int())
	tt.Equal([]string{`{"h":"ddd"}`}, res.String())
	tt.Equal(map[string]interface{}{"h": "ddd"}, res.Index(0).Value())

	rmaps := res.Maps()
	tt.Equal(Maps{{"h": "ddd"}}, rmaps)

	tt.Equal("[]interface {}", GetType(res.Value()))
}

func TestSliceStrToIface(t *testing.T) {
	tt := zlsgo.NewTest(t)
	value := []string{"ddd", "222"}
	res := SliceStrToIface(value)
	tt.Equal(len(value), len(res))
	t.Log(res)
}
