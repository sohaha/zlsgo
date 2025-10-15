package zjson

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zstring"
)

var demo = `{
	"i":100,"f":1.6,"ii":-999,"time":"2019-09-10 13:48:22","index.key":"66.6",
"quality":"highLevel","user":{"name":"暴龙兽"},"children":["阿古兽","暴龙兽","机器暴龙兽",{}],"other":["\"",666,"1.8","$1",{"rank":["t",1,2,3]}],"bool":false,"boolTrue":true,"none":"","friends":[{"name":"天使兽","quality":"highLevel","age":1},{"age":5,"name":"天女兽",
  "quality":"super"}]}`

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

func TestModifierPrettyEdgeCases(t *testing.T) {
	tst := zlsgo.NewTest(t)
	SetModifiersState(true)
	defer SetModifiersState(false)

	tests := []struct {
		name  string
		json  string
		path  string
		check func(tst *zlsgo.TestUtil, result string)
	}{
		{
			name: "pretty with default indent",
			json: `{"a":1,"b":2}`,
			path: "@format",
			check: func(tst *zlsgo.TestUtil, result string) {
				// Check for formatting: should contain whitespace/newlines
				hasFormatting := false
				for i := range result {
					if result[i] == '\n' || (i > 0 && result[i] == ' ' && result[i-1] == ':') {
						hasFormatting = true
						break
					}
				}
				tst.EqualTrue(hasFormatting)
			},
		},
		{
			name: "ugly modifier",
			json: `{
  "a": 1,
  "b": 2
}`,
			path: "@ugly",
			check: func(tst *zlsgo.TestUtil, result string) {
				// Check for compact formatting: should not contain newlines
				hasNewline := false
				for i := range result {
					if result[i] == '\n' {
						hasNewline = true
						break
					}
				}
				tst.EqualTrue(!hasNewline)
			},
		},
		{
			name: "reverse modifier on array",
			json: `[1,2,3]`,
			path: "@reverse",
			check: func(tst *zlsgo.TestUtil, result string) {
				tst.Equal("[3,2,1]", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(tt.json, tt.path)
			tt.check(tst, result.String())
		})
	}
}

func TestAddModifier(t *testing.T) {
	tt := zlsgo.NewTest(t)
	SetModifiersState(true)
	defer SetModifiersState(false)

	// Add a custom modifier
	AddModifier("test_upper", func(json, arg string) string {
		res := Parse(json)
		if res.Exists() && res.typ == String {
			return `"` + toUpper(res.String()) + `"`
		}
		return json
	})

	tt.EqualTrue(ModifierExists("test_upper"))

	json := `{"name":"hello"}`
	result := Get(json, "name|@test_upper")
	tt.Equal("HELLO", result.String())

	// Test non-existent modifier
	tt.EqualTrue(!ModifierExists("nonexistent"))
}

func toUpper(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'a' && b[i] <= 'z' {
			b[i] -= 32
		}
	}
	return string(b)
}
