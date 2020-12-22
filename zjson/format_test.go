package zjson

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zstring"
)

var demo = `{
	"i":100,"f":1.6,"time":"2019-09-10 13:48:22","index.key":"66.6",
"quality":"highLevel","user":{"name":"暴龙兽"},"children":["阿古兽","暴龙兽","机器暴龙兽",{}],"other":["\"",666,"1.8","$1",{"rank":["t",1,2,3]}],"bool":false,"boolTrue":true,"none":"","friends":[{"name":"天使兽","quality":"highLevel","age":1},{"age":5,"name":"天女兽","quality":"super"}]}`

func TestDiscard(T *testing.T) {
	t := zlsgo.NewTest(T)
	t.Log(Discard(`{
// 这是测试
"user":{"name":"暴龙兽"}
}`))
}

func TestFormat(t *testing.T) {
	tt := zlsgo.NewTest(t)
	pretty := Format(zstring.String2Bytes(demo))
	tt.Log(zstring.Bytes2String(pretty))

	str2 := Ugly(pretty)
	tt.Log(zstring.Bytes2String(str2))

	str3 := FormatOptions(str2, &StFormatOptions{Width: 5, Prefix: "", SortKeys: true})
	tt.Log(zstring.Bytes2String(str3))

	str4 := Ugly(str3)
	tt.Log(zstring.Bytes2String(str4))

	str5 := Format([]byte("1668"))
	tt.Log(zstring.Bytes2String(str5))

	str6 := Ugly(str5)
	tt.Log(zstring.Bytes2String(str6))
}
