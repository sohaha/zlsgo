//go:build go1.21
// +build go1.21

package zarray_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zarray"
)

func TestIntersection(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Run("Basic intersection", func(subTT *zlsgo.TestUtil) {
		list1 := []int{1, 2, 3, 4, 5}
		list2 := []int{4, 5, 6, 7, 8}
		expected1 := []int{4, 5}
		subTT.Equal(expected1, zarray.Intersection(list1, list2))
		subTT.Equal([]int{4, 5}, zarray.Intersection(list2, list1))
	})

	tt.Run("No common elements", func(subTT *zlsgo.TestUtil) {
		list3 := []int{1, 2, 3}
		list4 := []int{4, 5, 6}
		expected2 := []int{}
		subTT.Equal(expected2, zarray.Intersection(list3, list4))
	})

	tt.Run("One list empty", func(subTT *zlsgo.TestUtil) {
		list5 := []int{1, 2, 3}
		list6 := []int{}
		expected3 := []int{}
		subTT.Equal(expected3, zarray.Intersection(list5, list6))
		subTT.Equal(expected3, zarray.Intersection(list6, list5))
	})

	tt.Run("Both lists empty", func(subTT *zlsgo.TestUtil) {
		list7 := []int{}
		list8 := []int{}
		expected4 := []int{}
		subTT.Equal(expected4, zarray.Intersection(list7, list8))
	})

	tt.Run("With duplicate elements", func(subTT *zlsgo.TestUtil) {
		list9 := []int{1, 2, 2, 3, 4, 4, 5}
		list10 := []int{2, 4, 4, 5, 6, 6}
		expected5 := []int{2, 4, 5}
		subTT.Equal(expected5, zarray.Intersection(list9, list10))
		subTT.Equal([]int{2, 4, 5}, zarray.Intersection(list10, list9))
	})

	tt.Run("All elements common", func(subTT *zlsgo.TestUtil) {
		list11 := []int{1, 2, 3}
		list12 := []int{1, 2, 3}
		expected6 := []int{1, 2, 3}
		subTT.Equal(expected6, zarray.Intersection(list11, list12))
	})

	tt.Run("Larger lists with some common elements", func(subTT *zlsgo.TestUtil) {
		list13 := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
		list14 := []int{5, 15, 25, 30, 35, 45, 50, 55, 65, 70, 75}
		expected7 := []int{30, 50, 70}
		subTT.Equal(expected7, zarray.Intersection(list13, list14))
		subTT.Equal([]int{30, 50, 70}, zarray.Intersection(list14, list13))
	})

	tt.Run("String slices", func(subTT *zlsgo.TestUtil) {
		strList1 := []string{"apple", "banana", "cherry"}
		strList2 := []string{"banana", "date", "fig", "apple"}
		expectedStr1 := []string{"apple", "banana"}
		subTT.Equal(expectedStr1, zarray.Intersection(strList1, strList2))
		expectedStr2 := []string{"banana", "apple"}
		subTT.Equal(expectedStr2, zarray.Intersection(strList2, strList1))
	})

	tt.Run("One list is a subset of the other", func(subTT *zlsgo.TestUtil) {
		list15 := []int{1, 2, 3, 4, 5}
		list16 := []int{2, 3, 4}
		expected9_1 := []int{2, 3, 4}
		subTT.Equal(expected9_1, zarray.Intersection(list15, list16))
		expected9_2 := []int{2, 3, 4}
		subTT.Equal(expected9_2, zarray.Intersection(list16, list15))
	})

	tt.Run("Elements at beginning and end", func(subTT *zlsgo.TestUtil) {
		list17 := []int{1, 2, 9, 10}
		list18 := []int{1, 5, 8, 10}
		expected10_1 := []int{1, 10}
		subTT.Equal(expected10_1, zarray.Intersection(list17, list18))
		expected10_2 := []int{1, 10}
		subTT.Equal(expected10_2, zarray.Intersection(list18, list17))
	})
}
