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
	res = Slice(m)
	tt.Log(res)
}

func TestSliceStrToIface(t *testing.T) {
	tt := zlsgo.NewTest(t)
	value := []string{"ddd", "222"}
	res := SliceStrToIface(value)
	tt.Equal(len(value), len(res))
	t.Log(res)
}
