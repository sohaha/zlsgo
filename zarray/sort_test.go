//go:build go1.18
// +build go1.18

package zarray_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zarray"
)

func TestNewSortMap(t *testing.T) {
	tt := zlsgo.NewTest(t)
	count := 50
	arr := make([]int, 0, count)

	m := zarray.NewSortMap[int, int]()
	for i := 0; i < count; i++ {
		arr = append(arr, i)
		m.Set(i, i)
	}

	v, has := m.Get(2)
	tt.EqualTrue(has)
	tt.Equal(2, v)

	res := make([]int, 0, count)
	m.ForEach(func(key int, value int) bool {
		tt.Equal(key, value)
		res = append(res, key)
		return true
	})
	t.Log(res)

	tt.Equal(arr, res)
	tt.Equal(count, m.Len())

	m.Delete(1, 10, 20, 20)
	tt.Equal(count-3, m.Len())

	t.Log(m.Keys())
}
