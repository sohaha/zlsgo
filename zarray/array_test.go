package zarray_test

import (
	"testing"

	zls "github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zarray"
)

func TestArray(t *testing.T) {
	tt := zls.NewTest(t)
	var err error
	array := zarray.New()
	tt.Equal(true, array.IsEmpty())
	for i := 0; i < 10; i++ {
		if err := array.Add(i, i+1); err != nil {
			t.Error(err)
			break
		}
	}
	err = array.Add(99, "无效")
	tt.Equal(true, err != nil)
	_, err = array.Get(99)
	tt.Equal(true, err != nil)
	err = array.Set(99, "无效")
	tt.Equal(true, err != nil)
	_ = array.Unshift("第一")
	array.Push("最后")
	tt.Equal(true, array.Contains("第一"))
	tt.Equal(false, array.Contains("第一百"))
	tt.Equal(0, array.Index("第一"))
	tt.Equal(-1, array.Index("第一百"))
	tt.Equal(20, array.CapLength())
	tt.Equal(12, array.Length())
	last, _ := array.Get(0)
	tt.Equal("第一", last)
	_ = array.Set(0, "one")
	one := []string{"one"}
	shift, _ := array.Shift()
	oneArr, _ := zarray.Copy(shift)
	_ = array.Raw()
	_, copyErr := zarray.Copy("shift")
	tt.Equal(true, copyErr != nil)
	tt.Equal(one[0], shift.([]interface{})[0])
	copyValue, _ := oneArr.Get(0)
	tt.Equal(one[0], copyValue)
	_, _ = array.Remove(99)
	_, _ = array.RemoveValue("最后")
	pop, _ := array.Pop()
	tt.Equal(10, pop.([]interface{})[0])
	tt.Equal(9, array.Length())
	for i := 0; i < 9; i++ {
		_, _ = array.Remove(i, 2)
	}
	array.Format()
	tt.Equal(3, array.Length())
	array.Clear()
	tt.Equal(0, array.Length())
	v, _ := array.Get(1991, "成功")
	tt.Equal("成功", v)

	array = zarray.New(100)
	array = array.Map(func(i int, v interface{}) interface{} {
		return i
	})
	newArray := array.Shuffle()
	t.Log(newArray.Raw())
	t.Log(array.Raw())
}

var testdata = []interface{}{1, 2, 3, 4, 5, 6, 7}

func BenchmarkArrayNew(b *testing.B) {
	arr, _ := zarray.Copy(testdata)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v, _ := arr.Get(i, 7)
		_ = v
	}
}

func BenchmarkArrayRaw(b *testing.B) {
	arr := testdata
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if len(arr) <= i {
			v := "2"
			_ = v
			continue
		}
		v := arr[i]
		_ = v
	}
}
