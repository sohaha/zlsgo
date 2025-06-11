//go:build go1.18
// +build go1.18

package zarray_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/ztype"
)

var (
	l  = []int{0, 1, 2, 3, 4, 5}
	l2 = []int{0, 1, 2, 3, 4, 5, 2, 34, 5, 6, 7, 98, 6, 67, 54, 543, 345, 435, 43543, 435, 3, 2, 42, 3423, 54, 6, 5}
)

func TestShuffle(t *testing.T) {
	t.Log(zarray.Shuffle(l))
	t.Log(zarray.Shuffle(l2))
}

func TestRand(t *testing.T) {
	t.Log(zarray.Rand(l))
}

func TestReverse(t *testing.T) {
	t.Log(zarray.Reverse(l))
}

func TestFilter(t *testing.T) {
	tt := zlsgo.NewTest(t)
	nl := zarray.Filter(l, func(index int, item int) bool {
		t.Log(index, item)
		return item%2 == 0
	})
	tt.Equal([]int{0, 2, 4}, nl)
}

func TestMap(t *testing.T) {
	tt := zlsgo.NewTest(t)
	s := []int{1, 2, 3}
	nl := zarray.Map(s, func(i int, v int) string {
		return ztype.ToString(v) + "//"
	})
	tt.Equal([]string{"1//", "2//", "3//"}, nl)
}

func TestParallelMap(t *testing.T) {
	tt := zlsgo.NewTest(t)
	{
		expected := zarray.Map(l2, func(i int, v int) string {
			return ztype.ToString(v) + "//"
		})

		actual := zarray.ParallelMap(l2, func(i int, v int) string {
			return ztype.ToString(v) + "//"
		}, uint(len(l2)+1))
		tt.Equal(expected, actual)
	}

	{
		expected := zarray.Map(l2, func(i int, v int) string {
			return ztype.ToString(v) + "//"
		})

		actual := zarray.ParallelMap(l2, func(i int, v int) string {
			return ztype.ToString(v) + "//"
		}, 0)
		tt.Equal(expected, actual)
	}
}

func TestDiff(t *testing.T) {
	tt := zlsgo.NewTest(t)

	n1, n2 := zarray.Diff(l2, l)

	t.Log(l2, l)
	t.Log(n1, n2)
	tt.Equal([]int{34, 6, 7, 98, 6, 67, 54, 543, 345, 435, 43543, 435, 42, 3423, 54, 6}, n1)
	tt.Equal([]int{}, n2)
}

func TestPop(t *testing.T) {
	tt := zlsgo.NewTest(t)

	l1 := []int{0, 1, 2, 3, 4, 5}

	tt.Equal(5, zarray.Pop(&l1))
	tt.Equal(4, zarray.Pop(&l1))

	tt.Equal([]int{0, 1, 2, 3}, l1)
}

func TestShift(t *testing.T) {
	tt := zlsgo.NewTest(t)

	l1 := []int{0, 1, 2, 3, 4, 5}

	tt.Equal(0, zarray.Shift(&l1))
	tt.Equal(1, zarray.Shift(&l1))

	tt.Equal([]int{2, 3, 4, 5}, l1)
}

func TestContains(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.EqualTrue(!zarray.Contains(l, 54))
	tt.EqualTrue(!zarray.Contains(l, 6))
	tt.EqualTrue(zarray.Contains(l2, 5))
	tt.EqualTrue(zarray.Contains(l2, 6))
	tt.EqualTrue(zarray.Contains(l2, 54))
}

func TestUnique(t *testing.T) {
	tt := zlsgo.NewTest(t)
	a := append(l, l2...)
	unia := zarray.Unique(a)
	tt.Equal(18, len(unia))
	tt.EqualTrue(len(a) != len(unia))
	t.Log(unia)
}

func TestFind(t *testing.T) {
	tt := zlsgo.NewTest(t)
	a := []map[string]string{
		{"name": "a"},
		{"name": "b"},
		{"name": "c"},
	}

	v, ok := zarray.Find(a, func(_ int, v map[string]string) bool {
		return v["name"] == "b"
	})
	tt.EqualTrue(ok)
	tt.Equal("b", v["name"])

	v, ok = zarray.Find(a, func(_ int, v map[string]string) bool {
		return v["name"] == "z"
	})
	tt.EqualTrue(!ok)
	tt.Equal("", v["name"])
}

func TestRandPickN(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.Equal(0, len(zarray.RandPickN[int](nil, 3)))
	tt.Equal(3, len(zarray.RandPickN(l, 3)))
	tt.Equal(len(l), len(zarray.RandPickN(l, 10)))
}

func TestChunk(t *testing.T) {
	tt := zlsgo.NewTest(t)
	tt.Equal([][]int{{0, 1, 2, 3, 4, 5}}, zarray.Chunk(l, 6))
	tt.Equal([][]int{{0, 1, 2}, {3, 4, 5}}, zarray.Chunk(l, 3))
	tt.Equal([][]int{{0, 1, 2, 3, 4}, {5}}, zarray.Chunk(l, 5))
	tt.Equal([][]int{{0, 1, 2, 3, 4, 5}}, zarray.Chunk(l, 8))
	tt.Equal([][]int{}, zarray.Chunk(l, 0))
	tt.Equal([][]int{}, zarray.Chunk[int](nil, 8))
}

func TestRandShift(t *testing.T) {
	tt := zlsgo.NewTest(t)
	ll := zarray.RandShift(l)
	for i := 0; i < 8; i++ {
		v, err := ll()
		if i > 5 {
			tt.NotNil(err)
		} else {
			tt.NoError(err)
		}
		tt.Log(v, err)
	}
}

func TestSortWithPriority(t *testing.T) {
	tt := zlsgo.NewTest(t)

	tt.Run("Base", func(tt *zlsgo.TestUtil) {
		tt.Equal([]int{5, 10, 4, 6, 7, 8, 9, 2, 3, 1}, zarray.SortWithPriority([]int{1, 4, 2, 3, 5, 6, 7, 8, 9, 10}, []int{5, 10}, []int{2, 3, 1}))

		v := []string{"是", "!", "我", "测试", "的"}

		d := zarray.SortWithPriority(v, []string{"我", "是"}, []string{"的", "!"})
		tt.Log(d)
		tt.Equal([]string{"我", "是", "测试", "的", "!"}, d)

		d = zarray.SortWithPriority(v, []string{"我", "是"}, nil)
		tt.Log(d)
		tt.Equal([]string{"我", "是", "!", "测试", "的"}, d)

		d = zarray.SortWithPriority(v, nil, []string{"测试", "是"})
		tt.Log(d)
		tt.Equal([]string{"!", "我", "的", "测试", "是"}, d)
	})

	tt.Run("Empty", func(tt *zlsgo.TestUtil) {
		tt.Equal([]int{}, zarray.SortWithPriority(nil, []int{}, []int{}))
		tt.Equal([]int{}, zarray.SortWithPriority([]int{}, nil, nil))
		tt.Equal([]int{2, 6}, zarray.SortWithPriority([]int{2, 6}, nil, nil))
	})

	tt.Run("DuplicateElements", func(tt *zlsgo.TestUtil) {
		slice := []int{1, 2, 3, 2, 4, 3, 5}
		result := zarray.SortWithPriority(slice, []int{3, 1}, []int{5, 2})
		tt.Equal([]int{3, 3, 1, 4, 5, 2, 2}, result)

		result = zarray.SortWithPriority(slice, []int{3, 1, 3}, []int{5})
		tt.Equal([]int{1, 3, 3, 2, 2, 4, 5}, result)

		result = zarray.SortWithPriority(slice, []int{1}, []int{2, 5, 2})
		tt.Equal([]int{1, 3, 4, 3, 5, 2, 2}, result)
	})

	tt.Run("NonExistentElements", func(tt *zlsgo.TestUtil) {
		slice := []int{1, 2, 3, 4, 5}
		result := zarray.SortWithPriority(slice, []int{6, 7, 3}, []int{5})
		tt.Equal([]int{3, 1, 2, 4, 5}, result)

		result = zarray.SortWithPriority(slice, []int{1}, []int{6, 7, 5})
		tt.Equal([]int{1, 2, 3, 4, 5}, result)

		result = zarray.SortWithPriority(slice, []int{6, 1}, []int{7, 5})
		tt.Equal([]int{1, 2, 3, 4, 5}, result)
	})

	tt.Run("OverlappingElements", func(tt *zlsgo.TestUtil) {
		slice := []int{1, 2, 3, 4, 5}
		result := zarray.SortWithPriority(slice, []int{3, 5}, []int{2, 3})
		tt.Equal([]int{5, 1, 4, 2, 3}, result)
	})

	tt.Run("LargeSlice", func(tt *zlsgo.TestUtil) {
		size := 1000
		slice := make([]int, size)
		for i := 0; i < size; i++ {
			slice[i] = size - i
		}

		first := []int{500, 600, 700}
		last := []int{100, 200, 300}

		result := zarray.SortWithPriority(slice, first, last)

		tt.Equal(500, result[0])
		tt.Equal(600, result[1])
		tt.Equal(700, result[2])

		tt.Equal(100, result[size-3])
		tt.Equal(200, result[size-2])
		tt.Equal(300, result[size-1])
	})

	tt.Run("StructSlice", func(tt *zlsgo.TestUtil) {
		type Person struct {
			Name string
			Age  int
		}

		people := []Person{
			{"Alice", 30},
			{"Bob", 25},
			{"Charlie", 35},
			{"David", 28},
			{"Eve", 22},
		}

		first := []Person{{"Charlie", 35}, {"Alice", 30}}
		last := []Person{{"Eve", 22}, {"Bob", 25}}

		result := zarray.SortWithPriority(people, first, last)

		expected := []Person{
			{"Charlie", 35},
			{"Alice", 30},
			{"David", 28},
			{"Eve", 22},
			{"Bob", 25},
		}
		tt.Equal(expected, result)
	})
}
