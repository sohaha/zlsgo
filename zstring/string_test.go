/*
 * @Author: seekwe
 * @Date:   2019-05-09 12:44:23
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-06-06 16:11:33
 */
package zstring

import (
	"testing"

	zls "github.com/sohaha/zlsgo"
)

func TestString(t *testing.T) {

	T := zls.NewTest(t)
	l := "我的这里一共8字"
	T.Equal(8, Len(l))

	s := "我的长度是二十,不够右边补零"
	T.Equal("我的长度是二十,不够右边补零000000", Pad(s, 20, "0", PadRight))

	s2 := "我的长度是二十,不够左边补零"
	T.Equal("000000我的长度是二十,不够左边补零", Pad(s2, 20, "0", PadLeft))

	s3 := "我的长度很长不需要填充"
	T.Equal("我的长度很长不需要填充", Pad(s3, 5, "我的长度很长不需要填充", PadRight))

	T.Equal("长度", Substr(s3, 2, 2))

	s4 := "我的长度是二十,不够两边补零"
	T.Equal("000我的长度是二十,不够两边补零000", Pad(s4, 20, "0", PadSides))
}
