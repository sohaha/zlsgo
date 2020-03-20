package zstring

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestUrl(t *testing.T) {
	tt := zlsgo.NewTest(t)
	str := "?d=1 &c=2"
	res := UrlEncode(str)
	t.Log(res)
	d, err := UrlDecode(str)
	tt.EqualNil(err)
	tt.Equal(str, d)

	res = UrlRawEncode(str)
	t.Log(res)
	d, err = UrlRawDecode(str)
	tt.EqualNil(err)
	tt.Equal(str, d)
}
