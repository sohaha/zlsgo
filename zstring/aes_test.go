package zstring_test

import (
	"testing"

	zls "github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zstring"
)

func TestAes(t *testing.T) {
	tt := zls.NewTest(t)

	key := "DIS"
	str := zstring.String2Bytes("me")

	cypted, err := zstring.AesEnCrypt(str, key)
	tt.EqualNil(err)

	origdata, err := zstring.AesDeCrypt(cypted, key)
	tt.EqualNil(err)

	tt.Equal(str, origdata)
	tt.Equal(string(str), string(origdata))

	key = ""
	_, err = zstring.AesEnCrypt(str, key)
	tt.EqualNil(err)

	key = zstring.Pad("k", 16, "1", zstring.PadLeft)
	_, err = zstring.AesEnCrypt(str, key)
	tt.EqualNil(err)

	key = zstring.Pad("k", 17, "1", zstring.PadLeft)
	_, err = zstring.AesEnCrypt(str, key)
	tt.EqualNil(err)

	key = zstring.Pad("k", 25, "1", zstring.PadLeft)
	_, err = zstring.AesEnCrypt(str, key)
	tt.EqualNil(err)

	key = zstring.Pad("k", 38, "1", zstring.PadLeft)
	_, err = zstring.AesEnCrypt(str, key)
	tt.EqualNil(err)

	key = "是我呀"
	_, err = zstring.AesEnCrypt(str, key)
	tt.EqualNil(err)

	key = "是我呀，我是测试的人呢，你想干嘛呀？？？我就是试试看这么长会发生什么情况呢"
	cypted, err = zstring.AesEnCrypt(str, key)
	tt.EqualNil(err)
	str, err = zstring.AesDeCrypt(cypted, key)
	t.Log(string(origdata), string(str))

	_, err = zstring.AesDeCrypt([]byte("123"), "")
	tt.Log(err)
	tt.EqualTrue(err != nil)
}

func TestAesString(t *testing.T) {
	tt := zls.NewTest(t)

	key := "DIS"
	str := "待加密数据"

	crypt, err := zstring.AesEnCryptString(str, key)
	tt.EqualNil(err)
	t.Log(crypt)

	orig, err := zstring.AesDeCryptString(crypt, key)
	tt.EqualNil(err)
	t.Log(orig)

	tt.EqualExit(str, orig)

	key = ""
	str = ""
	crypt, err = zstring.AesEnCryptString(str, key)
	tt.EqualNil(err)
	t.Log(crypt)

	orig, err = zstring.AesDeCryptString(crypt, key)
	tt.EqualNil(err)
	t.Log(orig)

	tt.EqualExit(str, orig)

	t.Log(crypt)

	orig, err = zstring.AesDeCryptString("crypt", key)
	tt.Log(err)
	tt.EqualTrue(err != nil)
}
