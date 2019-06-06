package zarray_test

import (
	"testing"

	zls "github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zarray"
)

func TestArray(t *testing.T) {
	T := zls.NewTest(t)
	var err error
	array := zarray.New(20)
	array = zarray.New()
	T.Equal(true, array.IsEmpty())
	for i := 0; i < 10; i++ {
		if err := array.Add(i, i+1); err != nil {
			t.Error(err)
			break
		}
	}
	err = array.Add(99, "无效")
	T.Equal(true, err != nil)
	_, err = array.Get(99)
	T.Equal(true, err != nil)
	err = array.Set(99, "无效")
	T.Equal(true, err != nil)
	array.Unshift("第一")
	array.Push("最后")
	T.Equal(true, array.Contains("第一"))
	T.Equal(false, array.Contains("第一百"))
	T.Equal(0, array.Index("第一"))
	T.Equal(-1, array.Index("第一百"))
	T.Equal(20, array.CapLength())
	T.Equal(12, array.Length())
	last, _ := array.Get(0)
	T.Equal("第一", last)
	array.Set(0, "one")
	one := []string{"one"}
	shift, _ := array.Shift()
	oneArr, _ := zarray.Copy(shift)
	_ = array.Raw()
	_, copyErr := zarray.Copy("shift")
	T.Equal(true, copyErr != nil)
	T.Equal(one[0], shift.([]interface{})[0])
	copyValue, _ := oneArr.Get(0)
	T.Equal(one[0], copyValue)
	array.Remove(99)
	array.RemoveValue("最后")
	pop, _ := array.Pop()
	T.Equal(10, pop.([]interface{})[0])
	T.Equal(9, array.Length())
	for i := 0; i < 9; i++ {
		array.Remove(i, 2)
	}
	array.Format()
	T.Equal(3, array.Length())
	array.Clear()
	T.Equal(0, array.Length())
	v, _ := array.Get(1991, "成功")
	T.Equal("成功", v)
}
