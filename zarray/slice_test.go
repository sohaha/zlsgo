//go:build go1.18
// +build go1.18

package zarray_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zarray"
)

var l = []int{0, 1, 2, 3, 4, 5}
var l2 = []int{0, 1, 2, 3, 4, 5, 2, 34, 5, 6, 7, 98, 6, 67, 54, 543, 345, 435, 43543, 435, 3, 2, 42, 3423, 54, 6, 5}

func TestShuffle(t *testing.T) {
	t.Log(zarray.Shuffle(l))
	t.Log(zarray.Shuffle(l2))
}

func TestFilter(t *testing.T) {
	tt := zlsgo.NewTest(t)
	nl := zarray.Filter(l, func(index int, item int) bool {
		t.Log(index, item)
		return item%2 == 0
	})
	tt.Equal([]int{0, 2, 4}, nl)
}
