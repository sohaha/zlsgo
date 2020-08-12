package zstring

import (
	"regexp"
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestExtract(t *testing.T) {
	tt := zlsgo.NewTest(t)
	res, err := RegexExtract(`abc(\d{2}).*(\w)`, "abc123999ok")
	tt.Equal(true, err == nil)
	tt.Equal([]string{"12", "k"}, res[1:])
	res2, err := RegexExtractAll(`a(\d{2})`, "a123999oa23kdsfsda323")
	tt.EqualExit(true, err == nil)
	t.Log(res2)
	tt.Equal(res2[0][1], "12")
	tt.Equal(res2[1][1], "23")
	tt.Equal(res2[2][1], "32")
	res2, err = RegexExtractAll(`a(\d{2})`, "a123999oa23kdsfsda323", 1)
	tt.EqualExit(true, err == nil)
	t.Log(res2)
	tt.Equal(res2[0][1], "12")
}

func TestRegex(T *testing.T) {
	t := zlsgo.NewTest(T)
	t.Equal(true, RegexMatch("是我啊", "这就是我啊!"))
	t.Equal(false, RegexMatch("是你呀", "这就是我啊!"))

	phone := "13800138000"
	isPhone := RegexMatch(`^1[\d]{10}$`, phone)
	t.Equal(true, isPhone)
	phone = "1380013800x"
	isPhone = RegexMatch(`^1[\d]{10}$`, phone)
	t.Equal(false, isPhone)

	t.Equal(2, len(RegexFind(`\d{2}`, "a1b23c456", -1)))
	t.Equal(0, len(RegexFind(`\d{2}`, "abc", -1)))

	str, _ := RegexReplace(`b\d{2}`, "a1b23c456", "*")
	t.Equal("a1*c456", str)

	str, _ = RegexReplaceFunc(`\w{2}`, "abcd", Ucfirst)
	t.Equal("AbCd", str)

	clearRegexpCompile()

	regexCache = map[string]*regexMapStruct{}
	clearRegexpCompile()
}

func BenchmarkRegex1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RegexMatch("是我啊", "这就是我啊!")
	}
}

func BenchmarkRegex2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r, _ := regexp.Compile("是我啊")
		r.Match(String2Bytes("这就是我啊!"))
	}
}

func BenchmarkRegex3(b *testing.B) {
	r, _ := regexp.Compile("是我啊")
	for i := 0; i < b.N; i++ {
		r.Match(String2Bytes("这就是我啊!"))
	}
}
