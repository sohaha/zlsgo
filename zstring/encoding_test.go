package zstring_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zhttp"
	"github.com/sohaha/zlsgo/zstring"
)

func TestBase64(t *testing.T) {
	tt := zlsgo.NewTest(t)
	str := "hi,是我"
	strbyte := []byte(str)
	s := zstring.Base64Encode(strbyte)
	deByte, err := zstring.Base64Decode(s)
	tt.EqualNil(err)
	tt.Equal(strbyte, deByte)

	s2 := zstring.Base64EncodeString(str)
	tt.Equal(zstring.Bytes2String(s), s2)

	de, err := zstring.Base64DecodeString(s2)
	tt.EqualNil(err)
	tt.Equal(str, de)

	de, _ = zstring.Base64DecodeString(string(s))
	tt.Equal(str, de)

}

type testSt struct {
	Name string
}

func TestSerialize(t *testing.T) {
	tt := zlsgo.NewTest(t)
	test := &testSt{"hi"}

	s, err := zstring.Serialize(test)
	tt.EqualNil(err)

	v, err := zstring.UnSerialize(s, &testSt{})
	tt.EqualNil(err)

	test2, ok := v.(*testSt)
	tt.EqualTrue(ok)
	tt.Equal(test.Name, test2.Name)
}

func TestImg2Base64(t *testing.T) {
	tt := zlsgo.NewTest(t)
	res, err := zhttp.Get("https://seekwe.73zls.com/signed/https:%2F%2Fs3-us-west-2.amazonaws.com%2Fsecure.notion-static.com%2Fa4bcc6b2-32ef-4a7d-ba1c-65a0330f632d%2Flogo.png")
	if err == nil {
		file := "tmp/logo.png"
		err = res.ToFile(file)
		if err == nil {
			s, err := zstring.Img2Base64(file)
			tt.EqualNil(err)
			t.Log(s)
			zfile.Rmdir("tmp")
		}
	}
}
