//go:build go1.18
// +build go1.18

package zarray_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zarray"
)

func TestSlice(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.Equal([]string{"a", "b", "c"}, zarray.Slice[string]("a,b,c"))
	tt.Equal([]int{1, 2, 3}, zarray.Slice[int]("1,2,3"))
	tt.Equal([]float64{1.1, 2.2, 3.3}, zarray.Slice[float64]("1.1,2.2,3.3"))
	tt.Equal([]string{"1.1", "2.2,3.3"}, zarray.Slice[string]("1.1,2.2,3.3", 2))
	tt.Equal([]int{}, zarray.Slice[int](""))
}

func TestJoin(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.Equal("a,b,c", zarray.Join([]string{"a", "b", "c"}, ","))
	tt.Equal("1,2,3", zarray.Join([]int{1, 2, 3}, ","))
	tt.Equal("1.1,2.2,3.3", zarray.Join([]float64{1.1, 2.2, 3.3}, ","))
	tt.Equal("1.1,2.2,3.3", zarray.Join([]string{"1.1", "2.2", "3.3"}, ","))
	tt.Equal("1.1,3.3", zarray.Join([]string{"1.1", "", "3.3"}, ","))
	tt.Equal("", zarray.Join([]string{}, ","))
}
