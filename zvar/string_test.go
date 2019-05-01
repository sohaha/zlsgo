package zvar_test

import (
	. "github.com/sohaha/zlsgo/ztest"
	. "github.com/sohaha/zlsgo/zvar"
	"testing"
)

func TestString(t *testing.T) {
	l := "我的这里一共8字"
	Equal(t, 8, StrLen(l))

	s := "我的长度是二十,不够就右边补零"
	Equal(t, "我的长度是二十,不够就右边补零零零零零零", StrPad(s, 20, "零", StrPadRight))

	s2 := "我的长度是二十,不够就左边补零"
	Equal(t, "零零零零零我的长度是二十,不够就左边补零", StrPad(s2, 20, "零", StrPadLeft))

	s3 := "我的长度很长不需要填充"
	Equal(t, "我的长度很长不需要填充", StrPad(s3, 5, "我的长度很长不需要填充", StrPadRight))
}
