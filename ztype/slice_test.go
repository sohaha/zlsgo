package ztype

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestSlice(T *testing.T) {
	t := zlsgo.NewTest(T)
	value := "ddd"
	res := Slice(value)
	t.Log(res)
	m := []map[string]interface{}{{"h": "ddd"}}
	res = Slice(m)
	t.Log(res)
}
