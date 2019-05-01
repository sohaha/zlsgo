package zarray_test

import (
	"testing"

	"github.com/sohaha/zlsgo/zarray"
	. "github.com/sohaha/zlsgo/ztest"
)

func TestArray(t *testing.T) {
	var err error

	array := garray.New(20)

	array = garray.New()
	Equal(t, true, array.IsEmpty())
	for i := 0; i < 10; i++ {
		if err := array.Add(i, i+1); err != nil {
			t.Error(err)
			break
		}
	}
	err = array.Add(99, "无效")

	Equal(t, true, err != nil)
	_, err = array.Get(99)
	Equal(t, true, err != nil)
	err = array.Set(99, "无效")

	Equal(t, true, err != nil)
	array.Unshift("第一")
	array.Push("最后")
	Equal(t, true, array.Includes("第一"))
	Equal(t, false, array.Includes("第一百"))
	Equal(t, 0, array.IndexOf("第一"))
	Equal(t, -1, array.IndexOf("第一百"))
	Equal(t, 20, array.CapLength())
	Equal(t, 12, array.Length())
	last, _ := array.Get(0)
	Equal(t, "第一", last)
	array.Set(0, "one")
	one := []string{"one"}
	shift, _ := array.Shift()
	Equal(t, one[0], shift.([]interface{})[0])
	array.Remove(99)
	array.RemoveValue("最后")
	pop, _ := array.Pop()
	Equal(t, 10, pop.([]interface{})[0])
	Equal(t, 9, array.Length())
	for i := 0; i < 9; i++ {
		array.Remove(i)
	}
	Equal(t, 4, array.Length())
	array.Clear()
	Equal(t, 0, array.Length())
}
